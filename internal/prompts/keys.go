package prompts

// Key constants mirroring Laravel\Prompts\Key.
const (
	KeyUp         = "\x1b[A"
	KeyShiftUp    = "\x1b[1;2A"
	KeyDown       = "\x1b[B"
	KeyShiftDown  = "\x1b[1;2B"
	KeyRight      = "\x1b[C"
	KeyLeft       = "\x1b[D"
	KeyUpArrow    = "\x1bOA"
	KeyDownArrow  = "\x1bOB"
	KeyRightArrow = "\x1bOC"
	KeyLeftArrow  = "\x1bOD"
	KeyEscape     = "\x1b"
	KeyDelete     = "\x1b[3~"
	KeyBackspace  = "\x7f"
	KeyEnter      = "\n"
	KeySpace      = " "
	KeyTab        = "\t"
	KeyShiftTab   = "\x1b[Z"
	KeyCtrlC      = "\x03"
	KeyCtrlP      = "\x10"
	KeyCtrlN      = "\x0e"
	KeyCtrlF      = "\x06"
	KeyCtrlB      = "\x02"
	KeyCtrlH      = "\x08"
	KeyCtrlA      = "\x01"
	KeyCtrlD      = "\x04"
	KeyCtrlE      = "\x05"
	KeyCtrlU      = "\x15"
)

// KeyHome is the set of HOME escape sequences.
var KeyHome = []string{"\x1b[1~", "\x1bOH", "\x1b[H", "\x1b[7~"}

// KeyEnd is the set of END escape sequences.
var KeyEnd = []string{"\x1b[4~", "\x1bOF", "\x1b[F", "\x1b[8~"}

// KeyOneOf reports whether key matches any of the given keys
// (arrays are flattened, matching Key::oneOf).
func KeyOneOf(keys []string, key string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}
