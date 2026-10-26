package cli

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/rsbh/personal-context-manager/internal/client"
	"github.com/rsbh/personal-context-manager/internal/store"
)

// contextFlags are scope flags plus -s <section>.
type contextFlags struct {
	scopeFlags
	section string
}

func (f *contextFlags) register(cmd *cobra.Command) {
	f.scopeFlags.register(cmd)
	cmd.Flags().StringVarP(&f.section, "section", "s", "", "section name")
}

func requireSection(cmd *cobra.Command, section string) error {
	if section == "" {
		return newUsageError(cmd, "missing section: use -s <section>")
	}
	return nil
}

func newGetContextCmd(a *app) *cobra.Command {
	var (
		flags  contextFlags
		output string
	)
	cmd := &cobra.Command{
		Use:     "context [name]",
		Aliases: []string{"contexts"},
		Short:   "List contexts, or show one context",
		Args:    optionalNameArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOutput(cmd, output, args); err != nil {
				return err
			}
			if output == outputMD {
				return a.resolveContexts(cmd, flags)
			}
			if len(args) == 1 {
				return a.getContext(cmd, flags, args[0])
			}
			return a.listContexts(cmd, flags)
		},
	}
	flags.register(cmd)
	cmd.Flags().StringVarP(&output, "output", "o", outputTable, "output format: table or md (merged global + project contexts)")
	return cmd
}

// listContexts lists one scope with -g. Else it lists the current contexts:
// global first, then the project (-p or detected), merged like -o md.
func (a *app) listContexts(cmd *cobra.Command, flags contextFlags) error {
	if !flags.global {
		return a.listCurrentContexts(cmd, flags)
	}
	sc, err := a.scope(cmd, flags.scopeFlags)
	if err != nil {
		return err
	}
	scopeName, project := sc.api()
	contexts, err := a.client.ListContexts(cmd.Context(), scopeName, project, flags.section)
	if err != nil {
		return a.handleErr(cmd, err)
	}

	out := cmd.OutOrStdout()
	if len(contexts) == 0 {
		fmt.Fprintln(out, "No contexts found.")
		return nil
	}
	now := time.Now()
	rows := make([][]string, len(contexts))
	for i, c := range contexts {
		rows[i] = []string{c.Name, c.Section, c.Frontmatter.Title, c.Frontmatter.Description, age(c.UpdatedAt, now)}
	}
	printTable(out, []string{"NAME", "SECTION", "TITLE", "DESCRIPTION", "UPDATED"}, rows)
	return nil
}

// getContext prints the md file as stored: frontmatter, then content.
func (a *app) getContext(cmd *cobra.Command, flags contextFlags, name string) error {
	if err := requireSection(cmd, flags.section); err != nil {
		return err
	}
	sc, err := a.scope(cmd, flags.scopeFlags)
	if err != nil {
		return err
	}
	scopeName, project := sc.api()
	c, err := a.client.GetContext(cmd.Context(), client.ContextID(scopeName, project, flags.section, name))
	if err != nil {
		return a.handleErr(cmd, err)
	}
	data, err := store.Render(c.Frontmatter, c.Content)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	out.Write(data)
	if len(data) > 0 && data[len(data)-1] != '\n' {
		fmt.Fprintln(out)
	}
	return nil
}

func newCreateContextCmd(a *app) *cobra.Command {
	var (
		flags   contextFlags
		content string
		file    string
		force   bool
		fm      store.Frontmatter
	)
	cmd := &cobra.Command{
		Use:   "context <name>",
		Short: "Create a context in a section",
		Args:  nameArg("context"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireSection(cmd, flags.section); err != nil {
				return err
			}
			hasMsg, hasFile := cmd.Flags().Changed("message"), cmd.Flags().Changed("file")
			switch {
			case hasMsg && hasFile:
				return newUsageError(cmd, "use -m or -f, not both")
			case !hasMsg && !hasFile:
				return newUsageError(cmd, "missing content: use -m <text> or -f <file>")
			case hasFile:
				data, err := readContent(cmd, file)
				if err != nil {
					return err
				}
				content = string(data)
			}
			sc, err := a.scope(cmd, flags.scopeFlags)
			if err != nil {
				return err
			}
			if sc.detected {
				if err := a.ensureProject(cmd, sc.project); err != nil {
					return err
				}
			}
			scopeName, project := sc.api()
			verb := "created"
			_, err = a.client.CreateContext(cmd.Context(), store.Context{
				Scope:       scopeName,
				Project:     project,
				Section:     flags.section,
				Name:        args[0],
				Frontmatter: fm,
				Content:     content,
			})
			var apiErr *client.APIError
			if force && errors.As(err, &apiErr) && apiErr.Status == http.StatusConflict {
				id := client.ContextID(scopeName, project, flags.section, args[0])
				verb = "updated"
				_, err = a.client.UpdateContext(cmd.Context(), id, fm, content)
			}
			if err != nil {
				return a.handleErr(cmd, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "context/%s %s in %s (section %q)\n", args[0], verb, sc, flags.section)
			return nil
		},
	}
	flags.register(cmd)
	cmd.Flags().StringVarP(&content, "message", "m", "", "context content (markdown)")
	cmd.Flags().StringVarP(&file, "file", "f", "", "read content from file, - for stdin")
	cmd.Flags().BoolVar(&force, "force", false, "replace the context if it exists")
	cmd.Flags().StringVar(&fm.Title, "title", "", "frontmatter title")
	cmd.Flags().StringVar(&fm.Description, "description", "", "frontmatter description")
	cmd.Flags().StringArrayVar(&fm.Tags, "tag", nil, "frontmatter tag (repeat for more)")
	return cmd
}

// readContent reads path, or stdin when path is "-".
func readContent(cmd *cobra.Command, path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(cmd.InOrStdin())
	}
	return os.ReadFile(path)
}

func newDeleteContextCmd(a *app) *cobra.Command {
	var flags contextFlags
	cmd := &cobra.Command{
		Use:   "context <name>",
		Short: "Delete a context from a section",
		Args:  nameArg("context"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireSection(cmd, flags.section); err != nil {
				return err
			}
			sc, err := a.scope(cmd, flags.scopeFlags)
			if err != nil {
				return err
			}
			scopeName, project := sc.api()
			id := client.ContextID(scopeName, project, flags.section, args[0])
			if err := a.client.DeleteContext(cmd.Context(), id); err != nil {
				return a.handleErr(cmd, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "context %q deleted from %s (section %q)\n", args[0], sc, flags.section)
			return nil
		},
	}
	flags.register(cmd)
	return cmd
}

// age formats the time since t like kubectl AGE: 10s, 5m, 3h, 2d.
func age(t, now time.Time) string {
	d := now.Sub(t)
	switch {
	case d < 0:
		return "0s"
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

func (a *app) listCurrentContexts(cmd *cobra.Command, flags contextFlags) error {
	cwd, err := a.workDir()
	if err != nil {
		return err
	}
	sc, err := resolveScope(flags.scopeFlags, cwd)
	var label string
	switch {
	case errors.Is(err, errNotInRepo):
		sc, label = scope{global: true}, "none (not in a git repo)"
	case err != nil:
		return err
	case sc.detected:
		label = sc.project + " (from git repo)"
	default:
		label = sc.project
	}

	contexts, err := a.client.Resolve(cmd.Context(), sc.project, flags.section)
	if err != nil {
		return a.handleErr(cmd, err)
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "project: %s\n\n", label)
	if len(contexts) == 0 {
		fmt.Fprintln(out, "No contexts found.")
		return nil
	}
	now := time.Now()
	rows := make([][]string, len(contexts))
	for i, c := range contexts {
		rows[i] = []string{c.Scope, c.Name, c.Section, c.Frontmatter.Title, c.Frontmatter.Description, age(c.UpdatedAt, now)}
	}
	printTable(out, []string{"SCOPE", "NAME", "SECTION", "TITLE", "DESCRIPTION", "UPDATED"}, rows)
	return nil
}
