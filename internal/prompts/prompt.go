package prompts

import (
	"strings"
)

// Prompt is the base for all interactive prompts, mirroring
// Laravel\Prompts\Prompt.
type Prompt struct {
	State           string // initial, active, submit, cancel, error, searching
	Error           string
	CancelMessage   string
	newLinesWritten int
	prevFrame       string
	validated       bool
	listeners       []func(key string)
	// value holds the transformed value on submit.
	value any
	// required mirrors the "required" prompt option.
	required bool
	// rendererFn renders the current frame.
	rendererFn func() string
	// valueFn returns the prompt value.
	valueFn func() any
}

func newPrompt() *Prompt {
	return &Prompt{State: "initial", CancelMessage: "Cancelled."}
}

// Value returns the current prompt value.
func (p *Prompt) Value() any {
	if p.valueFn == nil {
		return nil
	}
	return p.valueFn()
}

// RenderTheme renders the current frame via the prompt's renderer.
func (p *Prompt) RenderTheme() string {
	return p.rendererFn()
}

// On registers a key listener.
func (p *Prompt) On(fn func(key string)) {
	p.listeners = append(p.listeners, fn)
}

func (p *Prompt) emit(key string) {
	for _, fn := range p.listeners {
		fn(key)
	}
}

// HandleKeyPress processes a key press and reports whether the prompt
// should keep running.
func (p *Prompt) HandleKeyPress(key string) bool {
	if p.State == "error" {
		p.State = "active"
	}
	p.emit(key)
	if p.State == "submit" {
		return false
	}
	if key == KeyCtrlU {
		p.State = "error"
		p.Error = "This cannot be reverted."
		return true
	}
	if key == KeyCtrlC {
		p.State = "cancel"
		return false
	}
	if p.validated {
		p.validate(p.Value())
	}
	return true
}

// validate mirrors Prompt::validate; the app only uses the "required"
// check, so no custom validators are wired in.
func (p *Prompt) validate(value any) {
	p.validated = true
	if p.required && p.isInvalidWhenRequired(value) {
		p.State = "error"
		p.Error = "Required."
		return
	}
}

func (p *Prompt) isInvalidWhenRequired(value any) bool {
	switch v := value.(type) {
	case string:
		return v == ""
	case bool:
		return !v
	case nil:
		return true
	default:
		return false
	}
}

// Submit validates and sets the submit state.
func (p *Prompt) Submit() {
	p.validate(p.Value())
	if p.State != "error" {
		p.State = "submit"
	}
}

// Render renders and redraws the prompt.
func (p *Prompt) Render() {
	_, lines := terminalDimensions()
	frame := p.RenderTheme()
	if frame == p.prevFrame {
		return
	}
	if p.State == "initial" {
		writeOutput(frame)
		p.State = "active"
		p.prevFrame = frame
		return
	}
	terminalHeight := lines
	previousFrameHeight := len(strings.Split(p.prevFrame, "\n"))
	frameLines := strings.Split(frame, "\n")
	start := 0
	if terminalHeight-previousFrameHeight < 0 {
		start = -(terminalHeight - previousFrameHeight)
	}
	moveCursorToColumn(1)
	if min(terminalHeight, previousFrameHeight)-1 > 0 {
		moveCursorUp(min(terminalHeight, previousFrameHeight) - 1)
	}
	eraseDown()
	writeOutput(strings.Join(frameLines[start:], "\n"))
	p.prevFrame = frame
}

// RunLoop reads keys and dispatches them until the prompt is submitted,
// returning the transformed value.
func (p *Prompt) RunLoop() (any, error) {
	for {
		key, err := Read()
		if err != nil {
			// EOF: treat as cancellation.
			return nil, errEOFInstance
		}
		if key == "" {
			continue
		}
		cont := p.HandleKeyPress(key)
		p.Render()
		if !cont {
			if key == KeyCtrlC {
				Exit(1)
			}
			return p.Value(), nil
		}
	}
}

// Run runs the interactive prompt and returns the value.
func (p *Prompt) Run() (any, error) {
	if !IsInteractive() {
		return p.nonInteractive()
	}
	MakeRaw()
	hideCursor()
	p.Render()
	v, err := p.RunLoop()
	RestoreRaw()
	showCursor()
	return v, err
}

// nonInteractive returns the default value or a validation error when
// stdin is not a terminal (Prompt::default()).
func (p *Prompt) nonInteractive() (any, error) {
	value := p.Value()
	p.validate(value)
	if p.State == "error" {
		return nil, &NonInteractiveValidationError{p.Error}
	}
	return value, nil
}

// errEOF is returned when stdin reaches EOF during a prompt.
type errEOF struct{}

var errEOFInstance error = errEOF{}

func (errEOF) Error() string { return "stdin closed" }

// NonInteractiveValidationError mirrors
// Laravel\Prompts\Exceptions\NonInteractiveValidationException.
type NonInteractiveValidationError struct {
	Message string
}

func (e *NonInteractiveValidationError) Error() string {
	return e.Message
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
