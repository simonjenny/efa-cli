package prompts

import (
	"strings"
)

// SelectOption is an option with a key (returned as the value) and a label
// (displayed), mirroring PHP associative option arrays. For list options
// Key equals Label.
type SelectOption struct {
	Key   string
	Label string
}

// Select runs an interactive selection prompt and returns the selected
// key (or label for list options), mirroring Laravel\Prompts\select.
func Select(label string, options []SelectOption, defaultValue any, hint string) (string, error) {
	p := &SelectPrompt{
		Prompt:       newPrompt(),
		label:        label,
		options:      options,
		defaultValue: defaultValue,
		hint:         hint,
	}
	p.required = true
	p.scroll = 5
	p.rendererFn = p.renderer
	p.valueFn = p.value
	p.highlighted = 0
	p.reduceScrollingToFitTerminal()
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

type SelectPrompt struct {
	*Prompt
	label        string
	options      []SelectOption
	defaultValue any
	hint         string
	scroll       int
	highlighted  int
	firstVisible int
}

func (p *SelectPrompt) reduceScrollingToFitTerminal() {
	_, lines := terminalDimensions()
	reserved := 5
	p.scroll = max(1, min(p.scroll, lines-reserved))
}

func (p *SelectPrompt) registerKeys() {
	p.On(func(key string) {
		switch {
		case KeyOneOf([]string{KeyUp, KeyUpArrow, KeyLeft, KeyLeftArrow, KeyShiftTab, KeyCtrlP, KeyCtrlB}, key) || key == "k" || key == "h":
			p.highlightPrevious(len(p.options))
		case KeyOneOf([]string{KeyDown, KeyDownArrow, KeyRight, KeyRightArrow, KeyTab, KeyCtrlN, KeyCtrlF}, key) || key == "j" || key == "l":
			p.highlightNext(len(p.options))
		case KeyOneOf(append([]string{KeyCtrlA}, KeyHome...), key):
			p.highlight(0)
		case KeyOneOf(append([]string{KeyCtrlE}, KeyEnd...), key):
			p.highlight(len(p.options) - 1)
		case key == KeyEnter:
			p.Submit()
		}
	})
}

func (p *SelectPrompt) highlight(index int) {
	if index < 0 {
		index = 0
	}
	p.highlighted = index
	if p.highlighted < p.firstVisible {
		p.firstVisible = p.highlighted
	} else if p.highlighted > p.firstVisible+p.scroll-1 {
		p.firstVisible = p.highlighted - p.scroll + 1
	}
}

func (p *SelectPrompt) highlightPrevious(total int) {
	if total == 0 {
		return
	}
	if p.highlighted == 0 {
		p.highlight(total - 1)
	} else {
		p.highlight(p.highlighted - 1)
	}
}

func (p *SelectPrompt) highlightNext(total int) {
	if total == 0 {
		return
	}
	if p.highlighted == total-1 {
		p.highlight(0)
	} else {
		p.highlight(p.highlighted + 1)
	}
}

// value returns the selected key (SelectPrompt::value).
func (p *SelectPrompt) value() any {
	if len(p.options) == 0 {
		return nil
	}
	return p.options[p.highlighted].Key
}

func (p *SelectPrompt) labelOf() string {
	if len(p.options) == 0 {
		return ""
	}
	return p.options[p.highlighted].Label
}

func (p *SelectPrompt) renderer() string {
	r := newRenderer(p.Prompt)
	cols, _ := terminalDimensions()
	maxWidth := cols - 6
	switch p.State {
	case "submit":
		r.box(dim(truncate(p.label, cols-6)), truncate(p.labelOf(), maxWidth), "", "gray")
	case "cancel":
		r.box(truncate(p.label, cols-6), p.renderOptions("dim"), "", "red").
			error(p.CancelMessage)
	case "error":
		r.box(truncate(p.label, cols-6), p.renderOptions("cyan"), "", "yellow").
			warning(truncate(p.Error, cols-5))
	default:
		r.box(cyan(truncate(p.label, cols-6)), p.renderOptions("cyan"), "", "white")
		if p.hint != "" {
			r.hint(p.hint)
		} else {
			r.newLine(1)
		}
	}
	return r.String()
}

func (p *SelectPrompt) renderOptions(scrollColor string) string {
	cols, _ := terminalDimensions()
	var lines []string
	end := min(p.firstVisible+p.scroll, len(p.options))
	for i := p.firstVisible; i < end; i++ {
		opt := p.options[i]
		label := truncate(opt.Label, cols-12)
		if p.State == "cancel" {
			if i == p.highlighted {
				lines = append(lines, dim("› ● "+strikethrough(label)+"  "))
			} else {
				lines = append(lines, dim("  ○ "+strikethrough(label)+"  "))
			}
		} else if i == p.highlighted {
			lines = append(lines, cyan("›")+" "+cyan("●")+" "+label+"  ")
		} else {
			lines = append(lines, "  "+dim("○")+" "+dim(label)+"  ")
		}
	}
	var labels []string
	for _, o := range p.options {
		labels = append(labels, o.Label)
	}
	width := min(longest(labels, 60, 6), cols-6)
	return strings.Join(scrollbar(lines, p.firstVisible, p.scroll, len(p.options), width, scrollColor), "\n")
}
