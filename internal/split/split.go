// Package split replicates the App\Helpers\Split helper of the original
// efa-cli: doubling newlines and wrapping text at 74 characters using PHP's
// wordwrap semantics.
package split

import "strings"

// Text doubles existing newlines and wraps the text at 74 characters,
// exactly like PHP's str_replace("\n", "\n\n", $text) followed by
// wordwrap($text, 74, "\n").
func Text(text string) string {
	text = strings.ReplaceAll(text, "\n", "\n\n")
	return WordWrap(text, 74, "\n")
}

// WordWrap replicates PHP's wordwrap() for a single-character break string
// without forced cuts, operating on raw bytes like PHP does.
func WordWrap(text string, width int, breakStr string) string {
	if len(text) == 0 {
		return ""
	}
	if len(breakStr) == 1 {
		b := []byte(text)
		laststart, lastspace := 0, 0
		for current := 0; current < len(text); current++ {
			c := text[current]
			switch {
			case c == breakStr[0]:
				laststart = current + 1
				lastspace = laststart
			case c == ' ':
				if current-laststart >= width {
					b[current] = breakStr[0]
					laststart = current + 1
				}
				lastspace = current
			default:
				if current-laststart >= width && laststart != lastspace {
					b[lastspace] = breakStr[0]
					laststart = lastspace + 1
				}
			}
		}
		return string(b)
	}

	// General multi-character break path, mirroring the second branch of
	// PHP's wordwrap implementation.
	var out strings.Builder
	laststart, lastspace := 0, 0
	for current := 0; current < len(text); current++ {
		if strings.HasPrefix(text[current:], breakStr) && current+len(breakStr) < len(text) {
			out.WriteString(text[laststart : current+len(breakStr)])
			current += len(breakStr) - 1
			laststart = current + 1
			lastspace = laststart
		} else if text[current] == ' ' {
			if current-laststart >= width {
				out.WriteString(text[laststart:current])
				out.WriteString(breakStr)
				laststart = current + 1
			}
			lastspace = current
		} else if current-laststart >= width && laststart < lastspace {
			out.WriteString(text[laststart:lastspace])
			out.WriteString(breakStr)
			laststart = lastspace + 1
			lastspace = laststart
		}
	}
	if laststart != len(text) {
		out.WriteString(text[laststart:])
	}
	return out.String()
}
