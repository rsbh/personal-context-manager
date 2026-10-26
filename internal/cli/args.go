package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// usageError is a wrong-arguments error. It prints with the usage line.
type usageError struct {
	msg   string
	usage string
}

func (e *usageError) Error() string { return e.msg }

func newUsageError(cmd *cobra.Command, format string, a ...any) error {
	return &usageError{msg: fmt.Sprintf(format, a...), usage: cmd.UseLine()}
}

// nameArg accepts exactly one <name> argument.
func nameArg(resource string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		switch {
		case len(args) == 0:
			return newUsageError(cmd, "missing %s name", resource)
		case len(args) > 1:
			return newUsageError(cmd, "too many arguments: want 1 (name), got %d", len(args))
		}
		return nil
	}
}

// optionalNameArg accepts zero or one [name] argument.
func optionalNameArg(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return newUsageError(cmd, "too many arguments: want at most 1 (name), got %d", len(args))
	}
	return nil
}

// noArgs accepts no arguments.
func noArgs(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return newUsageError(cmd, "unexpected argument %q", args[0])
	}
	return nil
}

// printError prints "error: ..." and, for a usageError, the usage line.
func printError(w io.Writer, err error) {
	fields := []field{{"error", err.Error()}}
	var ue *usageError
	if errors.As(err, &ue) {
		fields = append(fields, field{"usage", ue.usage})
	}
	printFields(w, fields)
}
