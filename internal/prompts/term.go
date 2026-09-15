// Package prompts replicates Laravel Prompts 0.3.2 rendering and
// interactivity (search, select, text, confirm, spinner, notes and tables)
// for the Go port of efa-cli.
package prompts

import (
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/term"
)

// terminalDimensions returns the width and height of the terminal, using
// the COLUMNS/LINES environment variables first (like Symfony's Terminal),
// then the ioctl size, defaulting to 80x25.
func terminalDimensions() (int, int) {
	if cols, ok := parseEnvInt("COLUMNS"); ok {
		if lines, ok2 := parseEnvInt("LINES"); ok2 {
			return cols, lines
		}
		return cols, 25
	}
	if lines, ok := parseEnvInt("LINES"); ok {
		return 80, lines
	}
	cols, lines, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		cols, lines, err = term.GetSize(int(os.Stdin.Fd()))
	}
	if err != nil {
		return 80, 25
	}
	if cols <= 0 {
		cols = 80
	}
	if lines <= 0 {
		lines = 25
	}
	return cols, lines
}

func parseEnvInt(name string) (int, bool) {
	v, ok := os.LookupEnv(name)
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return 0, false
	}
	return n, true
}

// fakeInput feeds scripted key presses for tests (Prompt::fake).
var (
	fakeInput    []string
	fakeInputIdx int
)

// SetFakeInput makes prompts interactive, reading from the given key
// presses instead of stdin (used by tests).
func SetFakeInput(keys []string) {
	fakeInput = keys
	fakeInputIdx = 0
}

// interactiveOverride forces the interactive state (Symfony -n flag).
var interactiveOverride *bool

// SetInteractive forces the interactive state of the prompts.
func SetInteractive(b bool) {
	interactiveOverride = &b
}

// IsInteractive reports whether stdin is a terminal (stream_isatty(STDIN)).
func IsInteractive() bool {
	if fakeInput != nil {
		return true
	}
	if interactiveOverride != nil {
		return *interactiveOverride
	}
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// rawState is the previous terminal state while a prompt is active.
var rawState *ttyState

// MakeRaw switches the terminal into raw input mode
// (stty -icanon -isig -echo).
func MakeRaw() {
	if fakeInput != nil {
		return
	}
	if rawState != nil {
		return
	}
	state, err := makeRawTTY(int(os.Stdin.Fd()))
	if err != nil {
		// Fall back gracefully: continue without raw mode.
		return
	}
	rawState = state
	installSignalHandler()
}

// RestoreRaw restores the terminal to its previous state.
func RestoreRaw() {
	if fakeInput != nil {
		return
	}
	if rawState == nil {
		return
	}
	_ = restoreTTY(int(os.Stdin.Fd()), rawState)
	rawState = nil
}

// Read reads up to 1024 bytes from stdin, like fread(STDIN, 1024).
func Read() (string, error) {
	if fakeInput != nil {
		if fakeInputIdx >= len(fakeInput) {
			return "", errEOFInstance
		}
		key := fakeInput[fakeInputIdx]
		fakeInputIdx++
		return key, nil
	}
	buf := make([]byte, 1024)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return "", err
	}
	return string(buf[:n]), nil
}

// Exit terminates the process with the given code, restoring the terminal.
func Exit(code int) {
	RestoreRaw()
	showCursor()
	os.Exit(code)
}

// signalHandler restores the terminal when the process is interrupted
// (SIGINT/SIGTERM), mirroring PHP's pcntl_signal(SIGINT, fn () => exit())
// which runs the destructor that restores the TTY.
func installSignalHandler() {
	signalOnce.Do(func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-ch
			Exit(1)
		}()
	})
}

var signalOnce sync.Once

// ioctlSize is a fallback for getting the terminal size directly.
func ioctlSize() (int, int, error) {
	type winsize struct {
		rows, cols, xpixel, ypixel uint16
	}
	var ws winsize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return 0, 0, errno
	}
	return int(ws.cols), int(ws.rows), nil
}
