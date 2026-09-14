// Package jsonx implements PHP-compatible JSON decoding and encoding.
//
// It mirrors the behaviour of PHP's json_decode/json_encode with
// JSON_PRETTY_PRINT as used by the original efa-cli application:
//   - object key order is preserved (insertion order of the response)
//   - numbers keep their original literal (json.Number) and are re-encoded
//     with PHP float semantics
//   - strings are escaped like PHP does without JSON_UNESCAPED_UNICODE:
//     non-ASCII characters become \uXXXX sequences (surrogate pairs for
//     astral characters), control characters become \b \t \n \f \r or
//     \u00XX, HTML characters are NOT escaped
//   - pretty printing uses a 4-space indent
package jsonx

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// Ordered is a JSON object that preserves key insertion order.
type Ordered struct {
	keys   []string
	values map[string]any
}

// NewOrdered creates an empty ordered object.
func NewOrdered() *Ordered {
	return &Ordered{values: make(map[string]any)}
}

// Set stores a value under the given key, preserving insertion order.
func (o *Ordered) Set(key string, value any) {
	if _, ok := o.values[key]; !ok {
		o.keys = append(o.keys, key)
	}
	o.values[key] = value
}

// Get returns the value stored under the given key.
func (o *Ordered) Get(key string) (any, bool) {
	v, ok := o.values[key]
	return v, ok
}

// Has reports whether the given key exists.
func (o *Ordered) Has(key string) bool {
	_, ok := o.values[key]
	return ok
}

// Delete removes the given key.
func (o *Ordered) Delete(key string) {
	if _, ok := o.values[key]; !ok {
		return
	}
	delete(o.values, key)
	for i, k := range o.keys {
		if k == key {
			o.keys = append(o.keys[:i], o.keys[i+1:]...)
			break
		}
	}
}

// Len returns the number of keys.
func (o *Ordered) Len() int {
	return len(o.keys)
}

// Keys returns the keys in insertion order.
func (o *Ordered) Keys() []string {
	return o.keys
}

// Path navigates nested ordered objects by key, returning nil if any
// step is missing or not an object. Numeric keys index into arrays.
func Path(v any, keys ...string) any {
	for _, k := range keys {
		switch current := v.(type) {
		case *Ordered:
			var ok bool
			v, ok = current.values[k]
			if !ok {
				return nil
			}
		case []any:
			idx, err := strconv.Atoi(k)
			if err != nil || idx < 0 || idx >= len(current) {
				return nil
			}
			v = current[idx]
		default:
			return nil
		}
	}
	return v
}

// Decode parses JSON data into Ordered objects, []any slices, json.Number
// numbers, strings, bools and nil. On any error it returns nil, matching
// PHP's json_decode returning null for invalid input.
func Decode(data []byte) (any, error) {
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.UseNumber()
	v, err := decodeValue(dec)
	if err != nil {
		return nil, err
	}
	// PHP json_decode fails on trailing data.
	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("trailing data after JSON value")
	}
	return v, nil
}

func decodeValue(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	return decodeToken(dec, tok)
}

func decodeToken(dec *json.Decoder, tok json.Token) (any, error) {
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			obj := NewOrdered()
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return nil, err
				}
				key := keyTok.(string)
				val, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				obj.Set(key, val)
			}
			if _, err := dec.Token(); err != nil { // closing '}'
				return nil, err
			}
			return obj, nil
		case '[':
			arr := []any{}
			for dec.More() {
				val, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				arr = append(arr, val)
			}
			if _, err := dec.Token(); err != nil { // closing ']'
				return nil, err
			}
			return arr, nil
		}
		return nil, fmt.Errorf("unexpected delimiter %v", t)
	default:
		return tok, nil
	}
}

// EncodePHP serializes v like PHP json_encode with JSON_PRETTY_PRINT.
func EncodePHP(v any) string {
	var b strings.Builder
	writePHPValue(&b, v, 0)
	return b.String()
}

func writePHPValue(b *strings.Builder, v any, depth int) {
	switch val := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		if val {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case string:
		b.WriteString(`"`)
		b.WriteString(escapePHPString(val))
		b.WriteString(`"`)
	case json.Number:
		b.WriteString(numberToPHPString(val))
	case float64:
		b.WriteString(FloatToPHPString(val))
	case float32:
		b.WriteString(FloatToPHPString(float64(val)))
	case int:
		b.WriteString(strconv.Itoa(val))
	case int64:
		b.WriteString(strconv.FormatInt(val, 10))
	case uint64:
		b.WriteString(strconv.FormatUint(val, 10))
	case []any:
		if len(val) == 0 {
			b.WriteString("[]")
			return
		}
		b.WriteString("[\n")
		inner := strings.Repeat("    ", depth+1)
		for i, e := range val {
			b.WriteString(inner)
			writePHPValue(b, e, depth+1)
			if i < len(val)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString(strings.Repeat("    ", depth))
		b.WriteString("]")
	case *Ordered:
		if len(val.keys) == 0 {
			b.WriteString("{}")
			return
		}
		b.WriteString("{\n")
		inner := strings.Repeat("    ", depth+1)
		for i, k := range val.keys {
			b.WriteString(inner)
			b.WriteString(`"`)
			b.WriteString(escapePHPString(k))
			b.WriteString(`": `)
			writePHPValue(b, val.values[k], depth+1)
			if i < len(val.keys)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString(strings.Repeat("    ", depth))
		b.WriteString("}")
	default:
		// Fallback for any other Go value.
		data, err := json.Marshal(val)
		if err != nil {
			b.WriteString("null")
			return
		}
		b.WriteString(string(data))
	}
}

// escapePHPString escapes a string the way PHP json_encode does without
// JSON_UNESCAPED_UNICODE: non-ASCII characters are written as \uXXXX
// (lowercase hex), astral characters as surrogate pairs.
func escapePHPString(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '/':
			b.WriteString(`\/`)
		case '\b':
			b.WriteString(`\b`)
		case '\t':
			b.WriteString(`\t`)
		case '\n':
			b.WriteString(`\n`)
		case '\f':
			b.WriteString(`\f`)
		case '\r':
			b.WriteString(`\r`)
		default:
			switch {
			case r < 0x20:
				fmt.Fprintf(&b, `\u%04x`, r)
			case r < 0x80:
				b.WriteRune(r)
			case r <= 0xFFFF:
				fmt.Fprintf(&b, `\u%04x`, r)
			default:
				r -= 0x10000
				fmt.Fprintf(&b, `\u%04x\u%04x`, 0xD800+(r>>10), 0xDC00+(r&0x3FF))
			}
		}
	}
	return b.String()
}

// numberToPHPString converts a json.Number literal the way PHP re-encodes
// a decoded JSON number: integer literals stay verbatim (as long as they fit
// into a PHP int), everything else is treated as a float and formatted with
// PHP float semantics.
func numberToPHPString(n json.Number) string {
	s := n.String()
	if isIntegerLiteral(s) {
		if _, err := strconv.ParseInt(s, 10, 64); err == nil {
			return s
		}
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return s
	}
	return FloatToPHPString(f)
}

func isIntegerLiteral(s string) bool {
	if s == "" {
		return false
	}
	i := 0
	if s[0] == '-' || s[0] == '+' {
		i++
	}
	if i >= len(s) {
		return false
	}
	for ; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// floatToPHPString formats a float like PHP's json_encode with
// serialize_precision=-1 (php_gcvt with 17 significant digits): decimal
// notation while the decimal exponent is in [-4, 17), scientific notation
// otherwise (e.g. 5968115, 0.5, -0, 1.0e+21, 1.5e-5).
func FloatToPHPString(f float64) string {
	if f != f || f > math.MaxFloat64 || f < -math.MaxFloat64 {
		return "null"
	}
	if f == 0 {
		if math.Signbit(f) {
			return "-0"
		}
		return "0"
	}
	exp := math.Floor(math.Log10(math.Abs(f)))
	if exp >= -4 && exp < 17 {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	s := strconv.FormatFloat(f, 'e', -1, 64)
	mantissa, expPart, _ := strings.Cut(s, "e")
	expPart = normalizeExponent(expPart)
	if !strings.Contains(mantissa, ".") {
		mantissa += ".0"
	}
	return mantissa + "e" + expPart
}

// normalizeExponent strips leading zeroes from an exponent like PHP does:
// "+05" -> "+5", "-05" -> "-5".
func normalizeExponent(exp string) string {
	if len(exp) > 2 && (exp[0] == '+' || exp[0] == '-') && exp[1] == '0' {
		return exp[:1] + exp[2:]
	}
	return exp
}
