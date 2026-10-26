package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

// openBrowser opens url in the default browser. Tests replace it.
var openBrowser = func(url string) error {
	name := "xdg-open"
	if runtime.GOOS == "darwin" {
		name = "open"
	}
	return exec.Command(name, url).Start()
}

func newUICmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "ui",
		Short: "Open the web UI in the browser",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if fields, ok := healthFields(cmd.Context(), a.client); !ok {
				printFields(cmd.OutOrStdout(), fields)
				return errSilent
			}
			url := a.client.BaseURL()
			fmt.Fprintf(cmd.OutOrStdout(), "ui: %s\n", url)
			// PCM_NO_BROWSER=1: print the URL only (tests, remote shells).
			if os.Getenv("PCM_NO_BROWSER") != "" {
				return nil
			}
			if err := openBrowser(url); err != nil {
				return fmt.Errorf("open browser: %w", err)
			}
			return nil
		},
	}
}
