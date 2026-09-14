package prompts

// Confirm runs a yes/no confirmation prompt, mirroring
// Laravel\Prompts\confirm.
func Confirm(label string, defaultConfirmed bool) (bool, error) {
	p := &ConfirmPrompt{
		Prompt:    newPrompt(),
		label:     label,
		confirmed: defaultConfirmed,
	}
	p.rendererFn = p.renderer
	p.valueFn = p.value
	p.registerKeys()
	v, err := p.Run()
	if err != nil {
		return false, err
	}
	if v == nil {
		return false, nil
	}
	return v.(bool), nil
}

type ConfirmPrompt struct {
	*Prompt
	label     string
	confirmed bool
}

func (p *ConfirmPrompt) value() any {
	return p.confirmed
}

func (p *ConfirmPrompt) labelOf() string {
	if p.confirmed {
		return "Yes"
	}
	return "No"
}

func (p *ConfirmPrompt) registerKeys() {
	p.On(func(key string) {
		switch {
		case key == "y":
			p.confirmed = true
		case key == "n":
			p.confirmed = false
		case KeyOneOf([]string{KeyTab, KeyUp, KeyUpArrow, KeyDown, KeyDownArrow, KeyLeft, KeyLeftArrow, KeyRight, KeyRightArrow, KeyCtrlP, KeyCtrlF, KeyCtrlN, KeyCtrlB}, key) || key == "h" || key == "j" || key == "k" || key == "l":
			p.confirmed = !p.confirmed
		case key == KeyEnter:
			p.Submit()
		}
	})
}

func (p *ConfirmPrompt) renderer() string {
	r := newRenderer(p.Prompt)
	cols, _ := terminalDimensions()
	switch p.State {
	case "submit":
		r.box(dim(truncate(p.label, cols-6)), truncate(p.labelOf(), cols-6), "", "gray")
	case "cancel":
		r.box(truncate(p.label, cols-6), p.renderOptions(), "", "red").
			error(p.CancelMessage)
	case "error":
		r.box(truncate(p.label, cols-6), p.renderOptions(), "", "yellow").
			warning(truncate(p.Error, cols-5))
	default:
		r.box(cyan(truncate(p.label, cols-6)), p.renderOptions(), "", "gray")
		r.newLine(1)
	}
	return r.String()
}

func (p *ConfirmPrompt) renderOptions() string {
	cols, _ := terminalDimensions()
	length := (cols - 14) / 2
	yes := truncate("Yes", length)
	no := truncate("No", length)
	if p.State == "cancel" {
		if p.confirmed {
			return dim("● " + strikethrough(yes) + " / ○ " + strikethrough(no))
		}
		return dim("○ " + strikethrough(yes) + " / ● " + strikethrough(no))
	}
	if p.confirmed {
		return green("●") + " " + yes + " " + dim("/ ○ "+no)
	}
	return dim("○ "+yes+" /") + " " + green("●") + " " + no
}
