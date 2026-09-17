package prompts

import (
	"strings"

	"github.com/simonjenny/efa-cli/internal/i18n"
)

// Search runs an interactive search prompt and returns the selected key
// (the GID for EFA stop searches), mirroring Laravel\Prompts\search.
func Search(label, placeholder string, options func(term string) []SelectOption, hint string) (string, error) {
	p := &SearchPrompt{
		Prompt:      newPrompt(),
		label:       label,
		placeholder: placeholder,
		optionsFn:   options,
		hint:        hint,
	}
	p.required = true
	p.scroll = 5
	p.rendererFn = p.renderer
	p.valueFn = p.value
	p.highlighted = nil
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

type SearchPrompt struct {
	*Prompt
	label        string
	placeholder  string
	optionsFn    func(term string) []SelectOption
	hint         string
	scroll       int
	highlighted  *int
	firstVisible int
	matchesCache []SelectOption
	matchesSet   bool
	typedValue   string
	cursorPos    int
}

func (p *SearchPrompt) reduceScrollingToFitTerminal() {
	_, lines := terminalDimensions()
	reserved := 7
	p.scroll = max(1, min(p.scroll, lines-reserved))
}

// matches returns the cached matches, fetching them via the options
// callback when needed (SearchPrompt::matches).
func (p *SearchPrompt) matches() []SelectOption {
	if p.matchesSet {
		return p.matchesCache
	}
	p.matchesCache = p.optionsFn(p.typedValue)
	p.matchesSet = true
	return p.matchesCache
}

// search invalidates the matches and resets the highlight.
func (p *SearchPrompt) search() {
	p.State = "searching"
	p.highlighted = nil
	p.Render()
	p.matchesCache = nil
	p.matchesSet = false
	p.firstVisible = 0
	p.State = "active"
}

// visible returns the currently visible matches.
func (p *SearchPrompt) visible() []SelectOption {
	end := min(p.firstVisible+p.scroll, len(p.matches()))
	if p.firstVisible >= end {
		return nil
	}
	return p.matches()[p.firstVisible:end]
}

func (p *SearchPrompt) value() any {
	if !p.matchesSet || p.highlighted == nil {
		return nil
	}
	idx := *p.highlighted
	if idx >= 0 && idx < len(p.matchesCache) {
		return p.matchesCache[idx].Key
	}
	return nil
}

func (p *SearchPrompt) labelOf() string {
	if !p.matchesSet || p.highlighted == nil {
		return ""
	}
	idx := *p.highlighted
	if idx >= 0 && idx < len(p.matchesCache) {
		return p.matchesCache[idx].Label
	}
	return ""
}

func (p *SearchPrompt) highlight(index *int) {
	p.highlighted = index
	if p.highlighted == nil {
		return
	}
	if *p.highlighted < p.firstVisible {
		p.firstVisible = *p.highlighted
	} else if *p.highlighted > p.firstVisible+p.scroll-1 {
		p.firstVisible = *p.highlighted - p.scroll + 1
	}
}

func (p *SearchPrompt) highlightPrevious(total int, allowNull bool) {
	if total == 0 {
		return
	}
	if p.highlighted == nil {
		i := total - 1
		p.highlight(&i)
	} else if *p.highlighted == 0 {
		if allowNull {
			p.highlight(nil)
		} else {
			i := total - 1
			p.highlight(&i)
		}
	} else {
		i := *p.highlighted - 1
		p.highlight(&i)
	}
}

func (p *SearchPrompt) highlightNext(total int, allowNull bool) {
	if total == 0 {
		return
	}
	if p.highlighted != nil && *p.highlighted == total-1 {
		if allowNull {
			p.highlight(nil)
		} else {
			i := 0
			p.highlight(&i)
		}
	} else {
		base := 0
		if p.highlighted != nil {
			base = *p.highlighted + 1
		}
		i := base
		p.highlight(&i)
	}
}

// trackTypedValue handles text input, mirroring TypedValue with
// submit=false and the search-specific ignore rule.
func (p *SearchPrompt) trackTypedValue(key string) {
	ignore := func(k string) bool {
		return KeyOneOf(append(append([]string{}, KeyHome...), append(KeyEnd, KeyCtrlA, KeyCtrlE)...), k) && p.highlighted != nil
	}
	if key != "" && (key[0] == '\x1b' || key == KeyCtrlB || key == KeyCtrlF || key == KeyCtrlA || key == KeyCtrlE) {
		if ignore(key) {
			return
		}
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
		if ignore(string(r)) {
			return
		}
		if r == '\n' {
			// submit=false: nothing.
			continue
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

func (p *SearchPrompt) registerKeys() {
	// TypedValue listener is registered first, like the constructor order
	// in SearchPrompt (trackTypedValue before the key handler).
	p.On(func(key string) {
		p.trackTypedValue(key)
	})
	p.On(func(key string) {
		switch {
		case KeyOneOf([]string{KeyUp, KeyUpArrow, KeyShiftTab, KeyCtrlP}, key):
			p.highlightPrevious(len(p.matches()), true)
		case KeyOneOf([]string{KeyDown, KeyDownArrow, KeyTab, KeyCtrlN}, key):
			p.highlightNext(len(p.matches()), true)
		case KeyOneOf(append([]string{KeyCtrlA}, KeyHome...), key):
			if p.highlighted != nil {
				i := 0
				p.highlight(&i)
			}
		case KeyOneOf(append([]string{KeyCtrlE}, KeyEnd...), key):
			if p.highlighted != nil {
				i := len(p.matches()) - 1
				p.highlight(&i)
			}
		case key == KeyEnter:
			if p.highlighted != nil {
				p.Submit()
			} else {
				p.search()
			}
		case KeyOneOf([]string{KeyLeft, KeyLeftArrow, KeyRight, KeyRightArrow, KeyCtrlB, KeyCtrlF}, key):
			p.highlighted = nil
		default:
			p.search()
		}
	})
}

// valueWithCursor mirrors SearchPrompt::valueWithCursor.
func (p *SearchPrompt) valueWithCursor(maxWidth int) string {
	if p.highlighted != nil {
		if p.typedValue == "" {
			return dim(truncate(p.placeholder, maxWidth))
		}
		return truncate(p.typedValue, maxWidth)
	}
	if p.typedValue == "" {
		return dim(addCursor(p.placeholder, 0, maxWidth))
	}
	return addCursor(p.typedValue, p.cursorPos, maxWidth)
}

// valueWithCursorAndSearchIcon mirrors
// SearchPromptRenderer::valueWithCursorAndSearchIcon.
func (p *SearchPrompt) valueWithCursorAndSearchIcon(maxWidth int) string {
	var labels []string
	for _, m := range p.matches() {
		labels = append(labels, m.Label)
	}
	width := min(longest(labels, 60, 2), maxWidth)
	value := pad(p.valueWithCursor(maxWidth-1)+"  ", width, " ")
	return replaceTrailingSpace(value, cyan("…"))
}

func replaceTrailingSpace(s, char string) string {
	if s == "" {
		return s
	}
	last := s[len(s)-1]
	if last == ' ' || last == '\t' || last == '\n' {
		return s[:len(s)-1] + char
	}
	return s
}

func (p *SearchPrompt) renderer() string {
	r := newRenderer(p.Prompt)
	cols, _ := terminalDimensions()
	maxWidth := cols - 6
	switch p.State {
	case "submit":
		r.box(dim(truncate(p.label, cols-6)), truncate(p.labelOf(), maxWidth), "", "gray")
	case "cancel":
		value := p.typedValue
		if value == "" {
			value = p.placeholder
		}
		r.box(dim(truncate(p.label, cols-6)), strikethrough(dim(truncate(value, maxWidth))), "", "red").
			error(p.CancelMessage)
	case "error":
		r.box(truncate(p.label, cols-6), p.valueWithCursor(maxWidth), p.renderOptions(), "yellow").
			warning(truncate(p.Error, cols-5))
	case "searching":
		r.box(cyan(truncate(p.label, cols-6)), p.valueWithCursorAndSearchIcon(maxWidth), p.renderOptions(), "gray").
			hint(p.hint)
	default:
		r.box(cyan(truncate(p.label, cols-6)), p.valueWithCursor(maxWidth), p.renderOptions(), "gray")
		if p.hint != "" {
			r.hint(p.hint)
		} else {
			r.newLine(1)
		}
		p.spaceForDropdown(r)
	}
	return r.String()
}

// spaceForDropdown reserves space to prevent jumping.
func (p *SearchPrompt) spaceForDropdown(r *Renderer) {
	if p.typedValue != "" {
		return
	}
	_, lines := terminalDimensions()
	count := len(p.matches())
	r.newLine(max(0, min(p.scroll, lines-7)-count))
	if count == 0 {
		r.newLine(1)
	}
}

func (p *SearchPrompt) renderOptions() string {
	cols, _ := terminalDimensions()
	if p.typedValue != "" && len(p.matches()) == 0 {
		msg := i18n.T("prompt.no_results")
		if p.State == "searching" {
			msg = i18n.T("prompt.searching")
		}
		return gray("  " + msg)
	}
	var lines []string
	visible := p.visible()
	for i, opt := range visible {
		label := truncate(opt.Label, cols-10)
		index := p.firstVisible + i
		if p.highlighted != nil && *p.highlighted == index {
			lines = append(lines, cyan("›")+" "+label+"  ")
		} else {
			lines = append(lines, "  "+dim(label)+"  ")
		}
	}
	var labels []string
	for _, m := range p.matches() {
		labels = append(labels, m.Label)
	}
	width := min(longest(labels, 60, 4), cols-6)
	return strings.Join(scrollbar(lines, p.firstVisible, p.scroll, len(p.matches()), width, "cyan"), "\n")
}
