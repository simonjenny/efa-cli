package prompts

import (
	"math"
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
)

// Renderer mirrors Laravel\Prompts\Themes\Default\Renderer.
type Renderer struct {
	output string
	prompt *Prompt
}

func newRenderer(p *Prompt) *Renderer {
	return &Renderer{prompt: p}
}

// line appends a line of output.
func (r *Renderer) line(message string) *Renderer {
	r.output += message + "\n"
	return r
}

// newLine appends blank lines.
func (r *Renderer) newLine(count int) *Renderer {
	r.output += strings.Repeat("\n", count)
	return r
}

// when applies the callback when value is truthy.
func (r *Renderer) when(cond bool, fn func(*Renderer), elseFn func(*Renderer)) *Renderer {
	if cond {
		fn(r)
	} else if elseFn != nil {
		elseFn(r)
	}
	return r
}

// warning renders a warning line.
func (r *Renderer) warning(message string) *Renderer {
	return r.line(yellow("  ⚠ " + message))
}

// error renders an error line.
func (r *Renderer) error(message string) *Renderer {
	return r.line(red("  ⚠ " + message))
}

// hint renders a hint line.
func (r *Renderer) hint(message string) *Renderer {
	if message == "" {
		return r
	}
	cols, _ := terminalDimensions()
	message = truncate(message, cols-6)
	return r.line(gray("  " + message))
}

// String mirrors Renderer::__toString.
func (r *Renderer) String() string {
	prefix := strings.Repeat("\n", max(2-r.prompt.newLinesWritten, 0))
	trailing := ""
	if r.prompt.State == "submit" || r.prompt.State == "cancel" {
		trailing = "\n"
	}
	return prefix + r.output + trailing
}

// ansiEscapePattern strips ANSI escape sequences and Symfony style tags.
var ansiEscapePattern = regexp.MustCompile(`\x1b[^m]*m|<(?:[fb]g|options)=[a-z,;]+>(.*?)</>`)

// stripEscapeSequences removes ANSI escape sequences from text.
func stripEscapeSequences(text string) string {
	return ansiEscapePattern.ReplaceAllString(text, "$1")
}

// displayWidth returns the display width of a string.
func displayWidth(s string) int {
	return runewidth.StringWidth(stripEscapeSequences(s))
}

// truncate truncates a value with an ellipsis if it exceeds the given
// width (character based, like mb_strimwidth).
func truncate(s string, width int) string {
	if width <= 0 {
		return s
	}
	if displayWidth(s) <= width {
		return s
	}
	runes := []rune(s)
	if width-1 >= len(runes) {
		return s
	}
	return string(runes[:width-1]) + "…"
}

// longest returns the length of the longest line (with padding), floored
// at the renderer's min width.
func longest(lines []string, minWidth, padding int) int {
	maxLen := 0
	for _, l := range lines {
		if w := displayWidth(l) + padding; w > maxLen {
			maxLen = w
		}
	}
	if maxLen < minWidth {
		return minWidth
	}
	return maxLen
}

// pad pads text to the given length ignoring ANSI escape sequences.
func pad(text string, length int, char string) string {
	if char == "" {
		char = " "
	}
	right := max(0, length-displayWidth(text))
	return text + strings.Repeat(char, right)
}

// mbTrimWidthBackwards truncates a string from the end (character based).
func mbTrimWidthBackwards(s string, start, width int) string {
	runes := []rune(s)
	reversed := reverseRunes(runes)
	trimmed := trimWidth(reversed, width)
	return string(reverseRunes([]rune(trimmed)))
}

func reverseRunes(r []rune) []rune {
	out := make([]rune, len(r))
	for i, c := range r {
		out[len(r)-1-i] = c
	}
	return out
}

func trimWidth(s []rune, width int) string {
	if width <= 0 {
		return ""
	}
	if width >= len(s) {
		return string(s)
	}
	return string(s[:width])
}

// addCursor renders a value with a virtual cursor, mirroring
// TypedValue::addCursor.
func addCursor(value string, cursorPosition int, maxWidth int) string {
	runes := []rune(value)
	before := string(runes[:cursorPosition])
	var current string
	if cursorPosition < len(runes) {
		current = string(runes[cursorPosition])
	}
	after := ""
	if cursorPosition+1 < len(runes) {
		after = string(runes[cursorPosition+1:])
	}

	cursor := current
	if current == "" || current == "\n" {
		cursor = " "
	}

	spaceBefore := maxWidth - displayWidth(cursor)
	if displayWidth(after) > 0 {
		spaceBefore--
	}
	truncatedBefore, wasTruncatedBefore := before, false
	if displayWidth(before) > spaceBefore {
		truncatedBefore = mbTrimWidthBackwards(before, 0, spaceBefore-1)
		wasTruncatedBefore = true
	}

	spaceAfter := maxWidth
	if wasTruncatedBefore {
		spaceAfter--
	}
	spaceAfter -= displayWidth(truncatedBefore) + displayWidth(cursor)
	truncatedAfter, wasTruncatedAfter := after, false
	if displayWidth(after) > spaceAfter {
		truncatedAfter = trimWidth([]rune(after), spaceAfter-1)
		wasTruncatedAfter = true
	}

	out := ""
	if wasTruncatedBefore {
		out += dim("…")
	}
	out += truncatedBefore
	out += inverse(cursor)
	if current == "\n" {
		out += "\n"
	}
	out += truncatedAfter
	if wasTruncatedAfter {
		out += dim("…")
	}
	return out
}

// scrollbar renders a scrollbar beside the visible items.
func scrollbar(visible []string, firstVisible, height, total, width int, color string) []string {
	if height >= total {
		return visible
	}
	scrollPos := scrollPosition(firstVisible, height, total)
	out := make([]string, len(visible))
	for i, line := range visible {
		padded := pad(line, width, " ")
		if i == scrollPos {
			out[i] = replaceLastChar(padded, cyan("┃"))
		} else {
			out[i] = replaceLastChar(padded, gray("│"))
		}
	}
	return out
}

// replaceLastChar replaces the last byte of a string, matching PHP's
// preg_replace('/.$/', $char, $line) without the /u modifier.
func replaceLastChar(s, char string) string {
	if s == "" {
		return s
	}
	return s[:len(s)-1] + char
}

// scrollPosition returns the position of the scrollbar handle.
func scrollPosition(firstVisible, height, total int) int {
	if firstVisible == 0 {
		return 0
	}
	maxPosition := total - height
	if firstVisible == maxPosition {
		return height - 1
	}
	if height <= 2 {
		return -1
	}
	percent := float64(firstVisible) / float64(maxPosition)
	return int(math.Round(percent*float64(height-3))) + 1
}

// Box draws a box, mirroring DrawsBoxes::box.
func (r *Renderer) box(title, body, footer, color string) *Renderer {
	cols, _ := terminalDimensions()
	minWidth := min(60, cols-6)

	bodyLines := strings.Split(body, "\n")
	footerLines := []string{}
	if footer != "" {
		for _, l := range strings.Split(footer, "\n") {
			if strings.TrimSpace(l) != "" {
				footerLines = append(footerLines, l)
			}
		}
	}
	allLines := append(append([]string{}, bodyLines...), footerLines...)
	allLines = append(allLines, title)
	width := longest(allLines, minWidth, 0)

	titleLength := displayWidth(title)
	titleLabel := ""
	if titleLength > 0 {
		titleLabel = " " + title + " "
	}
	extra := 0
	if titleLength == 0 {
		extra = 2
	}
	topBorder := strings.Repeat("─", width-titleLength+extra)
	r.line(colorWrap(color, " ┌") + titleLabel + colorWrap(color, topBorder+"┐"))

	for _, line := range bodyLines {
		r.line(colorWrap(color, " │") + " " + pad(line, width, " ") + " " + colorWrap(color, "│"))
	}

	if len(footerLines) > 0 {
		r.line(colorWrap(color, " ├"+strings.Repeat("─", width+2)+"┤"))
		for _, line := range footerLines {
			r.line(colorWrap(color, " │") + " " + pad(line, width, " ") + " " + colorWrap(color, "│"))
		}
	}

	r.line(colorWrap(color, " └"+strings.Repeat("─", width+2)+"┘"))
	return r
}

func colorWrap(color, text string) string {
	switch color {
	case "gray":
		return gray(text)
	case "red":
		return red(text)
	case "yellow":
		return yellow(text)
	default:
		return text
	}
}
