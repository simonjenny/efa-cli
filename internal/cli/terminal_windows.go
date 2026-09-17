//go:build windows

package cli

import (
	"os"

	"golang.org/x/term"
)

func terminalCheck() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
