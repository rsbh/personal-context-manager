package cli

import (
	"strings"
	"testing"
)

func TestConfigSetGet(t *testing.T) {
	a := &app{home: t.TempDir()}
	if out := mustRun(t, a, "config", "set", "remote", "git@example.com:me/ctx.git"); out != "config/remote set\n" {
		t.Errorf("set out = %q", out)
	}
	if out := mustRun(t, a, "config", "get", "remote"); out != "git@example.com:me/ctx.git\n" {
		t.Errorf("get out = %q", out)
	}

	out := mustRun(t, a, "config", "get")
	for _, want := range []string{"path:    " + a.home + "/.pcm/repo\n", "remote:  git@example.com:me/ctx.git\n", "port:    8080\n", "swagger: false\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("get all out = %q, want %q", out, want)
		}
	}
}

func TestConfigGetRemoteNotSet(t *testing.T) {
	a := &app{home: t.TempDir()}
	if out := mustRun(t, a, "config", "get"); !strings.Contains(out, "remote:  none\n") {
		t.Errorf("out = %q", out)
	}
}

func TestConfigSetRestartHint(t *testing.T) {
	for _, key := range []string{"port", "path", "swagger"} {
		a := &app{home: t.TempDir()}
		value := map[string]string{"port": "9999", "path": "/tmp/ctx", "swagger": "true"}[key]
		out := mustRun(t, a, "config", "set", key, value)
		want := "config/" + key + " set\nhint: restart the server: pcm server stop && pcm server start\n"
		if out != want {
			t.Errorf("set %s out = %q, want %q", key, out, want)
		}
	}
}

func TestConfigErrors(t *testing.T) {
	a := &app{home: t.TempDir()}
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"config", "set", "color", "red"}, `unknown key "color": use path, remote, port or swagger`},
		{[]string{"config", "get", "color"}, `unknown key "color": use path, remote, port or swagger`},
		{[]string{"config", "set", "port", "abc"}, `invalid port "abc": use 1-65535`},
		{[]string{"config", "set", "swagger", "yes"}, `invalid bool "yes": use true or false`},
		{[]string{"config", "set", "remote"}, "missing key or value"},
	}
	for _, tt := range tests {
		_, _, err := runCLI(t, a, tt.args...)
		if err == nil || err.Error() != tt.want {
			t.Errorf("pcm %v: err = %v, want %q", tt.args, err, tt.want)
		}
	}
}
