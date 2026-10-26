// Package cli is the pcm command line client.
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/rsbh/personal-context-manager/internal/client"
	"github.com/rsbh/personal-context-manager/internal/config"
)

// errSilent makes pcm exit 1 without printing an error. The command already printed output.
var errSilent = errors.New("silent")

type app struct {
	dir    string
	cfg    config.Config
	client *client.Client
}

func (a *app) load() error {
	dir, err := config.DefaultDir()
	if err != nil {
		return err
	}
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}
	a.dir, a.cfg, a.client = dir, cfg, client.New(client.LocalURL(cfg.Port))
	return nil
}

// newRootCmd builds pcm. Config loads on run, unless a already has a client (tests).
func newRootCmd(a *app) *cobra.Command {
	root := &cobra.Command{
		Use:           "pcm",
		Short:         "Personal context manager",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			if a.client != nil {
				return nil
			}
			return a.load()
		},
	}
	root.AddCommand(
		newHealthCmd(a),
		newServerCmd(a),
		newGetCmd(a),
		newCreateCmd(a),
		newDeleteCmd(a),
	)
	return root
}

// Execute runs pcm and returns the exit code.
func Execute() int {
	if err := newRootCmd(&app{}).Execute(); err != nil {
		if !errors.Is(err, errSilent) {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		return 1
	}
	return 0
}

// field is one "key: value" output line.
type field struct {
	key, value string
}

// printFields prints aligned "key: value" lines.
func printFields(w io.Writer, fields []field) {
	tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)
	for _, f := range fields {
		fmt.Fprintf(tw, "%s:\t%s\n", f.key, f.value)
	}
	tw.Flush()
}
