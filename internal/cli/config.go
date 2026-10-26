package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rsbh/personal-context-manager/internal/config"
)

const restartHint = "restart the server: pcm server stop && pcm server start"

func newConfigCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Read and change ~/.pcm/config.yaml",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "get [key]",
			Short: "Print all keys, or one key: path, remote or port",
			Args:  optionalNameArg,
			RunE: func(cmd *cobra.Command, args []string) error {
				home, err := a.homeDir()
				if err != nil {
					return err
				}
				cfg, err := config.Load(config.DirIn(home))
				if err != nil {
					return err
				}
				out := cmd.OutOrStdout()
				if len(args) == 1 {
					v, err := cfg.Get(args[0])
					if err != nil {
						return err
					}
					fmt.Fprintln(out, v)
					return nil
				}
				fields := make([]field, 0, len(config.Keys))
				for _, k := range config.Keys {
					v, _ := cfg.Get(k)
					if v == "" {
						v = "none"
					}
					fields = append(fields, field{k, v})
				}
				printFields(out, fields)
				return nil
			},
		},
		&cobra.Command{
			Use:   "set <key> <value>",
			Short: "Set a key: path, remote or port",
			Args: func(cmd *cobra.Command, args []string) error {
				if len(args) != 2 {
					return newUsageError(cmd, "missing key or value")
				}
				return nil
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				home, err := a.homeDir()
				if err != nil {
					return err
				}
				key := args[0]
				if err := config.Set(config.DirIn(home), key, args[1]); err != nil {
					return err
				}
				out := cmd.OutOrStdout()
				fmt.Fprintf(out, "config/%s set\n", key)
				// The server reads path and port only when it starts.
				if key == "path" || key == "port" {
					printFields(out, []field{{"hint", restartHint}})
				}
				return nil
			},
		},
	)
	return cmd
}
