//go:build !windows

package cli

import "golang.org/x/term"

func terminalCheck() bool {
	return term.IsTerminal(1)
}
