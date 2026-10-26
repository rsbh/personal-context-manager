//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

// world is the state of one scenario. Each scenario has its own HOME and port.
type world struct {
	home   string
	cwd    string // pcm runs here. Default: home, which is not a git repo.
	port   int
	output string
	code   int
	resp   *httpResult // last API call from an "I GET" step
}

type httpResult struct {
	status int
	header http.Header
	body   string
}

func initializeScenario(sc *godog.ScenarioContext) {
	w := &world{}

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		return ctx, w.setup()
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		w.teardown()
		return ctx, err
	})

	sc.Step(`^the server is not running$`, w.serverNotRunning)
	sc.Step(`^the server is running$`, w.serverRunning)
	sc.Step(`^a pid file with a dead pid$`, w.deadPIDFile)
	sc.Step(`^a project "([^"]*)"$`, w.aProject)
	sc.Step(`^a project "([^"]*)" with a section "([^"]*)"$`, w.aProjectWithSection)
	sc.Step(`^a global section "([^"]*)"$`, w.aGlobalSection)
	sc.Step(`^a global section "([^"]*)" with a context "([^"]*)"$`, w.aGlobalSectionWithContext)
	sc.Step(`^a global section "([^"]*)" with a context "([^"]*)" and content "([^"]*)"$`, w.globalContext)
	sc.Step(`^a global section "([^"]*)" with a context "([^"]*)" and content "([^"]*)" and description "([^"]*)"$`, w.globalContextWithDescription)
	sc.Step(`^a project "([^"]*)" section "([^"]*)" with a context "([^"]*)" and content "([^"]*)"$`, w.projectContext)
	sc.Step(`^I am in a git repo named "([^"]*)"$`, w.inGitRepo)
	sc.Step(`^the global section "([^"]*)" has no contexts$`, w.globalSectionEmpty)
	sc.Step(`^I GET "([^"]*)"$`, w.iGet)
	sc.Step(`^I GET "([^"]*)" with header "([^:]+): ([^"]*)"$`, w.iGetWithHeader)
	sc.Step(`^the status is (\d+)$`, w.statusIs)
	sc.Step(`^the header "([^"]*)" is "([^"]*)"$`, w.headerIs)
	sc.Step(`^the body contains "(.*)"$`, w.bodyContains)
	sc.Step(`^the body is "(.*)"$`, w.bodyIs)
	sc.Step(`^the error is "(.*)"$`, w.errorIs)
	sc.Step(`^I am not in a git repo$`, w.notInGitRepo)
	sc.Step(`^I run "([^"]*)"$`, w.iRun)
	sc.Step(`^the exit code is (\d+)$`, w.exitCodeIs)
	sc.Step(`^the output contains "(.*)"$`, w.outputContains)
	sc.Step(`^the output does not contain "(.*)"$`, w.outputNotContains)
	sc.Step(`^the output is:$`, w.outputIs)
	sc.Step(`^the output contains the server pid$`, w.outputContainsPID)
	sc.Step(`^the line with "([^"]*)" comes before the line with "([^"]*)"$`, w.lineBefore)
	sc.Step(`^the line with "([^"]*)" contains "([^"]*)"$`, w.lineContains)
	sc.Step(`^the output contains one line with "([^"]*)"$`, w.oneLineWith)
	sc.Step(`^the pid file exists$`, w.pidFileExists)
	sc.Step(`^the file "([^"]*)" has content "([^"]*)"$`, w.fileHasContent)
	sc.Step(`^the file "([^"]*)" contains "([^"]*)"$`, w.fileContains)
	sc.Step(`^the file "([^"]*)" does not exist$`, w.fileNotExists)
	sc.Step(`^the pid file does not exist$`, w.pidFileNotExists)
}

func (w *world) pcmDir() string  { return filepath.Join(w.home, ".pcm") }
func (w *world) pidFile() string { return filepath.Join(w.pcmDir(), "server.pid") }

func (w *world) setup() error {
	home, err := os.MkdirTemp("", "pcm-e2e-home-")
	if err != nil {
		return err
	}
	w.home, w.cwd = home, home
	if w.port, err = freePort(); err != nil {
		return err
	}
	if err := os.MkdirAll(w.pcmDir(), 0o755); err != nil {
		return err
	}
	cfg := fmt.Sprintf("port: %d\n", w.port)
	return os.WriteFile(filepath.Join(w.pcmDir(), "config.yaml"), []byte(cfg), 0o644)
}

func (w *world) teardown() {
	_, _ = w.run("server", "stop")
	os.RemoveAll(w.home)
}

// run executes pcm with HOME set to the scenario home.
func (w *world) run(args ...string) (int, error) {
	cmd := exec.Command(filepath.Join(binDir, "pcm"), args...)
	cmd.Env = append(os.Environ(), "HOME="+w.home)
	cmd.Dir = w.cwd
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	w.output = out.String()

	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return 0, nil
	case errors.As(err, &exitErr):
		return exitErr.ExitCode(), nil
	default:
		return -1, err
	}
}

// Given

func (w *world) serverNotRunning() error {
	if code, err := w.run("health"); err != nil || code == 0 {
		return fmt.Errorf("server is running: code %d, err %v", code, err)
	}
	return nil
}

func (w *world) serverRunning() error {
	code, err := w.run("server", "start")
	if err != nil || code != 0 {
		return fmt.Errorf("server start: code %d, err %v, output:\n%s", code, err, w.output)
	}
	return nil
}

func (w *world) deadPIDFile() error {
	// A finished process gives a pid that is not in use.
	p := exec.Command("true")
	if err := p.Run(); err != nil {
		return err
	}
	return os.WriteFile(w.pidFile(), []byte(strconv.Itoa(p.Process.Pid)), 0o644)
}

func (w *world) aProject(name string) error {
	return w.post("/api/v1/projects", fmt.Sprintf(`{"name":%q}`, name))
}

func (w *world) aProjectWithSection(project, section string) error {
	if err := w.aProject(project); err != nil {
		return err
	}
	return w.post("/api/v1/sections", fmt.Sprintf(`{"scope":"project","project":%q,"name":%q}`, project, section))
}

func (w *world) aGlobalSection(name string) error {
	return w.post("/api/v1/sections", fmt.Sprintf(`{"scope":"global","name":%q}`, name))
}

func (w *world) aGlobalSectionWithContext(section, name string) error {
	if err := w.aGlobalSection(section); err != nil {
		return err
	}
	return w.post("/api/v1/contexts", fmt.Sprintf(`{"scope":"global","section":%q,"name":%q,"content":"x"}`, section, name))
}

func (w *world) globalContext(section, name, content string) error {
	return w.globalContextWithDescription(section, name, content, "")
}

func (w *world) globalContextWithDescription(section, name, content, description string) error {
	if err := w.ensure("/api/v1/sections", fmt.Sprintf(`{"scope":"global","name":%q}`, section)); err != nil {
		return err
	}
	return w.post("/api/v1/contexts", fmt.Sprintf(
		`{"scope":"global","section":%q,"name":%q,"content":%q,"frontmatter":{"description":%q}}`,
		section, name, content, description))
}

func (w *world) projectContext(project, section, name, content string) error {
	if err := w.ensure("/api/v1/projects", fmt.Sprintf(`{"name":%q}`, project)); err != nil {
		return err
	}
	if err := w.ensure("/api/v1/sections", fmt.Sprintf(`{"scope":"project","project":%q,"name":%q}`, project, section)); err != nil {
		return err
	}
	return w.post("/api/v1/contexts", fmt.Sprintf(
		`{"scope":"project","project":%q,"section":%q,"name":%q,"content":%q}`, project, section, name, content))
}

func (w *world) inGitRepo(name string) error {
	dir := filepath.Join(w.home, "src", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	init := exec.Command("git", "init", "--quiet")
	init.Dir = dir
	if out, err := init.CombinedOutput(); err != nil {
		return fmt.Errorf("git init: %w: %s", err, out)
	}
	w.cwd = dir
	return nil
}

func (w *world) notInGitRepo() error {
	w.cwd = w.home
	return nil
}

// post calls the API directly, to set up state without the CLI.
func (w *world) post(path, body string) error {
	return w.postExpect(path, body, http.StatusCreated)
}

// ensure is post, but "already exists" (409) is also ok.
func (w *world) ensure(path, body string) error {
	return w.postExpect(path, body, http.StatusCreated, http.StatusConflict)
}

func (w *world) postExpect(path, body string, ok ...int) error {
	url := fmt.Sprintf("http://localhost:%d%s", w.port, path)
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if !slices.Contains(ok, resp.StatusCode) {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("POST %s: status %d: %s", path, resp.StatusCode, msg)
	}
	return nil
}

// When

func (w *world) iRun(command string) error {
	args, err := splitArgs(command)
	if err != nil {
		return err
	}
	if len(args) == 0 || args[0] != "pcm" {
		return fmt.Errorf("command must start with pcm: %q", command)
	}
	code, err := w.run(args[1:]...)
	w.code = code
	return err
}

// Then

func (w *world) exitCodeIs(want int) error {
	if w.code != want {
		return fmt.Errorf("exit code = %d, want %d, output:\n%s", w.code, want, w.output)
	}
	return nil
}

func (w *world) outputContains(want string) error {
	// Gherkin keeps \" as is in step text.
	want = strings.ReplaceAll(want, `\"`, `"`)
	if !strings.Contains(w.output, want) {
		return fmt.Errorf("output does not contain %q:\n%s", want, w.output)
	}
	return nil
}

func (w *world) outputIs(doc *godog.DocString) error {
	if got := strings.TrimRight(w.output, "\n"); got != doc.Content {
		return fmt.Errorf("output:\n%s\nwant:\n%s", got, doc.Content)
	}
	return nil
}

func (w *world) outputNotContains(text string) error {
	text = strings.ReplaceAll(text, `\"`, `"`)
	if strings.Contains(w.output, text) {
		return fmt.Errorf("output contains %q:\n%s", text, w.output)
	}
	return nil
}

// linesWith returns the index of each output line that has word as a field.
func (w *world) linesWith(word string) []int {
	var out []int
	for i, line := range strings.Split(w.output, "\n") {
		if slices.Contains(strings.Fields(line), word) {
			out = append(out, i)
		}
	}
	return out
}

func (w *world) lineWith(word string) (int, string, error) {
	idx := w.linesWith(word)
	if len(idx) == 0 {
		return 0, "", fmt.Errorf("no line with %q:\n%s", word, w.output)
	}
	return idx[0], strings.Split(w.output, "\n")[idx[0]], nil
}

func (w *world) lineBefore(first, second string) error {
	i, _, err := w.lineWith(first)
	if err != nil {
		return err
	}
	j, _, err := w.lineWith(second)
	if err != nil {
		return err
	}
	if i >= j {
		return fmt.Errorf("line with %q is not before line with %q:\n%s", first, second, w.output)
	}
	return nil
}

func (w *world) lineContains(word, text string) error {
	_, line, err := w.lineWith(word)
	if err != nil {
		return err
	}
	if !strings.Contains(line, text) {
		return fmt.Errorf("line %q does not contain %q", line, text)
	}
	return nil
}

func (w *world) oneLineWith(word string) error {
	if n := len(w.linesWith(word)); n != 1 {
		return fmt.Errorf("%d lines with %q, want 1:\n%s", n, word, w.output)
	}
	return nil
}

func (w *world) outputContainsPID() error {
	data, err := os.ReadFile(w.pidFile())
	if err != nil {
		return fmt.Errorf("read pid file: %w", err)
	}
	return w.outputContains("pid:    " + strings.TrimSpace(string(data)))
}

func (w *world) pidFileExists() error {
	if _, err := os.Stat(w.pidFile()); err != nil {
		return fmt.Errorf("pid file: %w", err)
	}
	return nil
}

func (w *world) pidFileNotExists() error {
	if _, err := os.Stat(w.pidFile()); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("pid file exists or stat failed: %v", err)
	}
	return nil
}

// splitArgs splits a command line like a shell: 'a b' and "a b" are one argument.
func splitArgs(line string) ([]string, error) {
	var (
		args  []string
		cur   strings.Builder
		quote rune
		inArg bool
	)
	for _, r := range line {
		switch {
		case quote != 0 && r == quote:
			quote = 0
		case quote != 0:
			cur.WriteRune(r)
		case r == '\'' || r == '"':
			quote, inArg = r, true
		case r == ' ' || r == '\t':
			if inArg {
				args = append(args, cur.String())
				cur.Reset()
				inArg = false
			}
		default:
			cur.WriteRune(r)
			inArg = true
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("unclosed quote in %q", line)
	}
	if inArg {
		args = append(args, cur.String())
	}
	return args, nil
}

// path expands a leading ~ to the scenario home.
func (w *world) path(p string) string {
	if rest, ok := strings.CutPrefix(p, "~/"); ok {
		return filepath.Join(w.home, rest)
	}
	return p
}

func (w *world) fileHasContent(p, content string) error {
	p = w.path(p)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(content), 0o644)
}

func (w *world) fileContains(p, text string) error {
	data, err := os.ReadFile(w.path(p))
	if err != nil {
		return err
	}
	if !strings.Contains(string(data), text) {
		return fmt.Errorf("file %s does not contain %q:\n%s", p, text, data)
	}
	return nil
}

func (w *world) fileNotExists(p string) error {
	if _, err := os.Stat(w.path(p)); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("file %s exists or stat failed: %v", p, err)
	}
	return nil
}

func (w *world) globalSectionEmpty(section string) error {
	if err := w.iGet("/api/v1/contexts?scope=global&section=" + section); err != nil {
		return err
	}
	var contexts []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(w.resp.body), &contexts); err != nil {
		return err
	}
	for _, c := range contexts {
		req, err := http.NewRequest(http.MethodDelete, w.url("/api/v1/contexts/"+c.ID), nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			return fmt.Errorf("delete %s: status %d", c.ID, resp.StatusCode)
		}
	}
	return nil
}

func (w *world) url(path string) string {
	return fmt.Sprintf("http://localhost:%d%s", w.port, path)
}

func (w *world) iGet(path string) error {
	return w.iGetWithHeader(path, "", "")
}

func (w *world) iGetWithHeader(path, key, value string) error {
	req, err := http.NewRequest(http.MethodGet, w.url(path), nil)
	if err != nil {
		return err
	}
	if key != "" {
		req.Header.Set(key, strings.TrimSpace(value))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	w.resp = &httpResult{status: resp.StatusCode, header: resp.Header, body: string(body)}
	return nil
}

func (w *world) statusIs(want int) error {
	if w.resp.status != want {
		return fmt.Errorf("status = %d, want %d, body:\n%s", w.resp.status, want, w.resp.body)
	}
	return nil
}

func (w *world) headerIs(key, want string) error {
	if got := w.resp.header.Get(key); got != want {
		return fmt.Errorf("header %s = %q, want %q", key, got, want)
	}
	return nil
}

// unescape converts \" and \n in step text, which Gherkin keeps as is.
func unescape(s string) string {
	return strings.NewReplacer(`\"`, `"`, `\n`, "\n").Replace(s)
}

func (w *world) bodyContains(text string) error {
	if text = unescape(text); !strings.Contains(w.resp.body, text) {
		return fmt.Errorf("body does not contain %q:\n%s", text, w.resp.body)
	}
	return nil
}

func (w *world) bodyIs(want string) error {
	if want = unescape(want); w.resp.body != want {
		return fmt.Errorf("body = %q, want %q", w.resp.body, want)
	}
	return nil
}

// errorIs reads the "error" field of a JSON error body.
func (w *world) errorIs(want string) error {
	var e struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(w.resp.body), &e); err != nil {
		return fmt.Errorf("body is not a JSON error: %w: %s", err, w.resp.body)
	}
	if want = unescape(want); e.Error != want {
		return fmt.Errorf("error = %q, want %q", e.Error, want)
	}
	return nil
}

func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
