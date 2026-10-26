//go:build e2e

// Package e2e runs the BDD specs in features/ against real pcm and pcm-server binaries.
package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/cucumber/godog"
)

// binDir holds pcm and pcm-server, built one time in TestMain.
var binDir string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "pcm-e2e-bin-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "create bin dir:", err)
		os.Exit(1)
	}
	build := exec.Command("go", "build", "-o", dir+string(os.PathSeparator), "../cmd/...")
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "build binaries:", err)
		os.Exit(1)
	}
	binDir = dir

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("feature tests failed")
	}
}
