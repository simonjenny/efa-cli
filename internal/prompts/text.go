package prompts

// Text runs a text input prompt and returns the entered value (or the
// default), mirroring Laravel\Prompts\text.
func Text(label, placeholder, defaultValue, hint string) (string, error) {
	p := &TextPrompt{
		Prompt:      newPrompt(),
		label:       label,
		placeholder: placeholder,
		typedValue:  defaultValue,
		hint:        hint,
	}
	p.rendererFn = p.renderer
	p.valueFn = p.value
	p.cursorPos = len([]rune(defaultValue))
	p.registerKeys()
	v, err := p.Run()
	if err != nil {
		return "", err
	}
	if v == nil {
		return "", nil
	}
	return v.(string), nil
}

type TextPrompt struct {
	*Prompt
	label       string
	placeholder string
	hint        string
	typedValue  string
	cursorPos   int
}

// value returns the entered text.
func (p *TextPrompt) value() any {
	return p.typedValue
}

// valueWithCursor mirrors TextPrompt::valueWithCursor.
func (p *TextPrompt) valueWithCursor(maxWidth int) string {
	if p.typedValue == "" {
		return dim(addCursor(p.placeholder, 0, maxWidth))
	}
	return addCursor(p.typedValue, p.cursorPos, maxWidth)
}

func (p *TextPrompt) registerKeys() {
	p.On(func(key string) {
		p.trackTypedValue(key)
	})
}

// trackTypedValue mirrors TypedValue with submit=true (ENTER submits).
func (p *TextPrompt) trackTypedValue(key string) {
	if key != "" && (key[0] == '\x1b' || key == KeyCtrlB || key == KeyCtrlF || key == KeyCtrlA || key == KeyCtrlE) {
		runes := []rune(p.typedValue)
		switch {
		case key == KeyLeft || key == KeyLeftArrow || key == KeyCtrlB:
			p.cursorPos = max(0, p.cursorPos-1)
		case key == KeyRight || key == KeyRightArrow || key == KeyCtrlF:
			p.cursorPos = min(len(runes), p.cursorPos+1)
		case KeyOneOf(append([]string{KeyCtrlA}, KeyHome...), key):
			p.cursorPos = 0
		case KeyOneOf(append([]string{KeyCtrlE}, KeyEnd...), key):
			p.cursorPos = len(runes)
		case key == KeyDelete:
			if p.cursorPos < len(runes) {
				p.typedValue = string(append(runes[:p.cursorPos], runes[p.cursorPos+1:]...))
			}
		}
		return
	}
	for _, r := range key {
		if r == '\n' {
			p.Submit()
			return
		}
		if r == '\x7f' || r == '\x08' {
			if p.cursorPos == 0 {
				return
			}
			runes := []rune(p.typedValue)
			if p.cursorPos-1 < len(runes) {
				p.typedValue = string(append(runes[:p.cursorPos-1], runes[p.cursorPos:]...))
			}
			p.cursorPos--
			continue
		}
		if r >= 32 {
			runes := []rune(p.typedValue)
			p.typedValue = string(append(append(runes[:p.cursorPos], r), runes[p.cursorPos:]...))
			p.cursorPos++
		}
	}
}

func (p *TextPrompt) renderer() string {
	r := newRenderer(p.Prompt)
	cols, _ := terminalDimensions()
	maxWidth := cols - 6
	switch p.State {
	case "submit":
		r.box(dim(truncate(p.label, cols-6)), truncate(p.typedValue, maxWidth), "", "gray")
	case "cancel":
		value := p.typedValue
		if value == "" {
			value = p.placeholder
		}
		r.box(truncate(p.label, cols-6), strikethrough(dim(truncate(value, maxWidth))), "", "red").
			error(p.CancelMessage)
	case "error":
		r.box(truncate(p.label, cols-6), p.valueWithCursor(maxWidth), "", "yellow").
			warning(truncate(p.Error, cols-5))
	default:
		r.box(cyan(truncate(p.label, cols-6)), p.valueWithCursor(maxWidth), "", "gray")
		if p.hint != "" {
			r.hint(p.hint)
		} else {
			r.newLine(1)
		}
	}
	return r.String()
}
