package prompts

import (
	"os"
	"strings"
	"testing"
)

// renderFrame renders a prompt frame with a fixed 80x24 terminal.
func renderFrame(t *testing.T, p *Prompt) string {
	t.Helper()
	t.Setenv("COLUMNS", "80")
	t.Setenv("LINES", "24")
	p.newLinesWritten = 1
	return p.RenderTheme()
}

func compareGolden(t *testing.T, name, got string) {
	t.Helper()
	want, err := os.ReadFile("../../testdata/prompts/golden_" + name + ".txt")
	if err != nil {
		t.Fatalf("golden file %s: %v", name, err)
	}
	// Strip the cursor-restore sequence captured after submit frames.
	wantStr := strings.TrimSuffix(string(want), "\x1b[?25h")
	if got != wantStr {
		t.Errorf("frame %s differs.\nGOT:\n%s\nWANT:\n%s", name, got, wantStr)
	}
}

func testSelectPrompt() *SelectPrompt {
	options := []SelectOption{
		{Key: "_0", Label: "Eingeschränkter Betrieb zwischen Basel, Aeschenplatz und Basel, Burgfelderplatz"},
		{Key: "_1", Label: "Unregelmässiger Betrieb im Bereich Basel, Bruderholz"},
	}
	p := &SelectPrompt{
		Prompt:      newPrompt(),
		label:       "Which alert would you like to read?",
		options:     options,
		scroll:      5,
		highlighted: 0,
	}
	p.rendererFn = p.renderer
	p.valueFn = p.value
	p.reduceScrollingToFitTerminal()
	return p
}

func TestSelectFrames(t *testing.T) {
	p := testSelectPrompt()
	compareGolden(t, "select_active", renderFrame(t, p.Prompt))

	p.State = "submit"
	compareGolden(t, "select_submit", renderFrame(t, p.Prompt))

	p.State = "initial"
	p.highlighted = 1
	p.State = "submit"
	compareGolden(t, "select_down_submit", renderFrame(t, p.Prompt))
}

func testSearchPrompt() *SearchPrompt {
	options := []SelectOption{
		{Key: "ch:23005:300", Label: "Basel, SBB"},
		{Key: "ch:23005:7", Label: "Basel, Bahnhof SBB"},
		{Key: "ch:23005:64", Label: "Basel, St.Jakob"},
	}
	p := &SearchPrompt{
		Prompt:      newPrompt(),
		label:       "Witch Stop do you want to see the departures for?",
		placeholder: "Basel, Claraplatz",
		optionsFn: func(term string) []SelectOption {
			if term == "bas" {
				return options
			}
			return nil
		},
		scroll: 5,
	}
	p.rendererFn = p.renderer
	p.valueFn = p.value
	p.reduceScrollingToFitTerminal()
	return p
}

func TestSearchFrames(t *testing.T) {
	p := testSearchPrompt()
	// Stale cache: results from a previous 'bas' search.
	p.matchesCache = []SelectOption{
		{Key: "ch:23005:300", Label: "Basel, SBB"},
		{Key: "ch:23005:7", Label: "Basel, Bahnhof SBB"},
		{Key: "ch:23005:64", Label: "Basel, St.Jakob"},
	}
	p.matchesSet = true
	p.typedValue = "bas"
	p.cursorPos = 3
	compareGolden(t, "search_results", renderFrame(t, p.Prompt))

	p.highlighted = intPtr(0)
	compareGolden(t, "search_highlighted", renderFrame(t, p.Prompt))

	p.highlighted = intPtr(2)
	compareGolden(t, "search_highlighted3", renderFrame(t, p.Prompt))

	p.typedValue = "xyz"
	p.cursorPos = 3
	p.highlighted = nil
	compareGolden(t, "search_noresults", renderFrame(t, p.Prompt))

	p.State = "searching"
	compareGolden(t, "search_searching", renderFrame(t, p.Prompt))

	p.State = "active"
	p.typedValue = ""
	p.cursorPos = 0
	compareGolden(t, "search_empty", renderFrame(t, p.Prompt))

	p.State = "submit"
	p.typedValue = "bas"
	p.cursorPos = 3
	p.highlighted = intPtr(0)
	compareGolden(t, "search_submit", renderFrame(t, p.Prompt))
}

func intPtr(i int) *int { return &i }

func testTextPrompt() *TextPrompt {
	p := &TextPrompt{
		Prompt:     newPrompt(),
		label:      "At wich time would you like to departure/arrive ?",
		typedValue: "20:30",
		cursorPos:  5,
	}
	p.rendererFn = p.renderer
	p.valueFn = p.value
	return p
}

func TestTextFrames(t *testing.T) {
	p := testTextPrompt()
	compareGolden(t, "text_active", renderFrame(t, p.Prompt))

	p.State = "submit"
	compareGolden(t, "text_submit", renderFrame(t, p.Prompt))
}

func testConfirmPrompt() *ConfirmPrompt {
	p := &ConfirmPrompt{
		Prompt:    newPrompt(),
		label:     "there are alerts for this route? Would you like to see them now?",
		confirmed: true,
	}
	p.rendererFn = p.renderer
	p.valueFn = p.value
	return p
}

func TestConfirmFrames(t *testing.T) {
	p := testConfirmPrompt()
	compareGolden(t, "confirm_active", renderFrame(t, p.Prompt))

	p.State = "submit"
	compareGolden(t, "confirm_submit", renderFrame(t, p.Prompt))
}

func TestAddCursor(t *testing.T) {
	// Value with cursor at the end.
	got := addCursor("bas", 3, 74)
	want := "bas" + inverse(" ")
	if got != want {
		t.Errorf("addCursor end: %q != %q", got, want)
	}
	// Placeholder with cursor at start.
	got = addCursor("Basel, Claraplatz", 0, 74)
	want = inverse("B") + "asel, Claraplatz"
	if got != want {
		t.Errorf("addCursor start: %q != %q", got, want)
	}
}

func TestReplaceTrailingSpace(t *testing.T) {
	if got := replaceTrailingSpace("ab ", "…"); got != "ab…" {
		t.Errorf("got %q", got)
	}
	if got := replaceTrailingSpace("ab", "…"); got != "ab" {
		t.Errorf("got %q", got)
	}
}

func TestScrollPosition(t *testing.T) {
	cases := []struct {
		first, height, total, want int
	}{
		{0, 5, 20, 0},
		{15, 5, 20, 4},
		{3, 5, 20, 1},
		{7, 5, 20, 2},
	}
	for _, c := range cases {
		if got := scrollPosition(c.first, c.height, c.total); got != c.want {
			t.Errorf("scrollPosition(%d,%d,%d) = %d, want %d", c.first, c.height, c.total, got, c.want)
		}
	}
}
