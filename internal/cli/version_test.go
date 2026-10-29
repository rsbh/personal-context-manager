package cli

import (
	"bytes"
	"testing"

	"github.com/rsbh/personal-context-manager/internal/version"
)

func TestVersionFlag(t *testing.T) {
	// No config and no server: --version must not need them.
	t.Setenv("HOME", t.TempDir())
	cmd := newRootCmd(&app{})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("err = %v", err)
	}
	if want := version.String("pcm") + "\n"; out.String() != want {
		t.Errorf("out = %q, want %q", out.String(), want)
	}
}
