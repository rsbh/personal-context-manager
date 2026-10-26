package cli

import (
	"bytes"
	"errors"
	"testing"
)

func TestArgErrors(t *testing.T) {
	tests := []struct {
		args    []string
		wantMsg string
		wantUse string
	}{
		{[]string{"create", "project"}, "missing project name", "pcm create project <name> [flags]"},
		{[]string{"delete", "section"}, "missing section name", "pcm delete section <name> [flags]"},
		{[]string{"create", "section", "a", "b"}, "too many arguments: want 1 (name), got 2", "pcm create section <name> [flags]"},
		{[]string{"get", "projects", "a", "b"}, "too many arguments: want at most 1 (name), got 2", "pcm get project [name] [flags]"},
		{[]string{"get", "sections", "x"}, `unexpected argument "x"`, "pcm get section [flags]"},
		{[]string{"health", "x"}, `unexpected argument "x"`, "pcm health [flags]"},
	}
	for _, tt := range tests {
		t.Run(tt.wantMsg, func(t *testing.T) {
			// No client call happens: args are checked first.
			_, _, err := runCLI(t, &app{client: nil, cwd: t.TempDir()}, tt.args...)
			var ue *usageError
			if !errors.As(err, &ue) {
				t.Fatalf("err = %v, want *usageError", err)
			}
			if ue.msg != tt.wantMsg || ue.usage != tt.wantUse {
				t.Errorf("usageError = {%q, %q}, want {%q, %q}", ue.msg, ue.usage, tt.wantMsg, tt.wantUse)
			}
		})
	}
}

func TestPrintError(t *testing.T) {
	var buf bytes.Buffer
	printError(&buf, &usageError{msg: "missing section name", usage: "pcm create section <name> [flags]"})
	want := "error: missing section name\nusage: pcm create section <name> [flags]\n"
	if buf.String() != want {
		t.Errorf("out = %q, want %q", buf.String(), want)
	}

	buf.Reset()
	printError(&buf, errors.New("boom"))
	if buf.String() != "error: boom\n" {
		t.Errorf("out = %q", buf.String())
	}
}
