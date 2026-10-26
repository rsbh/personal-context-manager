// Package skill holds the Claude Code skill for pcm and installs it.
package skill

import (
	_ "embed"
	"errors"
	"os"
	"path/filepath"
)

//go:embed SKILL.md
var content []byte

// Content returns the embedded SKILL.md.
func Content() []byte {
	return content
}

// Dir returns the personal skill folder: <home>/.claude/skills/pcm.
func Dir(home string) string {
	return filepath.Join(home, ".claude", "skills", "pcm")
}

// Install writes SKILL.md into Dir(home). It overwrites an old file.
func Install(home string) (string, error) {
	dir := Dir(home)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	file := filepath.Join(dir, "SKILL.md")
	return file, os.WriteFile(file, content, 0o644)
}

// Uninstall removes Dir(home). It returns false if the skill was not installed.
func Uninstall(home string) (bool, error) {
	dir := Dir(home)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return true, os.RemoveAll(dir)
}
