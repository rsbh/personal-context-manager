// Package markdown renders merged contexts as one markdown document.
package markdown

import (
	"fmt"
	"io"
	"strings"

	"github.com/rsbh/personal-context-manager/internal/store"
)

// ContentType is the media type of Write output.
const ContentType = "text/markdown; charset=utf-8"

// Write prints one heading per context, the description as a quote, then the content.
func Write(w io.Writer, contexts []store.Context) {
	if len(contexts) == 0 {
		fmt.Fprintln(w, "No contexts found.")
		return
	}
	fmt.Fprintln(w, "# Context")
	for _, c := range contexts {
		scope := "global"
		if c.Scope != store.ScopeGlobal {
			scope = fmt.Sprintf("project %q", c.Project)
		}
		fmt.Fprintf(w, "\n## %s/%s (%s)\n", c.Section, c.Name, scope)
		if d := c.Frontmatter.Description; d != "" {
			fmt.Fprintf(w, "> %s\n", d)
		}
		fmt.Fprintf(w, "\n%s\n", strings.TrimRight(c.Content, "\n"))
	}
}
