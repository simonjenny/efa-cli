// Package table replicates the Symfony Console Table rendering used by
// Laravel Prompts 0.3.2 (TableRenderer): rounded box-drawing borders, dim
// headers, one space padding, multi-line cells and byte-based padding with
// a display-width correction.
package table

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// Render produces the grid lines of a Symfony-style table.
func Render(headers []string, rows [][]string) []string {
	numColumns := len(headers)
	for _, row := range rows {
		if len(row) > numColumns {
			numColumns = len(row)
		}
	}

	// Column widths: display width of the widest cell (including headers),
	// plus the two padding characters of the " %s " content format.
	widths := make([]int, numColumns)
	for c := 0; c < numColumns; c++ {
		maxCell := 0
		if c < len(headers) {
			maxCell = cellDisplayWidth(headers[c])
		}
		for _, row := range rows {
			if c < len(row) {
				if w := cellDisplayWidth(row[c]); w > maxCell {
					maxCell = w
				}
			}
		}
		widths[c] = maxCell + 2
	}

	var out []string
	if numColumns == 0 {
		return out
	}
	out = append(out, borderLine(widths, '┌', '┬', '┐'))
	if len(headers) > 0 {
		out = append(out, renderRow(widths, headers, true)...)
		out = append(out, borderLine(widths, '├', '┼', '┤'))
	}
	for _, row := range rows {
		out = append(out, renderRow(widths, row, false)...)
	}
	out = append(out, borderLine(widths, '└', '┴', '┘'))
	return out
}

// cellDisplayWidth is the display width of the widest line of a cell.
func cellDisplayWidth(cell string) int {
	max := 0
	for _, line := range strings.Split(cell, "\n") {
		if w := runewidth.StringWidth(line); w > max {
			max = w
		}
	}
	return max
}

func renderRow(widths []int, cells []string, header bool) []string {
	// Split cells into lines.
	cellLines := make([][]string, len(widths))
	height := 1
	for c := 0; c < len(widths); c++ {
		if c < len(cells) {
			cellLines[c] = strings.Split(cells[c], "\n")
		} else {
			cellLines[c] = []string{""}
		}
		if len(cellLines[c]) > height {
			height = len(cellLines[c])
		}
	}

	var out []string
	for l := 0; l < height; l++ {
		var b strings.Builder
		b.WriteRune('│')
		for c := 0; c < len(widths); c++ {
			cell := ""
			if l < len(cellLines[c]) {
				cell = cellLines[c][l]
			}
			b.WriteString(renderCell(cell, widths[c], header))
			b.WriteRune('│')
		}
		out = append(out, b.String())
	}
	return out
}

// renderCell pads the cell the way Symfony Table does: the padding targets
// the byte length (widths[c] + len(cell) - displayWidth(cell)).
func renderCell(cell string, width int, header bool) string {
	padTarget := width + len(cell) - runewidth.StringWidth(cell)
	content := " " + cell + " "
	if pad := padTarget - len(content); pad > 0 {
		content += strings.Repeat(" ", pad)
	}
	if header {
		return "\x1b[2m" + content + "\x1b[22m"
	}
	return content
}

func borderLine(widths []int, left, mid, right rune) string {
	var b strings.Builder
	b.WriteRune(left)
	for c, w := range widths {
		b.WriteString(strings.Repeat("─", w))
		if c < len(widths)-1 {
			b.WriteRune(mid)
		}
	}
	b.WriteRune(right)
	return b.String()
}
