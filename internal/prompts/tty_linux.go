//go:build linux

package prompts

import (
	"golang.org/x/sys/unix"
)

// ttyState holds the terminal state saved before entering raw mode.
type ttyState = unix.Termios

const (
	ttyReadIoctl  = unix.TCGETS
	ttyWriteIoctl = unix.TCSETS
)

// makeRawTTY switches the terminal into raw input mode, mirroring
// PHP's "stty -icanon -isig -echo": input processing (canonical mode,
// signals, echo) is disabled while the output processing (OPOST/ONLCR)
// stays untouched so that "\n" keeps translating to "\r\n".
func makeRawTTY(fd int) (*ttyState, error) {
	t, err := unix.IoctlGetTermios(fd, ttyReadIoctl)
	if err != nil {
		return nil, err
	}
	saved := *t
	t.Lflag &^= unix.ICANON | unix.ISIG | unix.ECHO
	if err := unix.IoctlSetTermios(fd, ttyWriteIoctl, t); err != nil {
		return nil, err
	}
	return &saved, nil
}

// restoreTTY restores the terminal to its previous state.
func restoreTTY(fd int, saved *ttyState) error {
	return unix.IoctlSetTermios(fd, ttyWriteIoctl, saved)
}
