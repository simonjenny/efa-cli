package prompts

import (
	"strings"

	"github.com/simonjenny/efa-cli/internal/table"
)

// DisplayTable renders a Symfony-style table, mirroring
// Laravel\Prompts\table (TableRenderer): a leading blank line, every grid
// line prefixed with a single space, and a trailing blank line.
func DisplayTable(headers []string, rows [][]string) {
	r := newRenderer(nil)
	lines := table.Render(headers, rows)
	for _, l := range lines {
		r.line(" " + l)
	}
	writeOutput(strings.Repeat("\n", max(2-newLinesWritten, 0)) + r.output + "\n")
}
