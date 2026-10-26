//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

// world is the state of one scenario. Each scenario has its own HOME and port.
type world struct {
	home   string
	cwd    string // pcm runs here. Default: home, which is not a git repo.
	port   int
	output string
	code   int
}

func initializeScenario(sc *godog.ScenarioContext) {
	w := &world{}

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		return ctx, w.setup()
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		w.teardown()
		return ctx, err
	})

	sc.Step(`^the server is not running$`, w.serverNotRunning)
	sc.Step(`^the server is running$`, w.serverRunning)
	sc.Step(`^a pid file with a dead pid$`, w.deadPIDFile)
	sc.Step(`^a project "([^"]*)"$`, w.aProject)
	sc.Step(`^a project "([^"]*)" with a section "([^"]*)"$`, w.aProjectWithSection)
	sc.Step(`^a global section "([^"]*)"$`, w.aGlobalSection)
	sc.Step(`^a global section "([^"]*)" with a context "([^"]*)"$`, w.aGlobalSectionWithContext)
	sc.Step(`^I am in a git repo named "([^"]*)"$`, w.inGitRepo)
	sc.Step(`^I am not in a git repo$`, w.notInGitRepo)
	sc.Step(`^I run "([^"]*)"$`, w.iRun)
	sc.Step(`^the exit code is (\d+)$`, w.exitCodeIs)
	sc.Step(`^the output contains "(.*)"$`, w.outputContains)
	sc.Step(`^the output does not contain "(.*)"$`, w.outputNotContains)
	sc.Step(`^the output contains the server pid$`, w.outputContainsPID)
	sc.Step(`^the pid file exists$`, w.pidFileExists)
	sc.Step(`^the pid file does not exist$`, w.pidFileNotExists)
}

func (w *world) pcmDir() string  { return filepath.Join(w.home, ".pcm") }
func (w *world) pidFile() string { return filepath.Join(w.pcmDir(), "server.pid") }

func (w *world) setup() error {
	home, err := os.MkdirTemp("", "pcm-e2e-home-")
	if err != nil {
		return err
	}
	w.home, w.cwd = home, home
	if w.port, err = freePort(); err != nil {
		return err
	}
	if err := os.MkdirAll(w.pcmDir(), 0o755); err != nil {
		return err
	}
	cfg := fmt.Sprintf("port: %d\n", w.port)
	return os.WriteFile(filepath.Join(w.pcmDir(), "config.yaml"), []byte(cfg), 0o644)
}

func (w *world) teardown() {
	_, _ = w.run("server", "stop")
	os.RemoveAll(w.home)
}

// run executes pcm with HOME set to the scenario home.
func (w *world) run(args ...string) (int, error) {
	cmd := exec.Command(filepath.Join(binDir, "pcm"), args...)
	cmd.Env = append(os.Environ(), "HOME="+w.home)
	cmd.Dir = w.cwd
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	w.output = out.String()

	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return 0, nil
	case errors.As(err, &exitErr):
		return exitErr.ExitCode(), nil
	default:
		return -1, err
	}
}

// Given

func (w *world) serverNotRunning() error {
	if code, err := w.run("health"); err != nil || code == 0 {
		return fmt.Errorf("server is running: code %d, err %v", code, err)
	}
	return nil
}

func (w *world) serverRunning() error {
	code, err := w.run("server", "start")
	if err != nil || code != 0 {
		return fmt.Errorf("server start: code %d, err %v, output:\n%s", code, err, w.output)
	}
	return nil
}

func (w *world) deadPIDFile() error {
	// A finished process gives a pid that is not in use.
	p := exec.Command("true")
	if err := p.Run(); err != nil {
		return err
	}
	return os.WriteFile(w.pidFile(), []byte(strconv.Itoa(p.Process.Pid)), 0o644)
}

func (w *world) aProject(name string) error {
	return w.post("/api/v1/projects", fmt.Sprintf(`{"name":%q}`, name))
}

func (w *world) aProjectWithSection(project, section string) error {
	if err := w.aProject(project); err != nil {
		return err
	}
	return w.post("/api/v1/sections", fmt.Sprintf(`{"scope":"project","project":%q,"name":%q}`, project, section))
}

func (w *world) aGlobalSection(name string) error {
	return w.post("/api/v1/sections", fmt.Sprintf(`{"scope":"global","name":%q}`, name))
}

func (w *world) aGlobalSectionWithContext(section, name string) error {
	if err := w.aGlobalSection(section); err != nil {
		return err
	}
	return w.post("/api/v1/contexts", fmt.Sprintf(`{"scope":"global","section":%q,"name":%q,"content":"x"}`, section, name))
}

func (w *world) inGitRepo(name string) error {
	dir := filepath.Join(w.home, "src", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	init := exec.Command("git", "init", "--quiet")
	init.Dir = dir
	if out, err := init.CombinedOutput(); err != nil {
		return fmt.Errorf("git init: %w: %s", err, out)
	}
	w.cwd = dir
	return nil
}

func (w *world) notInGitRepo() error {
	w.cwd = w.home
	return nil
}

// post calls the API directly, to set up state without the CLI.
func (w *world) post(path, body string) error {
	url := fmt.Sprintf("http://localhost:%d%s", w.port, path)
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("POST %s: status %d: %s", path, resp.StatusCode, msg)
	}
	return nil
}

// When

func (w *world) iRun(command string) error {
	args, err := splitArgs(command)
	if err != nil {
		return err
	}
	if len(args) == 0 || args[0] != "pcm" {
		return fmt.Errorf("command must start with pcm: %q", command)
	}
	code, err := w.run(args[1:]...)
	w.code = code
	return err
}

// Then

func (w *world) exitCodeIs(want int) error {
	if w.code != want {
		return fmt.Errorf("exit code = %d, want %d, output:\n%s", w.code, want, w.output)
	}
	return nil
}

func (w *world) outputContains(want string) error {
	// Gherkin keeps \" as is in step text.
	want = strings.ReplaceAll(want, `\"`, `"`)
	if !strings.Contains(w.output, want) {
		return fmt.Errorf("output does not contain %q:\n%s", want, w.output)
	}
	return nil
}

func (w *world) outputNotContains(text string) error {
	text = strings.ReplaceAll(text, `\"`, `"`)
	if strings.Contains(w.output, text) {
		return fmt.Errorf("output contains %q:\n%s", text, w.output)
	}
	return nil
}

func (w *world) outputContainsPID() error {
	data, err := os.ReadFile(w.pidFile())
	if err != nil {
		return fmt.Errorf("read pid file: %w", err)
	}
	return w.outputContains("pid:    " + strings.TrimSpace(string(data)))
}

func (w *world) pidFileExists() error {
	if _, err := os.Stat(w.pidFile()); err != nil {
		return fmt.Errorf("pid file: %w", err)
	}
	return nil
}

func (w *world) pidFileNotExists() error {
	if _, err := os.Stat(w.pidFile()); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("pid file exists or stat failed: %v", err)
	}
	return nil
}

// splitArgs splits a command line like a shell: 'a b' and "a b" are one argument.
func splitArgs(line string) ([]string, error) {
	var (
		args  []string
		cur   strings.Builder
		quote rune
		inArg bool
	)
	for _, r := range line {
		switch {
		case quote != 0 && r == quote:
			quote = 0
		case quote != 0:
			cur.WriteRune(r)
		case r == '\'' || r == '"':
			quote, inArg = r, true
		case r == ' ' || r == '\t':
			if inArg {
				args = append(args, cur.String())
				cur.Reset()
				inArg = false
			}
		default:
			cur.WriteRune(r)
			inArg = true
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("unclosed quote in %q", line)
	}
	if inArg {
		args = append(args, cur.String())
	}
	return args, nil
}

func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
