package table

import (
	"os"
	"strings"
	"testing"
)

func TestGoldenTables(t *testing.T) {
	cases := []struct {
		name    string
		headers []string
		rows    [][]string
	}{
		{
			name:    "departures",
			headers: []string{"", "Nr.", "Destination", "Departure"},
			rows: [][]string{
				{"🚊", "11", "Reinach BL, Surbaum", "25 seconds from now"},
				{"🚊", "16", "Schifflände", "3 minutes from now"},
				{"🚌", "30", "Weil am Rhein Bahnhof/Zentrum", "11 minutes from now"},
			},
		},
		{
			name:    "stopinfo",
			headers: []string{"Basel, SBB", ""},
			rows: [][]string{
				{"EFA Stop ID", "51000300"},
				{"GID", "ch:23005:300"},
				{"Coordinates", "7.589354,47.547394"},
				{"Google Maps", "https://www.google.com/maps/search/?api=1&query=47.547394,7.589354"},
				{"Web Departure Monitor", "https://dfi.bvb.ch/?point=51000300"},
				{"", ""},
				{"Info", "Eingeschränkter Betrieb\nDetailed information available at http://info.bvb.ch (German only)"},
			},
		},
		{
			name:    "messages",
			headers: []string{"Eingeschränkter Betrieb zwischen Basel, Aeschenplatz und Basel, Burgfelderplatz"},
			rows: [][]string{
				{"Eingeschränkter Betrieb zwischen Basel, Aeschenplatz und Basel, Burgfelderplatz" +
					"\n\nBetroffen ist die Linie 3." +
					"\n\nDer Grund dafür sind Bauarbeiten." +
					"\n\nDie Einschränkung dauert von ca. 22.06.2026, 05:00 Uhr bis ca. 28.09.2026, 00:30 Uhr." +
					"\n\nDie Linie 3 wird zwischen Aeschenplatz und Burgfelderplatz umgeleitet. Die Umleitung verläuft über die Haltestellen Basel, Bahnhof SBB und Basel, Brausebad."},
			},
		},
	}
	for _, c := range cases {
		lines := Render(c.headers, c.rows)
		var b strings.Builder
		for _, l := range lines {
			b.WriteString(" " + l + "\n")
		}
		want, err := os.ReadFile("../../testdata/table/" + c.name + ".txt")
		if err != nil {
			t.Fatalf("golden file for %s: %v", c.name, err)
		}
		// The golden files include the leading and trailing blank lines
		// written by the prompts table renderer.
		got := "\n" + b.String() + "\n"
		if got != string(want) {
			t.Errorf("table %s differs:\nGOT:\n%sWANT:\n%s", c.name, got, want)
		}
	}
}

func TestRenderEmpty(t *testing.T) {
	lines := Render([]string{}, [][]string{})
	if len(lines) != 0 {
		t.Errorf("empty table should render nothing, got %q", lines)
	}
}

func TestRenderNoHeaders(t *testing.T) {
	lines := Render(nil, [][]string{{"a"}, {"b"}})
	want := []string{"┌───┐", "│ a │", "│ b │", "└───┘"}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("got %q want %q", lines, want)
	}
}
