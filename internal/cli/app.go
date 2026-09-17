// Package cli implements the command line interface of the Go port of
// efa-cli, replicating the Laravel Zero / Symfony Console behaviour of the
// original PHP application.
package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/simonjenny/efa-cli/internal/i18n"
	"github.com/simonjenny/efa-cli/internal/prompts"
)

// out writes to stdout unless the output is quieted.
func out(s string) {
	if prompts.Quiet() {
		return
	}
	fmt.Print(s)
}

// outf writes formatted output unless the output is quieted.
func outf(format string, args ...any) {
	if prompts.Quiet() {
		return
	}
	fmt.Printf(format, args...)
}

// Version of the application.
const Version = "v3.0"

// AppName is the application name.
const AppName = "efa-cli"

// command describes a console command (mirrors the PHP command classes).
type command struct {
	name        string
	description string
	args        []commandArg
	options     []commandOption
	run         func(c *parsedArgs) int
}

type commandArg struct {
	name        string
	description string
}

type commandOption struct {
	name          string
	short         string // short flag without dashes, e.g. "h"
	hasValue      bool
	valueOptional bool
	negatable     bool
	description   string
}

// globalOptions are the options every command accepts (Symfony globals).
var globalOptions = []commandOption{
	{name: "help", short: "h", description: "Display help for the given command. When no command is given display help for the list command"},
	{name: "quiet", short: "q", description: "Do not output any message"},
	{name: "version", short: "V", description: "Display this application version"},
	{name: "ansi", negatable: true, description: "Force (or disable --no-ansi) ANSI output"},
	{name: "no-interaction", short: "n", description: "Do not ask any interactive question"},
	{name: "verbose", short: "v|vv|vvv", description: "Increase the verbosity of messages: 1 for normal output, 2 for more verbose output and 3 for debug"},
	{name: "lang", hasValue: true, description: "Language for output (de, en) [default: auto]"},
}

// allCommands lists the available commands in display order.
var allCommands []*command

// Run is the application entry point; it returns the exit code.
func Run(args []string) int {
	// Detect the language from the locale environment variables; the
	// --lang option overrides it when present.
	_ = i18n.SetLang(i18n.DetectEnvLang())

	// Consume leading global options (like Symfony's ArgvInput).
	rest := args
	var noInter, quiet, ansi, noAnsi bool
	for len(rest) > 0 {
		if strings.HasPrefix(rest[0], "--lang=") {
			if !applyLang(strings.TrimPrefix(rest[0], "--lang=")) {
				return 1
			}
			rest = rest[1:]
			continue
		}
		switch rest[0] {
		case "-V", "--version":
			out(AppName + " " + Version + "\n")
			return 0
		case "-h", "--help":
			printListHelp()
			return 0
		case "-n", "--no-interaction":
			noInter = true
			rest = rest[1:]
		case "-q", "--quiet":
			quiet = true
			rest = rest[1:]
		case "--ansi":
			ansi = true
			rest = rest[1:]
		case "--no-ansi":
			noAnsi = true
			rest = rest[1:]
		case "-v", "-vv", "-vvv":
			rest = rest[1:]
		case "--lang":
			if len(rest) < 2 {
				commandError(i18n.T("err.option_requires_value", "--lang"), "")
				return 1
			}
			if !applyLang(rest[1]) {
				return 1
			}
			rest = rest[2:]
		default:
			goto done
		}
	}
done:
	// Build the command list after the leading options so a --lang flag
	// already switched the translation language.
	commands := buildCommands()
	applyGlobalFlags(noInter, quiet, ansi, noAnsi)
	defer resetGlobalFlags()
	if len(rest) == 0 {
		return runSummary()
	}

	first := rest[0]

	for _, c := range commands {
		if c.name == first {
			return runCommand(c, rest[1:])
		}
	}

	switch first {
	case "help":
		if len(rest) > 1 {
			for _, c := range commands {
				if c.name == rest[1] {
					printHelp(c)
					return 0
				}
			}
			return commandError(i18n.T("err.command_not_defined", rest[1]), "")
		}
		printListHelp()
		return 0
	case "list":
		return runSummary()
	}

	return commandError(i18n.T("err.command_not_defined", first), "")
}

// applyLang switches the current language, printing an error and returning
// false for unsupported languages.
func applyLang(lang string) bool {
	if err := i18n.SetLang(lang); err != nil {
		commandError(i18n.T("err.lang_not_supported", lang, strings.Join(i18n.Supported(), ", ")), "")
		return false
	}
	return true
}

// applyGlobalFlags applies the root-level option flags.
func applyGlobalFlags(noInter, quiet, ansi, noAnsi bool) {
	prompts.SetInteractive(!noInter)
	prompts.SetQuiet(quiet)
	if ansi {
		stdoutIsTTY = true
	} else if noAnsi {
		stdoutIsTTY = false
	}
}

func resetGlobalFlags() {
	prompts.SetInteractive(true)
	prompts.SetQuiet(false)
	stdoutIsTTY = isTerminalOutput()
}

// buildCommands constructs the command list.
func buildCommands() []*command {
	if allCommands != nil {
		return allCommands
	}
	allCommands = []*command{
		{
			name:        "departures",
			description: "Create a departure schedule for a specific bus stop.",
			args: []commandArg{
				{name: "stop", description: "Displays the departures for this stop (optional)"},
			},
			options: []commandOption{
				{name: "limit", hasValue: true, valueOptional: true, description: "Limits the number of displayed departures (default is 10, optional)"},
				{name: "gid", description: "Provided Stop ID is a GID (optional)"},
				{name: "json", description: "Shows data as JSON (optional)"},
			},
			run: runDepartures,
		},
		{
			name:        "messages",
			description: "Show current information, disruptions and alerts (currently only from the Basler Verkehrs-Betriebe network in german!)",
			options: []commandOption{
				{name: "json", description: "Shows data as JSON (optional)"},
			},
			run: runMessages,
		},
		{
			name:        "route",
			description: "Plan a trip from point A to point B",
			args: []commandArg{
				{name: "start", description: "Starting Stop (optional)"},
				{name: "destination", description: "Destination Stop (optional)"},
				{name: "time", description: "Time for arrival/departure (optional)"},
				{name: "date", description: "Date for arrival/departure (optional)"},
				{name: "mode", description: "Show trips on date for arrival or departure (optional)"},
			},
			options: []commandOption{
				{name: "json", description: "Shows data as JSON (optional)"},
			},
			run: runRoute,
		},
		{
			name:        "stopinfo",
			description: "Show Information for a stop.",
			args: []commandArg{
				{name: "stop", description: "Show Information for this stop (optional)"},
			},
			options: []commandOption{
				{name: "json", description: "Shows data as JSON (optional)"},
			},
			run: runStopinfo,
		},
		{
			name:        "mcp",
			description: "Start an MCP server exposing departures, messages, route and stopinfo as tools.",
			options: []commandOption{
				{name: "ip", hasValue: true, description: "IP address to bind the MCP server to (default 127.0.0.1)"},
				{name: "port", hasValue: true, description: "Port to bind the MCP server to (default 8090)"},
			},
			run: runMcp,
		},
	}
	return allCommands
}

// parsedArgs holds the result of parsing a command line.
type parsedArgs struct {
	command   *command
	values    map[string]string
	flags     map[string]string
	flagSet   map[string]bool
	help      bool
	version   bool
	noInter   bool
	quiet     bool
	ansi      bool
	noAnsi    bool
	verbosity int
}

func (p *parsedArgs) has(name string) bool {
	return p.flagSet[name]
}

func (p *parsedArgs) value(name string) string {
	return p.flags[name]
}

func (p *parsedArgs) arg(name string) string {
	return p.values[name]
}

// parseArgs parses a command line, mirroring Symfony's ArgvInput for the
// options and arguments used by this application.
func parseArgs(c *command, tokens []string) (*parsedArgs, int) {
	p := &parsedArgs{
		command: c,
		values:  map[string]string{},
		flags:   map[string]string{},
		flagSet: map[string]bool{},
	}
	optionByName := map[string]*commandOption{}
	optionByShort := map[string]*commandOption{}
	for i := range c.options {
		optionByName[c.options[i].name] = &c.options[i]
	}
	for i := range globalOptions {
		g := &globalOptions[i]
		name := g.name
		if i := strings.Index(name, "|"); i != -1 {
			optionByName[name[:i]] = g
			optionByName[name[i+1:]] = g
		} else {
			optionByName[name] = g
		}
		if g.short != "" {
			optionByShort[g.short] = g
		}
	}

	var positionals []string
	afterSeparator := false
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		if afterSeparator {
			positionals = append(positionals, token)
			continue
		}
		if token == "--" {
			afterSeparator = true
			continue
		}
		if strings.HasPrefix(token, "--") {
			body := token[2:]
			name := body
			value := ""
			hasValue := false
			if eq := strings.Index(body, "="); eq != -1 {
				name = body[:eq]
				value = body[eq+1:]
				hasValue = true
			}
			opt, ok := optionByName[name]
			if !ok {
				return p, commandError(i18n.T("err.option_not_exist", "--"+name), synopsis(c))
			}
			p.flagSet[opt.name] = true
			if opt.hasValue {
				if !hasValue {
					if i+1 < len(tokens) && !strings.HasPrefix(tokens[i+1], "-") {
						i++
						value = tokens[i]
					}
				}
				p.flags[opt.name] = value
			} else {
				if hasValue {
					return p, commandError(i18n.T("err.option_no_value", "--"+name), synopsis(c))
				}
				p.flags[opt.name] = ""
			}
			switch opt.name {
			case "help":
				p.help = true
			case "version":
				p.version = true
			case "no-interaction":
				p.noInter = true
			case "quiet":
				p.quiet = true
			case "ansi":
				p.ansi = true
			case "no-ansi":
				p.noAnsi = true
			case "verbose":
				p.verbosity++
			case "lang":
				if err := i18n.SetLang(value); err != nil {
					return p, commandError(i18n.T("err.lang_not_supported", value, strings.Join(i18n.Supported(), ", ")), synopsis(c))
				}
			}
			continue
		}
		if strings.HasPrefix(token, "-") && len(token) > 1 {
			short := token[1:]
			opt, ok := optionByShort[short]
			if !ok {
				if allSameRune(short, 'v') {
					p.verbosity += len(short)
					p.flagSet["verbose"] = true
					continue
				}
				return p, commandError(i18n.T("err.option_not_exist", "-"+short), synopsis(c))
			}
			p.flagSet[opt.name] = true
			switch opt.name {
			case "help":
				p.help = true
			case "version":
				p.version = true
			case "no-interaction":
				p.noInter = true
			case "quiet":
				p.quiet = true
			case "ansi":
				p.ansi = true
			case "no-ansi":
				p.noAnsi = true
			case "verbose":
				p.verbosity++
			}
			continue
		}
		positionals = append(positionals, token)
	}

	if p.help {
		return p, -1
	}
	if p.version {
		out(AppName + " " + Version + "\n")
		return p, 100
	}

	if len(positionals) > len(c.args) {
		var expected []string
		for _, a := range c.args {
			expected = append(expected, a.name)
		}
		return p, commandError(i18n.T("err.too_many_arguments", c.name, strings.Join(expected, " ")), synopsis(c))
	}
	for i, a := range c.args {
		if i < len(positionals) {
			p.values[a.name] = positionals[i]
		} else {
			p.values[a.name] = ""
		}
	}
	return p, 0
}

func allSameRune(s string, r rune) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c != r {
			return false
		}
	}
	return true
}

// synopsis builds the Symfony short synopsis used in error messages:
// "departures [--limit [LIMIT]] [--gid] [--json] [--] [<stop>]".
func synopsis(c *command) string {
	var b strings.Builder
	b.WriteString(c.name)
	for _, o := range c.options {
		b.WriteString(" [--" + o.name)
		if o.hasValue {
			b.WriteString(" [" + strings.ToUpper(o.name) + "]")
		}
		b.WriteString("]")
	}
	if len(c.args) > 0 {
		b.WriteString(" [--]")
		for _, a := range c.args {
			b.WriteString(" [<" + a.name + ">")
		}
		b.WriteString(strings.Repeat("]", len(c.args)))
	}
	return b.String()
}

// usageLine builds the Symfony usage line shown in help.
func usageLine(c *command) string {
	var b strings.Builder
	b.WriteString(c.name)
	b.WriteString(" [options]")
	if len(c.args) > 0 {
		b.WriteString(" [--]")
		for i, a := range c.args {
			if i == 0 {
				b.WriteString(" [<" + a.name + ">")
			} else {
				b.WriteString(" [<" + a.name + ">")
			}
		}
		b.WriteString(strings.Repeat("]", len(c.args)))
	}
	return b.String()
}

// runCommand dispatches a parsed command line.
func runCommand(c *command, tokens []string) int {
	p, code := parseArgs(c, tokens)
	if code == 100 {
		return 0
	}
	if code == -1 {
		printHelp(c)
		return 0
	}
	if code > 0 {
		return code
	}
	// Command-level flags override the root flags.
	if p.noInter {
		prompts.SetInteractive(false)
	}
	if p.quiet {
		prompts.SetQuiet(true)
	}
	if p.ansi {
		stdoutIsTTY = true
	} else if p.noAnsi {
		stdoutIsTTY = false
	}
	return c.run(p)
}

// commandError prints an error the way Symfony renders an exception.
func commandError(message, usage string) int {
	line := "  " + message + "  "
	pad := strings.Repeat(" ", len(line))
	outf("\n%s\n%s\n%s\n\n", pad, line, pad)
	if usage != "" {
		out(usage + "\n")
		out("\n")
	}
	return 1
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// printHelp renders the Symfony-style help for a command.
func printHelp(c *command) {
	var b strings.Builder
	b.WriteString(i18n.T("help.description") + "\n")
	b.WriteString("  " + commandDescription(c) + "\n\n")
	b.WriteString(i18n.T("help.usage") + "\n")
	b.WriteString("  " + usageLine(c) + "\n")

	allOpts := append(append([]commandOption{}, c.options...), globalOptions...)
	totalWidth := 0
	for _, o := range allOpts {
		if w := optionWidth(o); w > totalWidth {
			totalWidth = w
		}
	}
	for _, a := range c.args {
		if len(a.name) > totalWidth {
			totalWidth = len(a.name)
		}
	}

	if len(c.args) > 0 {
		b.WriteString("\n" + i18n.T("help.arguments") + "\n")
		for _, a := range c.args {
			b.WriteString("  " + a.name + "  " + strings.Repeat(" ", totalWidth-len(a.name)) + argDescription(c, a) + "\n")
		}
	}
	b.WriteString("\n" + i18n.T("help.options") + "\n")
	writeOptions(&b, allOpts, totalWidth)
	out(b.String())
}

// binaryName returns the base name of the current executable.
func binaryName() string {
	name := os.Args[0]
	if i := strings.LastIndexAny(name, "/\\"); i != -1 {
		name = name[i+1:]
	}
	return name
}

// optionWidth mirrors TextDescriptor::calculateTotalWidthForOptions.
func optionWidth(o commandOption) int {
	nameLen := 1 + maxInt(len(o.short), 1) + 4 + len(o.name)
	if o.negatable {
		nameLen += 6 + len(o.name)
	} else if o.hasValue {
		valueLen := 1 + len(o.name)
		if o.valueOptional {
			valueLen += 2 // [ + ]
		}
		nameLen += valueLen
	}
	return nameLen
}

// writeOptions writes the options block with descriptions aligned to the
// Symfony description column.
func writeOptions(b *strings.Builder, opts []commandOption, totalWidth int) {
	// Options with multi-character shortcuts are rendered last.
	var later []commandOption
	for i := range opts {
		if len(opts[i].short) > 1 {
			later = append(later, opts[i])
			continue
		}
		writeOption(b, opts[i], totalWidth)
	}
	for i := range later {
		writeOption(b, later[i], totalWidth)
	}
}

func writeOption(b *strings.Builder, o commandOption, totalWidth int) {
	synopsis := optionSynopsis(o)
	b.WriteString("  " + synopsis + "  " + strings.Repeat(" ", totalWidth-len(synopsis)) + optionDescription(o) + "\n")
}

// optionDescription translates the option description for the current
// language, falling back to the raw description for unknown options.
func optionDescription(o commandOption) string {
	key := "opt." + o.name
	if t := i18n.T(key); t != key {
		return t
	}
	return o.description
}

// commandDescription translates the command description for the current
// language, falling back to the raw description.
func commandDescription(c *command) string {
	key := "cmd." + c.name
	if t := i18n.T(key); t != key {
		return t
	}
	return c.description
}

// argDescription translates the argument description for the current
// language, falling back to the raw description.
func argDescription(c *command, a commandArg) string {
	key := "arg." + c.name + "." + a.name
	if t := i18n.T(key); t != key {
		return t
	}
	return a.description
}

func optionSynopsis(o commandOption) string {
	if o.short != "" {
		return "-" + o.short + ", --" + o.name + optionValueSuffix(o)
	}
	return "    --" + o.name + optionValueSuffix(o)
}

func optionValueSuffix(o commandOption) string {
	if o.negatable {
		return "|--no-" + o.name
	}
	if o.hasValue {
		if o.valueOptional {
			return "[=" + strings.ToUpper(o.name) + "]"
		}
		return "=" + strings.ToUpper(o.name)
	}
	return ""
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// runSummary prints the Laravel Zero summary (Describer output).
func runSummary() int {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString("  " + decorated(AppName+" ", "white-bold") + " " + decorated(Version, "green-bold") + "\n\n")
	b.WriteString("  " + decorated(i18n.T("summary.usage"), "yellow-bold") + "  <command> [options] [arguments]\n\n")
	width := 0
	for _, c := range allCommands {
		if len(c.name) > width {
			width = len(c.name)
		}
	}
	sorted := append([]*command{}, allCommands...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].name < sorted[j].name })
	for _, c := range sorted {
		b.WriteString("  " + decorated(c.name, "green") + strings.Repeat(" ", width-len(c.name)+1) + commandDescription(c) + "\n")
	}
	b.WriteString("\n")
	out(b.String())
	return 0
}

// printListHelp prints the help of the list command (efa --help).
func printListHelp() {
	var b strings.Builder
	b.WriteString(i18n.T("help.description") + "\n")
	b.WriteString("  " + i18n.T("help.list_description") + "\n\n")
	b.WriteString(i18n.T("help.usage") + "\n")
	b.WriteString("  list [options] [--] [<namespace>]\n\n")
	listOptions := []commandOption{
		{name: "raw", description: i18n.T("opt.raw")},
		{name: "format", hasValue: true, description: i18n.T("opt.format")},
		{name: "short", description: i18n.T("opt.short")},
	}
	allOpts := append(append([]commandOption{}, listOptions...), globalOptions...)
	totalWidth := 0
	for _, o := range allOpts {
		if w := optionWidth(o); w > totalWidth {
			totalWidth = w
		}
	}
	if 9 > totalWidth {
		totalWidth = 9
	}
	b.WriteString(i18n.T("help.arguments") + "\n")
	b.WriteString("  " + "namespace" + "  " + strings.Repeat(" ", totalWidth-9) + i18n.T("help.namespace_arg") + "\n")
	b.WriteString("\n" + i18n.T("help.options") + "\n")
	writeOptions(&b, allOpts, totalWidth)
	b.WriteString("\n" + i18n.T("help.help") + "\n")
	helpText := i18n.T("help.list_intro", decorated("list", "info")) + "\n\n" +
		"  " + decorated(binaryName()+" list", "info") + "\n\n" +
		i18n.T("help.list_namespace") + "\n\n" +
		"  " + decorated(binaryName()+" list test", "info") + "\n\n" +
		i18n.T("help.list_format", decorated("--format", "comment")) + "\n\n" +
		"  " + decorated(binaryName()+" list --format=xml", "info") + "\n\n" +
		i18n.T("help.list_raw") + "\n\n" +
		"  " + decorated(binaryName()+" list --raw", "info")
	b.WriteString("  " + strings.ReplaceAll(helpText, "\n", "\n  ") + "\n")
	out(b.String())
}
