package prompts

// ANSI colour helpers matching Laravel\Prompts\Concerns\Colors.
func reset(s string) string     { return "\x1b[0m" + s + "\x1b[0m" }
func bold(s string) string      { return "\x1b[1m" + s + "\x1b[22m" }
func dim(s string) string       { return "\x1b[2m" + s + "\x1b[22m" }
func italic(s string) string    { return "\x1b[3m" + s + "\x1b[23m" }
func underline(s string) string { return "\x1b[4m" + s + "\x1b[24m" }
func inverse(s string) string   { return "\x1b[7m" + s + "\x1b[27m" }
func hidden(s string) string    { return "\x1b[8m" + s + "\x1b[28m" }
func strikethrough(s string) string {
	return "\x1b[9m" + s + "\x1b[29m"
}
func black(s string) string   { return "\x1b[30m" + s + "\x1b[39m" }
func red(s string) string     { return "\x1b[31m" + s + "\x1b[39m" }
func green(s string) string   { return "\x1b[32m" + s + "\x1b[39m" }
func yellow(s string) string  { return "\x1b[33m" + s + "\x1b[39m" }
func blue(s string) string    { return "\x1b[34m" + s + "\x1b[39m" }
func magenta(s string) string { return "\x1b[35m" + s + "\x1b[39m" }
func cyan(s string) string    { return "\x1b[36m" + s + "\x1b[39m" }
func white(s string) string   { return "\x1b[37m" + s + "\x1b[39m" }
func gray(s string) string    { return "\x1b[90m" + s + "\x1b[39m" }

func bgBlack(s string) string   { return "\x1b[40m" + s + "\x1b[49m" }
func bgRed(s string) string     { return "\x1b[41m" + s + "\x1b[49m" }
func bgGreen(s string) string   { return "\x1b[42m" + s + "\x1b[49m" }
func bgYellow(s string) string  { return "\x1b[43m" + s + "\x1b[49m" }
func bgBlue(s string) string    { return "\x1b[44m" + s + "\x1b[49m" }
func bgMagenta(s string) string { return "\x1b[45m" + s + "\x1b[49m" }
func bgCyan(s string) string    { return "\x1b[46m" + s + "\x1b[49m" }
func bgWhite(s string) string   { return "\x1b[47m" + s + "\x1b[49m" }
