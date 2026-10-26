package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	fileName    = "config.yaml"
	defaultPort = 8080
)

// Config is read from <dir>/config.yaml. The file is optional.
type Config struct {
	Path   string `yaml:"path,omitempty"`
	Remote string `yaml:"remote,omitempty"`
	Port   int    `yaml:"port,omitempty"`
}

var (
	ErrUnknownKey  = errors.New("unknown key")
	ErrInvalidPort = errors.New("invalid port")
)

// Keys are the config keys, in output order.
var Keys = []string{"path", "remote", "port"}

func unknownKey(key string) error {
	return fmt.Errorf("%w %q: use path, remote or port", ErrUnknownKey, key)
}

// Get returns the value of key as text.
func (c Config) Get(key string) (string, error) {
	switch key {
	case "path":
		return c.Path, nil
	case "remote":
		return c.Remote, nil
	case "port":
		return strconv.Itoa(c.Port), nil
	}
	return "", unknownKey(key)
}

// Set writes one key to <dir>/config.yaml. Other keys stay. Defaults are not written.
func Set(dir, key, value string) error {
	cfg, err := readFile(dir)
	if err != nil {
		return err
	}
	switch key {
	case "path":
		cfg.Path = value
	case "remote":
		cfg.Remote = value
	case "port":
		port, err := strconv.Atoi(value)
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("%w %q: use 1-65535", ErrInvalidPort, value)
		}
		cfg.Port = port
	default:
		return unknownKey(key)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, fileName), data, 0o644)
}

// readFile reads config.yaml without defaults. A missing file gives an empty Config.
func readFile(dir string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(filepath.Join(dir, fileName))
	switch {
	case errors.Is(err, os.ErrNotExist):
		return cfg, nil
	case err != nil:
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
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

	cfg, err := readFile(dir)
	if err != nil {
		return Config{}, err
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
