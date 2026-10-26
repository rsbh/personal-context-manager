package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/rsbh/personal-context-manager/internal/skill"
)

func TestSkillCommands(t *testing.T) {
	home := t.TempDir()
	a := &app{home: home}
	dir := skill.Dir(home)

	out := mustRun(t, a, "skill", "install")
	if want := "skill/pcm installed in " + filepath.Join(dir, "SKILL.md") + "\n"; out != want {
		t.Errorf("install out = %q, want %q", out, want)
	}

	if out := mustRun(t, a, "skill", "show"); out != string(skill.Content()) {
		t.Errorf("show out differs from embedded content")
	}

	if out := mustRun(t, a, "skill", "uninstall"); out != "skill/pcm uninstalled from "+dir+"\n" {
		t.Errorf("uninstall out = %q", out)
	}
	if out := mustRun(t, a, "skill", "uninstall"); !strings.HasPrefix(out, "skill/pcm not installed") {
		t.Errorf("second uninstall out = %q", out)
	}
}
