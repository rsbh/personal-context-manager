package skill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContent(t *testing.T) {
	c := string(Content())
	for _, want := range []string{"---\nname: pcm\n", "description: ", "allowed-tools: Bash(pcm:*)", "pcm get contexts -o md", "pcm server start",
		"## 5. Dump session", "pcm create context <name> -s <section> -f <file>", "--force", "scratchpad",
		"If there is no branch", "Bash(gh pr view:*)", "Bash(git branch:*)", ", Write",
		"pcm search", "pcm index rebuild"} {
		if !strings.Contains(c, want) {
			t.Errorf("SKILL.md does not contain %q", want)
		}
	}
	if strings.Contains(c, "<<'EOF'") {
		t.Error("SKILL.md contains a heredoc")
	}
}

func TestDir(t *testing.T) {
	if got := Dir("/home/u"); got != filepath.Join("/home/u", ".claude", "skills", "pcm") {
		t.Errorf("Dir = %q", got)
	}
}

func TestInstallAndUninstall(t *testing.T) {
	home := t.TempDir()
	file := filepath.Join(Dir(home), "SKILL.md")

	// Overwrites an old file.
	if err := os.MkdirAll(Dir(home), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Install(home)
	if err != nil {
		t.Fatal(err)
	}
	if got != file {
		t.Errorf("Install path = %q, want %q", got, file)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(Content()) {
		t.Error("installed file differs from embedded content")
	}

	removed, err := Uninstall(home)
	if err != nil || !removed {
		t.Fatalf("Uninstall = %v, %v", removed, err)
	}
	if _, err := os.Stat(Dir(home)); !os.IsNotExist(err) {
		t.Errorf("skill dir still exists: %v", err)
	}

	removed, err = Uninstall(home)
	if err != nil || removed {
		t.Errorf("second Uninstall = %v, %v, want false, nil", removed, err)
	}
}
