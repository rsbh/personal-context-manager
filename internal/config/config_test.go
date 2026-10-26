package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Path != filepath.Join(dir, "repo") || cfg.Port != 8080 || cfg.Remote != "" || cfg.Swagger {
		t.Errorf("cfg = %+v", cfg)
	}
}

func TestSetKeepsOtherKeys(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("port: 9000\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Set(dir, "remote", "git@example.com:me/ctx.git"); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Remote != "git@example.com:me/ctx.git" || cfg.Port != 9000 {
		t.Errorf("cfg = %+v", cfg)
	}
	// Defaults are not written to the file.
	data, _ := os.ReadFile(filepath.Join(dir, "config.yaml"))
	if string(data) != "remote: git@example.com:me/ctx.git\nport: 9000\n" {
		t.Errorf("file = %q", data)
	}
}

func TestSetPathAndPort(t *testing.T) {
	dir := t.TempDir()
	if err := Set(dir, "path", "/tmp/ctx"); err != nil {
		t.Fatal(err)
	}
	if err := Set(dir, "port", "9999"); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Path != "/tmp/ctx" || cfg.Port != 9999 {
		t.Errorf("cfg = %+v", cfg)
	}
}

func TestSetErrors(t *testing.T) {
	dir := t.TempDir()
	if err := Set(dir, "color", "red"); !errors.Is(err, ErrUnknownKey) {
		t.Errorf("unknown key err = %v", err)
	}
	for _, v := range []string{"abc", "0", "70000"} {
		if err := Set(dir, "port", v); !errors.Is(err, ErrInvalidPort) {
			t.Errorf("port %q err = %v", v, err)
		}
	}
	for _, v := range []string{"yes", "1", "TRUE", ""} {
		if err := Set(dir, "swagger", v); !errors.Is(err, ErrInvalidBool) {
			t.Errorf("swagger %q err = %v", v, err)
		}
	}
}

func TestSetSwagger(t *testing.T) {
	dir := t.TempDir()
	if err := Set(dir, "swagger", "true"); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Swagger {
		t.Errorf("cfg = %+v", cfg)
	}
	if err := Set(dir, "swagger", "false"); err != nil {
		t.Fatal(err)
	}
	// false is the default, so it is not written.
	data, _ := os.ReadFile(filepath.Join(dir, "config.yaml"))
	if string(data) != "{}\n" {
		t.Errorf("file = %q", data)
	}
}

func TestGet(t *testing.T) {
	cfg := Config{Path: "/p", Remote: "r", Port: 1, Swagger: true}
	for key, want := range map[string]string{"path": "/p", "remote": "r", "port": "1", "swagger": "true"} {
		if got, err := cfg.Get(key); err != nil || got != want {
			t.Errorf("Get(%q) = %q, %v", key, got, err)
		}
	}
	if _, err := cfg.Get("x"); !errors.Is(err, ErrUnknownKey) {
		t.Errorf("Get(x) err = %v", err)
	}
}
