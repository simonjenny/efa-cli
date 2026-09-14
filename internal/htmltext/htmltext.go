// Package htmltext replicates the HTML-to-text transformation of
// Stevebauman\Hypertext\Transformer with keepNewLines() enabled, as used
// by the original efa-cli messages command.
package htmltext

import (
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// unicodeSpaces are the characters replaced with a regular space before
// purification (the $spaces property of the Transformer).
var unicodeSpaces = []string{
	"\u00AD", // Soft Hyphen
	"\u200B", // Zero Width Space
	"\u200C", // Zero Width Non-Joiner
	"\u200D", // Zero Width Joiner
	"\u200E", // Left-To-Right Mark
	"\u200F", // Right-To-Left Mark
	"\uFEFF", // Zero Width No-Break Space (Byte Order Mark)
	"\u2060", // Word Joiner
	"\u2002", // En Space
	"\u2003", // Em Space
	"\u2004", // Three-Per-Em Space
	"\u2005", // Four-Per-Em Space
	"\u2006", // Six-Per-Em Space
	"\u2007", // Figure Space
	"\u2008", // Punctuation Space
	"\u2009", // Thin Space
	"\u200A", // Hair Space
	"\u00A0", // Non-breaking Space
	"\u202F", // Narrow No-Break Space
	"\u205F", // Medium Mathematical Space
	"\u3000", // Ideographic Space
	"\u034F", // Combining Grapheme Joiner (CGJ)
}

var (
	tagBoundary    = regexp.MustCompile(`(>)(<)`)
	horizontalWS   = regexp.MustCompile(`[\t\p{Zs}]+`)
	aroundNewlines = regexp.MustCompile(`[\t\n\x0B\f\r\x85\x{2028}\x{2029}\p{Zs}]*\n[\t\n\x0B\f\r\x85\x{2028}\x{2029}\p{Zs}]*`)
	numericEntity  = regexp.MustCompile(`&#(x[0-9a-fA-F]+|[0-9]+);`)
)

// ToText transforms the given HTML into plain text, keeping new lines,
// exactly like (new Transformer)->keepNewLines()->toText($html).
func ToText(source string) string {
	s := quotedPrintableDecode(source)
	s = tagBoundary.ReplaceAllString(s, "$1 $2")
	s = strings.NewReplacer(toAny(unicodeSpaces)...).Replace(s)
	s = purifyText(s)
	s = horizontalWS.ReplaceAllString(s, " ")
	s = aroundNewlines.ReplaceAllString(s, "\n")
	s = htmlspecialcharsDecode(s)
	return strings.TrimSpace(s)
}

func toAny(in []string) []string {
	out := make([]string, 0, len(in)*2)
	for _, s := range in {
		out = append(out, s, " ")
	}
	return out
}

// quotedPrintableDecode decodes quoted-printable encoded text
// (=XX hex sequences, soft line breaks removed).
func quotedPrintableDecode(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			if i+1 < len(s) && s[i+1] == '\n' {
				i++
				continue
			}
			if i+2 < len(s) && isHex(s[i+1]) && isHex(s[i+2]) {
				b.WriteByte(hexVal(s[i+1])<<4 | hexVal(s[i+2]))
				i += 2
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func isHex(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

func hexVal(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	default:
		return c - 'A' + 10
	}
}

// purifyText extracts the text content of the HTML, decoding entities,
// mirroring the effect of HTMLPurifier with HTML.Allowed='p' followed by
// strip_tags: all tags are removed, their text content is kept.
func purifyText(s string) string {
	doc, err := html.ParseFragment(strings.NewReader(s), &html.Node{
		Type:     html.ElementNode,
		Data:     "div",
		DataAtom: atom.Div,
	})
	if err != nil {
		return s
	}
	var b strings.Builder
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
			return
		}
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	for _, n := range doc {
		walk(n)
	}
	return b.String()
}

// htmlspecialcharsDecode decodes the five special HTML entities and
// numeric character references, like PHP's htmlspecialchars_decode.
func htmlspecialcharsDecode(s string) string {
	s = strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&apos;", "'",
	).Replace(s)
	return numericEntity.ReplaceAllStringFunc(s, func(m string) string {
		body := m[2 : len(m)-1]
		var code uint64
		var err error
		if body[0] == 'x' || body[0] == 'X' {
			code, err = strconv.ParseUint(body[1:], 16, 32)
		} else {
			code, err = strconv.ParseUint(body, 10, 32)
		}
		if err != nil {
			return m
		}
		r := rune(code)
		if r == 0 || r > 0x10FFFF || r >= 0xD800 && r <= 0xDFFF {
			return m
		}
		return string(r)
	})
}
