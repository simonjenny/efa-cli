package cli

import (
	"fmt"
	"strings"
)

// decorated wraps text with Symfony-style colour tags, applied only when
// the output is a terminal (like Symfony's OutputFormatter).
func decorated(s, style string) string {
	if !stdoutIsTTY {
		return s
	}
	switch style {
	case "white-bold":
		return "\x1b[1m\x1b[37m" + s + "\x1b[39m\x1b[22m"
	case "green-bold":
		return "\x1b[1m\x1b[32m" + s + "\x1b[39m\x1b[22m"
	case "yellow-bold":
		return "\x1b[1m\x1b[33m" + s + "\x1b[39m\x1b[22m"
	case "green":
		return "\x1b[32m" + s + "\x1b[39m"
	case "cyan":
		return "\x1b[36m" + s + "\x1b[39m"
	default:
		return s
	}
}

// stdoutIsTTY reports whether stdout is a terminal.
var stdoutIsTTY = isTerminalOutput()

func isTerminalOutput() bool {
	// Populated in terminal_windows / terminal_unix files.
	return terminalCheck()
}

var _ = fmt.Sprintf
var _ = strings.Repeat
