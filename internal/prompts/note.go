package prompts

import "strings"

// Note mirrors Laravel\Prompts\Note with the default renderer.
type Note struct {
	message string
	kind    string
}

// Error displays an error note (Laravel\Prompts\error).
func Error(message string) {
	DisplayNote(message, "error")
}

// Warning displays a warning note (Laravel\Prompts\warning).
func Warning(message string) {
	DisplayNote(message, "warning")
}

// Info displays an info note (Laravel\Prompts\info).
func Info(message string) {
	DisplayNote(message, "info")
}

// DisplayNote renders a note of the given type.
func DisplayNote(message, kind string) {
	r := newRenderer(nil)
	lines := strings.Split(message, "\n")
	switch kind {
	case "intro", "outro":
		wrapped := make([]string, len(lines))
		longest := 0
		for i, l := range lines {
			wrapped[i] = " " + l + " "
			if len(wrapped[i]) > longest {
				longest = len(wrapped[i])
			}
		}
		for _, l := range wrapped {
			l = l + strings.Repeat(" ", longest-len(l))
			r.line(" " + bgCyan(black(l)))
		}
	case "warning":
		for _, l := range lines {
			r.line(yellow(" " + l))
		}
	case "error":
		for _, l := range lines {
			r.line(red(" " + l))
		}
	case "alert":
		for _, l := range lines {
			r.line(" " + bgRed(white(" "+l+" ")))
		}
	case "info":
		for _, l := range lines {
			r.line(green(" " + l))
		}
	default:
		for _, l := range lines {
			r.line(" " + l)
		}
	}
	writeOutput(r.output)
}
