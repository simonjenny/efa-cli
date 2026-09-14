package prompts

import (
	"fmt"
	"os"
	"strings"
)

// newLinesWritten tracks how many trailing newlines the last output wrote
// (ConsoleOutput::newLinesWritten, initialised to 1).
var newLinesWritten = 1

// quiet suppresses all output (Symfony -q flag).
var quiet bool

// SetQuiet suppresses or enables all output.
func SetQuiet(b bool) {
	quiet = b
}

// Quiet reports whether output is suppressed.
func Quiet() bool {
	return quiet
}

// Write output capturing the number of trailing new lines.
func writeOutput(message string) {
	if quiet {
		return
	}
	trailing := len(message) - len(strings.TrimRight(message, "\n"))
	if strings.TrimSpace(message) == "" {
		newLinesWritten += trailing
	} else {
		newLinesWritten = trailing
	}
	fmt.Fprint(os.Stdout, message)
}

// Write output directly, bypassing newline capture.
func writeDirectly(message string) {
	if quiet {
		return
	}
	fmt.Fprint(os.Stdout, message)
}

// LastNewLines returns the number of trailing new lines written by the
// last output.
func LastNewLines() int {
	return newLinesWritten
}

// CapturePreviousNewLines stores the current newline count on the prompt.
func CapturePreviousNewLines(p *Prompt) {
	p.newLinesWritten = newLinesWritten
}

// Cursor manipulation.

var cursorHidden bool

func hideCursor() {
	writeDirectly("\x1b[?25l")
	cursorHidden = true
}

func showCursor() {
	if cursorHidden {
		writeDirectly("\x1b[?25h")
		cursorHidden = false
	}
}

func moveCursor(x, y int) {
	var seq string
	if x < 0 {
		seq += fmt.Sprintf("\x1b[%dD", -x)
	} else if x > 0 {
		seq += fmt.Sprintf("\x1b[%dC", x)
	}
	if y < 0 {
		seq += fmt.Sprintf("\x1b[%dA", -y)
	} else if y > 0 {
		seq += fmt.Sprintf("\x1b[%dB", y)
	}
	writeDirectly(seq)
}

func moveCursorToColumn(column int) {
	writeDirectly(fmt.Sprintf("\x1b[%dG", column))
}

func moveCursorUp(lines int) {
	writeDirectly(fmt.Sprintf("\x1b[%dA", lines))
}

func eraseDown() {
	writeDirectly("\x1b[J")
}
