//go:build !darwin && !linux

package prompts

import "golang.org/x/term"

// ttyState holds the terminal state saved before entering raw mode.
type ttyState = term.State

// makeRawTTY switches the terminal into raw mode as a fallback for
// platforms without a custom implementation.
func makeRawTTY(fd int) (*ttyState, error) {
	return term.MakeRaw(fd)
}

// restoreTTY restores the terminal to its previous state.
func restoreTTY(fd int, saved *ttyState) error {
	return term.Restore(fd, saved)
}
