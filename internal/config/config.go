package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	fileName    = "config.yaml"
	defaultPort = 8080
)

// Config is read from <dir>/config.yaml. The file is optional.
type Config struct {
	Path   string `yaml:"path"`
	Remote string `yaml:"remote"`
	Port   int    `yaml:"port"`
}

// DefaultDir returns ~/.pcm.
func DefaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	return DirIn(home), nil
}

// DirIn returns <home>/.pcm.
func DirIn(home string) string {
	return filepath.Join(home, ".pcm")
}

// Load reads the config from dir and applies defaults. It creates dir if missing.
func Load(dir string) (Config, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Config{}, fmt.Errorf("create config dir: %w", err)
	}

	var cfg Config
	data, err := os.ReadFile(filepath.Join(dir, fileName))
	switch {
	case errors.Is(err, os.ErrNotExist):
	case err != nil:
		return Config{}, fmt.Errorf("read config: %w", err)
	default:
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse config: %w", err)
		}
	}

	if cfg.Path == "" {
		cfg.Path = filepath.Join(dir, "repo")
	}
	if cfg.Path, err = expandHome(cfg.Path); err != nil {
		return Config{}, err
	}
	if cfg.Port == 0 {
		cfg.Port = defaultPort
	}
	return cfg, nil
}

func expandHome(p string) (string, error) {
	if p != "~" && !strings.HasPrefix(p, "~/") {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	return filepath.Join(home, strings.TrimPrefix(p, "~")), nil
}
