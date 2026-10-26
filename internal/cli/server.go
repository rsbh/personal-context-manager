package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const (
	serverBin    = "pcm-server"
	pidFileName  = "server.pid"
	logFileName  = "server.log"
	startTimeout = 5 * time.Second
	// stopTimeout is longer than the server shutdown timeout.
	stopTimeout  = 12 * time.Second
	pollInterval = 100 * time.Millisecond
)

func newServerCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Manage the background server",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "start",
			Short: "Start the server in the background",
			Args:  noArgs,
			RunE:  a.serverStart,
		},
		&cobra.Command{
			Use:   "stop",
			Short: "Stop the background server",
			Args:  noArgs,
			RunE:  a.serverStop,
		},
		&cobra.Command{
			Use:   "status",
			Short: "Show server process and health",
			Args:  noArgs,
			RunE:  a.serverStatus,
		},
	)
	return cmd
}

func (a *app) pidFile() string { return filepath.Join(a.dir, pidFileName) }
func (a *app) logFile() string { return filepath.Join(a.dir, logFileName) }

func (a *app) serverStart(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()
	if pid, ok := a.runningPID(); ok {
		printFields(out, []field{{"server", "already running"}, {"pid", strconv.Itoa(pid)}})
		return nil
	}
	if a.client.Health(cmd.Context()) == nil {
		printFields(out, []field{{"server", "already running (not started by pcm)"}, {"url", a.client.BaseURL()}})
		return nil
	}

	bin, err := findServerBin()
	if err != nil {
		return err
	}
	logf, err := os.OpenFile(a.logFile(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open log: %w", err)
	}
	defer logf.Close()

	proc := exec.Command(bin)
	proc.Stdout, proc.Stderr = logf, logf
	// New session: the server keeps running after the terminal closes.
	proc.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := proc.Start(); err != nil {
		return fmt.Errorf("start %s: %w", bin, err)
	}
	pid := proc.Process.Pid
	if err := os.WriteFile(a.pidFile(), []byte(strconv.Itoa(pid)), 0o600); err != nil {
		return fmt.Errorf("write pid file: %w", err)
	}
	_ = proc.Process.Release()

	if !waitFor(cmd.Context(), startTimeout, func() bool { return a.client.Health(cmd.Context()) == nil }) {
		printFields(out, []field{{"server", "failed to start"}, {"log", a.logFile()}})
		return errSilent
	}
	printFields(out, []field{
		{"server", "started"},
		{"url", a.client.BaseURL()},
		{"pid", strconv.Itoa(pid)},
		{"log", a.logFile()},
	})
	return nil
}

func (a *app) serverStop(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()
	pid, ok := a.runningPID()
	if !ok {
		printFields(out, []field{{"server", "not running"}})
		return nil
	}
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("stop pid %d: %w", pid, err)
	}
	if !waitFor(cmd.Context(), stopTimeout, func() bool { return !alive(pid) }) {
		return fmt.Errorf("pid %d did not stop in %s", pid, stopTimeout)
	}
	_ = os.Remove(a.pidFile())
	printFields(out, []field{{"server", "stopped"}, {"pid", strconv.Itoa(pid)}})
	return nil
}

func (a *app) serverStatus(cmd *cobra.Command, _ []string) error {
	fields, ok := healthFields(cmd.Context(), a.client)
	pid := "-"
	if p, running := a.runningPID(); running {
		pid = strconv.Itoa(p)
	}
	printFields(cmd.OutOrStdout(), append(fields, field{"pid", pid}))
	if !ok {
		return errSilent
	}
	return nil
}

// runningPID reads the pid file. It removes a stale pid file.
func (a *app) runningPID() (int, bool) {
	data, err := os.ReadFile(a.pidFile())
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || !alive(pid) {
		_ = os.Remove(a.pidFile())
		return 0, false
	}
	return pid, true
}

func alive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

// findServerBin looks next to the pcm executable first, then in PATH.
func findServerBin() (string, error) {
	if exe, err := os.Executable(); err == nil {
		if exe, err = filepath.EvalSymlinks(exe); err == nil {
			p := filepath.Join(filepath.Dir(exe), serverBin)
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				return p, nil
			}
		}
	}
	p, err := exec.LookPath(serverBin)
	if err != nil {
		return "", fmt.Errorf("%s not found next to pcm or in PATH", serverBin)
	}
	return p, nil
}

// waitFor polls cond until it is true or timeout.
func waitFor(ctx context.Context, timeout time.Duration, cond func() bool) bool {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	t := time.NewTicker(pollInterval)
	defer t.Stop()
	for {
		if cond() {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-t.C:
		}
	}
}
