package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/rsbh/personal-context-manager/internal/client"
)

func newHealthCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check that the server is reachable",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fields, ok := healthFields(cmd.Context(), a.client)
			printFields(cmd.OutOrStdout(), fields)
			if !ok {
				return errSilent
			}
			return nil
		},
	}
}

// healthFields returns output lines for the server health and true if it is ok.
func healthFields(ctx context.Context, c *client.Client) ([]field, bool) {
	if err := c.Health(ctx); err != nil {
		return []field{
			{"server", "not reachable"},
			{"url", c.BaseURL()},
			{"hint", "run `pcm server start`"},
		}, false
	}
	return []field{
		{"server", "ok"},
		{"url", c.BaseURL()},
	}, true
}
