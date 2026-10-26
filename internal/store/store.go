// Package store keeps contexts as md files in a git repo:
//
//	<root>/global/<section>/<name>.md
//	<root>/projects/<project>/<section>/<name>.md
package store

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	ScopeGlobal  = "global"
	ScopeProject = "project"

	globalDir   = "global"
	projectsDir = "projects"
	ext         = ".md"
	// keepFile lets git track empty project and section folders.
	keepFile = ".gitkeep"
)

var (
	ErrNotFound = errors.New("not found")
	ErrExists   = errors.New("already exists")
	ErrNotEmpty = errors.New("not empty")
	ErrInvalid  = errors.New("invalid")
)

var namePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

// Context is one md file.
type Context struct {
	ID          string      `json:"id"`
	Scope       string      `json:"scope"`
	Project     string      `json:"project,omitempty"`
	Section     string      `json:"section"`
	Name        string      `json:"name"`
	Frontmatter Frontmatter `json:"frontmatter"`
	Content     string      `json:"content,omitempty"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// Filter selects contexts in List. Empty fields match all.
type Filter struct {
	Scope   string
	Project string
	Section string
}

type Store struct {
	root string
	mu   sync.RWMutex
}

// New opens the repo at root. It creates the folder and runs git init if missing.
func New(root string) (*Store, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create repo: %w", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".git")); errors.Is(err, os.ErrNotExist) {
		cmd := exec.Command("git", "init", "--quiet", "-b", "main")
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("git init: %w: %s", err, out)
		}
	}
	for _, d := range []string{globalDir, projectsDir} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			return nil, fmt.Errorf("create %s: %w", d, err)
		}
	}
	return &Store{root: root}, nil
}

func validName(kind, name string) error {
	if !namePattern.MatchString(name) {
		return fmt.Errorf("%w %s name %q: use a-z, 0-9, _ and -", ErrInvalid, kind, name)
	}
	return nil
}

// scopeDir returns the folder for a scope.
func (s *Store) scopeDir(scope, project string) (string, error) {
	switch scope {
	case ScopeGlobal:
		if project != "" {
			return "", fmt.Errorf("%w: project not allowed for global scope", ErrInvalid)
		}
		return filepath.Join(s.root, globalDir), nil
	case ScopeProject:
		if err := validName("project", project); err != nil {
			return "", err
		}
		dir := filepath.Join(s.root, projectsDir, project)
		if !isDir(dir) {
			return "", fmt.Errorf("project %q: %w", project, ErrNotFound)
		}
		return dir, nil
	default:
		return "", fmt.Errorf("%w scope %q: use global or project", ErrInvalid, scope)
	}
}

func (s *Store) sectionDir(scope, project, section string) (string, error) {
	dir, err := s.scopeDir(scope, project)
	if err != nil {
		return "", err
	}
	if err := validName("section", section); err != nil {
		return "", err
	}
	return filepath.Join(dir, section), nil
}

// Projects

func (s *Store) ListProjects() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return listDirs(filepath.Join(s.root, projectsDir))
}

// GetProject returns nil if the project exists.
func (s *Store) GetProject(name string) error {
	if err := validName("project", name); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !isDir(filepath.Join(s.root, projectsDir, name)) {
		return fmt.Errorf("project %q: %w", name, ErrNotFound)
	}
	return nil
}

func (s *Store) CreateProject(name string) error {
	if err := validName("project", name); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return createDir(filepath.Join(s.root, projectsDir, name), "project "+name)
}

func (s *Store) DeleteProject(name string, force bool) error {
	if err := validName("project", name); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return removeDir(filepath.Join(s.root, projectsDir, name), "project "+name, force)
}

// Sections

func (s *Store) ListSections(scope, project string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	dir, err := s.scopeDir(scope, project)
	if err != nil {
		return nil, err
	}
	return listDirs(dir)
}

func (s *Store) CreateSection(scope, project, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	dir, err := s.sectionDir(scope, project, name)
	if err != nil {
		return err
	}
	return createDir(dir, "section "+name)
}

func (s *Store) DeleteSection(scope, project, name string, force bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	dir, err := s.sectionDir(scope, project, name)
	if err != nil {
		return err
	}
	return removeDir(dir, "section "+name, force)
}

// Contexts

// List returns contexts without content.
func (s *Store) List(f Filter) ([]Context, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := []Context{}
	err := filepath.WalkDir(s.root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if d.IsDir() || filepath.Ext(p) != ext {
			return nil
		}
		rel, _ := filepath.Rel(s.root, p)
		c, err := parseID(filepath.ToSlash(strings.TrimSuffix(rel, ext)))
		if err != nil {
			return nil // file outside known layout
		}
		if (f.Scope != "" && f.Scope != c.Scope) ||
			(f.Project != "" && f.Project != c.Project) ||
			(f.Section != "" && f.Section != c.Section) {
			return nil
		}
		if err := s.load(&c, false); err != nil {
			return err
		}
		out = append(out, c)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (s *Store) Get(id string) (Context, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, err := parseID(id)
	if err != nil {
		return Context{}, err
	}
	if err := s.load(&c, true); err != nil {
		return Context{}, err
	}
	return c, nil
}

// Create writes a new context. The section must exist.
func (s *Store) Create(c Context) (Context, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	dir, err := s.sectionDir(c.Scope, c.Project, c.Section)
	if err != nil {
		return Context{}, err
	}
	if !isDir(dir) {
		return Context{}, fmt.Errorf("section %q: %w", c.Section, ErrNotFound)
	}
	if err := validName("context", c.Name); err != nil {
		return Context{}, err
	}
	c.ID = makeID(c.Scope, c.Project, c.Section, c.Name)
	path := s.path(c.ID)
	if _, err := os.Stat(path); err == nil {
		return Context{}, fmt.Errorf("context %q: %w", c.ID, ErrExists)
	}
	if err := writeContext(path, c.Frontmatter, c.Content); err != nil {
		return Context{}, err
	}
	return c, s.load(&c, true)
}

func (s *Store) Update(id string, fm Frontmatter, content string) (Context, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, err := parseID(id)
	if err != nil {
		return Context{}, err
	}
	path := s.path(c.ID)
	if _, err := os.Stat(path); err != nil {
		return Context{}, fmt.Errorf("context %q: %w", id, ErrNotFound)
	}
	if err := writeContext(path, fm, content); err != nil {
		return Context{}, err
	}
	return c, s.load(&c, true)
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, err := parseID(id)
	if err != nil {
		return err
	}
	if err := os.Remove(s.path(c.ID)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("context %q: %w", id, ErrNotFound)
		}
		return err
	}
	return nil
}

func (s *Store) path(id string) string {
	return filepath.Join(s.root, filepath.FromSlash(id)+ext)
}

// load fills frontmatter, updated_at and (if withContent) content from disk.
func (s *Store) load(c *Context, withContent bool) error {
	path := s.path(c.ID)
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("context %q: %w", c.ID, ErrNotFound)
		}
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fm, body, err := parseFile(data)
	if err != nil {
		return fmt.Errorf("context %q: %w", c.ID, err)
	}
	c.Frontmatter = fm
	c.UpdatedAt = info.ModTime().UTC()
	if withContent {
		c.Content = body
	}
	return nil
}

// makeID returns the repo path without .md, e.g. global/style/backend.
func makeID(scope, project, section, name string) string {
	if scope == ScopeGlobal {
		return strings.Join([]string{globalDir, section, name}, "/")
	}
	return strings.Join([]string{projectsDir, project, section, name}, "/")
}

func parseID(id string) (Context, error) {
	parts := strings.Split(id, "/")
	var c Context
	switch {
	case len(parts) == 3 && parts[0] == globalDir:
		c = Context{Scope: ScopeGlobal, Section: parts[1], Name: parts[2]}
	case len(parts) == 4 && parts[0] == projectsDir:
		c = Context{Scope: ScopeProject, Project: parts[1], Section: parts[2], Name: parts[3]}
		if err := validName("project", c.Project); err != nil {
			return Context{}, err
		}
	default:
		return Context{}, fmt.Errorf("%w id %q", ErrInvalid, id)
	}
	if err := validName("section", c.Section); err != nil {
		return Context{}, err
	}
	if err := validName("context", c.Name); err != nil {
		return Context{}, err
	}
	c.ID = id
	return c, nil
}

func writeContext(path string, fm Frontmatter, content string) error {
	data, err := Render(fm, content)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func listDirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := []string{}
	for _, e := range entries {
		if e.IsDir() && namePattern.MatchString(e.Name()) {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

func createDir(dir, label string) error {
	if isDir(dir) {
		return fmt.Errorf("%s: %w", label, ErrExists)
	}
	if err := os.Mkdir(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, keepFile), nil, 0o644)
}

func removeDir(dir, label string, force bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%s: %w", label, ErrNotFound)
		}
		return err
	}
	if !force {
		for _, e := range entries {
			if e.Name() != keepFile {
				return fmt.Errorf("%s: %w", label, ErrNotEmpty)
			}
		}
	}
	return os.RemoveAll(dir)
}
