// Command pcm is the personal context manager CLI.
package main

import (
	"os"

	"github.com/rsbh/personal-context-manager/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
