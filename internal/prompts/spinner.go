package prompts

import (
	"strings"
	"sync"
	"time"
)

// Spin renders a spinner while fn executes, mirroring
// Laravel\Prompts\spin with the default spinner renderer.
func Spin(message string, fn func() error) error {
	_, err := SpinResult(message, func() (any, error) {
		return nil, fn()
	})
	return err
}

// SpinResult renders a spinner while fn executes and returns its result.
func SpinResult[T any](message string, fn func() (T, error)) (T, error) {
	s := &Spinner{
		Prompt:  newPrompt(),
		message: message,
	}
	CapturePreviousNewLines(s.Prompt)
	hideCursor()
	stop := make(chan struct{})
	var done sync.WaitGroup
	done.Add(1)
	go s.spinLoop(stop, &done)
	result, err := fn()
	close(stop)
	done.Wait()
	s.eraseRenderedLines()
	showCursor()
	return result, err
}

// Spinner mirrors Laravel\Prompts\Spinner.
type Spinner struct {
	*Prompt
	message  string
	interval time.Duration
	count    int
	static   bool
}

// Spin renders the spinner and executes the callback.
func (s *Spinner) Spin(fn func() error) error {
	CapturePreviousNewLines(s.Prompt)
	s.static = true
	hideCursor()
	stop := make(chan struct{})
	var done sync.WaitGroup
	done.Add(1)
	go s.spinLoop(stop, &done)
	err := fn()
	close(stop)
	done.Wait()
	s.eraseRenderedLines()
	showCursor()
	return err
}

var spinnerFrames = []string{"⠂", "⠒", "⠐", "⠰", "⠠", "⠤", "⠄", "⠆"}

const spinnerInterval = 75 * time.Millisecond

func (s *Spinner) renderFrame() {
	frame := " " + cyan(spinnerFrames[s.count%len(spinnerFrames)]) + " " + s.message + "\n"
	if s.prevFrame == "" {
		written := strings.Repeat("\n", max(2-newLinesWritten, 0)) + frame
		writeOutput(written)
		s.prevFrame = written
		return
	}
	moveCursorToColumn(1)
	moveCursorUp(1)
	eraseDown()
	writeOutput(frame)
	s.prevFrame = frame
}

// eraseRenderedLines clears the lines rendered by the spinner.
func (s *Spinner) eraseRenderedLines() {
	lines := len(strings.Split(s.prevFrame, "\n"))
	moveCursor(-999, -lines+1)
	eraseDown()
	s.prevFrame = ""
}

// spinLoop animates the spinner until stop is closed.
func (s *Spinner) spinLoop(stop chan struct{}, done *sync.WaitGroup) {
	defer done.Done()
	for {
		select {
		case <-stop:
			return
		default:
			s.renderFrame()
			s.count++
			time.Sleep(spinnerInterval)
		}
	}
}
