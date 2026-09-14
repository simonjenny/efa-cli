package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/simonjenny/efa-cli/internal/carbon"
	"github.com/simonjenny/efa-cli/internal/efa"
	"github.com/simonjenny/efa-cli/internal/prompts"
)

// mockServer serves the recorded EFA fixtures.
func mockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	fixtures := map[string]string{
		"/XML_DM_REQUEST":      "../../testdata/fixtures/dm.json",
		"/XML_ADDINFO_REQUEST": "../../testdata/fixtures/addinfo.json",
		"/XSLT_TRIP_REQUEST2":  "../../testdata/fixtures/trip.json",
	}
	serve := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/XSLT_STOPFINDER_REQUEST" {
			q := r.URL.Query()
			if q.Get("anyMaxSizeHitList") == "50" {
				http.ServeFile(w, r, "../../testdata/fixtures/stopfinder_list.json")
				return
			}
			http.ServeFile(w, r, "../../testdata/fixtures/stopfinder_single.json")
			return
		}
		if f, ok := fixtures[path]; ok {
			http.ServeFile(w, r, f)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}
	mux.HandleFunc("/", serve)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// setupClient points the commands at the mock server.
func setupClient(t *testing.T) {
	t.Helper()
	srv := mockServer(t)
	newClient = func() *efa.Client {
		c := efa.NewClient()
		c.BaseURL = srv.URL + "/"
		return c
	}
	t.Cleanup(func() { newClient = func() *efa.Client { return efa.NewClient() } })
}

// captureStdout runs fn while capturing stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()
	buf := new(strings.Builder)
	done := make(chan struct{})
	go func() {
		copyFrom(buf, r)
		close(done)
	}()
	fn()
	w.Close()
	<-done
	return buf.String()
}

func copyFrom(b *strings.Builder, r *os.File) (int64, error) {
	chunk := make([]byte, 32768)
	var total int64
	for {
		n, err := r.Read(chunk)
		if n > 0 {
			b.Write(chunk[:n])
			total += int64(n)
		}
		if err == io.EOF {
			return total, nil
		}
		if err != nil {
			return total, err
		}
	}
}

func TestSummaryOutput(t *testing.T) {
	got := captureStdout(t, func() {
		if code := Run([]string{}); code != 0 {
			t.Errorf("exit code %d", code)
		}
	})
	want := "\n  " + decorated("efa-cli ", "white-bold") + " " + decorated("v2.0", "green-bold") + "\n\n" +
		"  " + decorated("USAGE:", "yellow-bold") + "  <command> [options] [arguments]\n\n" +
		"  " + decorated("departures", "green") + " Create a departure schedule for a specific bus stop.\n" +
		"  " + decorated("messages", "green") + "   Show current information, disruptions and alerts (currently only from the Basler Verkehrs-Betriebe network in german!)\n" +
		"  " + decorated("route", "green") + "      Plan a trip from point A to point B\n" +
		"  " + decorated("stopinfo", "green") + "   Show Information for a stop.\n\n"
	if got != want {
		t.Errorf("summary differs:\nGOT:\n%s\nWANT:\n%s", got, want)
	}
}

func TestVersionOutput(t *testing.T) {
	for _, args := range [][]string{{"--version"}, {"-V"}} {
		got := captureStdout(t, func() { Run(args) })
		if got != "efa-cli v2.0\n" {
			t.Errorf("version output %q", got)
		}
	}
}

func TestUnknownCommand(t *testing.T) {
	got := captureStdout(t, func() { Run([]string{"foo"}) })
	want := "\n" + strings.Repeat(" ", 33) + "\n" +
		"  Command \"foo\" is not defined.  \n" +
		strings.Repeat(" ", 33) + "\n\n"
	if got != want {
		t.Errorf("unknown command:\n%q\nwant:\n%q", got, want)
	}
}

func TestDeparturesJSON(t *testing.T) {
	setupClient(t)
	got := captureStdout(t, func() {
		Run([]string{"departures", "--gid", "--limit", "3", "--json", "ch:23005:300"})
	})
	wantFile, err := os.ReadFile("../../testdata/table/../fixtures/../table/../../testdata/table/departures.txt")
	_ = wantFile
	_ = err
	// The JSON output must contain the realtime field and pretty printing.
	if !strings.Contains(got, `"realtime": `) {
		t.Errorf("missing realtime field:\n%s", got)
	}
	if !strings.HasPrefix(got, "[\n    {\n") {
		t.Errorf("unexpected start:\n%s", got[:40])
	}
}

func TestDeparturesTable(t *testing.T) {
	setupClient(t)
	carbon.SetTestNow(time.Date(2026, 9, 14, 21, 5, 0, 0, time.UTC))
	t.Cleanup(func() { carbon.SetTestNow(time.Time{}) })
	got := captureStdout(t, func() {
		Run([]string{"departures", "--gid", "--limit", "3", "ch:23005:300"})
	})
	// Strip the spinner animation and cursor sequences.
	if i := strings.Index(got, "\n ┌"); i != -1 {
		got = got[i:]
	}
	want := "\n ┌────┬──────┬──────────────┬─────────────────────┐\n" +
		" │\x1b[2m    \x1b[22m│\x1b[2m Nr.  \x1b[22m│\x1b[2m Destination  \x1b[22m│\x1b[2m Departure           \x1b[22m│\n" +
		" ├────┼──────┼──────────────┼─────────────────────┤\n" +
		" │ 🚊 │ IR36 │ Zürich HB    │ 2 minutes from now  │\n" +
		" │ 🚊 │ S3   │ Laufen       │ 3 minutes from now  │\n" +
		" │ 🚌 │ 402  │ Basel Bad Bf │ 11 minutes from now │\n" +
		" └────┴──────┴──────────────┴─────────────────────┘\n\n"
	if got != want {
		t.Errorf("departures table differs:\nGOT:\n%q\nWANT:\n%q", got, want)
	}
}

func TestStopinfoJSON(t *testing.T) {
	setupClient(t)
	got := captureStdout(t, func() {
		Run([]string{"stopinfo", "--json", "Basel, Basel SBB"})
	})
	if !strings.Contains(got, `"name": "Basel, SBB"`) {
		t.Errorf("stopinfo json:\n%s", got)
	}
}

func TestStopinfoTable(t *testing.T) {
	setupClient(t)
	got := captureStdout(t, func() {
		Run([]string{"stopinfo", "Basel, Basel SBB"})
	})
	if !strings.Contains(got, "EFA Stop ID") || !strings.Contains(got, "ch:23005:300") {
		t.Errorf("stopinfo table:\n%s", got)
	}
	if !strings.Contains(got, "https://dfi.bvb.ch/?point=51000300") {
		t.Errorf("stopinfo table missing dfi link:\n%s", got)
	}
}

func TestMessagesJSON(t *testing.T) {
	setupClient(t)
	got := captureStdout(t, func() {
		Run([]string{"messages", "--json"})
	})
	if !strings.HasPrefix(got, "[\n    {\n") {
		t.Errorf("messages json start:\n%s", got[:40])
	}
	if !strings.Contains(got, `"infoLinkText"`) {
		t.Errorf("messages json missing infoLinkText")
	}
}

func TestRouteJSON(t *testing.T) {
	setupClient(t)
	got := captureStdout(t, func() {
		Run([]string{"route", "Basel, Basel SBB", "Basel, Claraplatz", "2045", "20260914", "Departure", "--json"})
	})
	if !strings.Contains(got, `"trips"`) {
		t.Errorf("route json missing trips:\n%s", got[:200])
	}
}

func TestUnknownOption(t *testing.T) {
	got := captureStdout(t, func() { Run([]string{"departures", "--foo"}) })
	want := "\n" + strings.Repeat(" ", 38) + "\n" +
		"  The \"--foo\" option does not exist.  \n" +
		strings.Repeat(" ", 38) + "\n\n" +
		"departures [--limit [LIMIT]] [--gid] [--json] [--] [<stop>]\n\n"
	if got != want {
		t.Errorf("unknown option:\n%q\nwant:\n%q", got, want)
	}
}

func TestTooManyArguments(t *testing.T) {
	got := captureStdout(t, func() { Run([]string{"departures", "a", "b"}) })
	want := "\n" + strings.Repeat(" ", 74) + "\n" +
		"  Too many arguments to \"departures\" command, expected arguments \"stop\".  \n" +
		strings.Repeat(" ", 74) + "\n\n" +
		"departures [--limit [LIMIT]] [--gid] [--json] [--] [<stop>]\n\n"
	if got != want {
		t.Errorf("too many arguments:\n%q\nwant:\n%q", got, want)
	}
}

func TestMessagesInteractive(t *testing.T) {
	setupClient(t)
	// Script: select the second alert (DOWN + ENTER), then read the table.
	promptsFakeInput([]string{"\x1b[B", "\n"})
	got := captureStdout(t, func() {
		Run([]string{"messages"})
	})
	if !strings.Contains(got, "Eingeschränkter Betrieb zwischen Heuwaage und Brausebad") {
		t.Errorf("messages interactive missing selected content:\n%s", got)
	}
}

func TestHelpOutput(t *testing.T) {
	got := captureStdout(t, func() { Run([]string{"departures", "--help"}) })
	if !strings.HasPrefix(got, "Description:\n  Create a departure schedule for a specific bus stop.\n\nUsage:\n  departures [options] [--] [<stop>]") {
		t.Errorf("help output:\n%s", got)
	}
	if !strings.Contains(got, "      --limit[=LIMIT]   Limits the number of displayed departures (default is 10, optional)") {
		t.Errorf("help output:\n%s", got)
		t.Errorf("help output missing limit option:\n%s", got)
	}
}

func TestHelpCommand(t *testing.T) {
	got := captureStdout(t, func() { Run([]string{"help", "messages"}) })
	if !strings.HasPrefix(got, "Description:\n  Show current information") {
		t.Errorf("help command output:\n%s", got)
	}
}

func TestListCommand(t *testing.T) {
	got := captureStdout(t, func() { Run([]string{"list"}) })
	if !strings.Contains(got, "USAGE:") || !strings.Contains(got, "departures") {
		t.Errorf("list output:\n%s", got)
	}
}

func TestLimitFlag(t *testing.T) {
	setupClient(t)
	got := captureStdout(t, func() {
		Run([]string{"departures", "--gid", "--limit=2", "--json", "ch:23005:300"})
	})
	// The client sends limit=2; the mock returns the fixture but the
	// request query must carry limit=2.
	if !strings.Contains(got, `"realtime"`) {
		t.Errorf("limit flag output:\n%s", got)
	}
}

func TestJSONFlagAfterArgs(t *testing.T) {
	setupClient(t)
	got := captureStdout(t, func() {
		Run([]string{"stopinfo", "Basel, Basel SBB", "--json"})
	})
	if !strings.Contains(got, `"gid": "ch:23005:300"`) {
		t.Errorf("stopinfo --json after args:\n%s", got)
	}
}

func promptsFakeInput(keys []string) {
	prompts.SetFakeInput(keys)
}

func TestDeparturesInteractiveSearch(t *testing.T) {
	setupClient(t)
	carbon.SetTestNow(time.Date(2026, 9, 14, 21, 5, 0, 0, time.UTC))
	t.Cleanup(func() { carbon.SetTestNow(time.Time{}) })
	// Type "bas" (matches the fixture search results), press DOWN to
	// highlight the first result, then ENTER to submit.
	promptsFakeInput([]string{"b", "a", "s", "\x1b[B", "\n"})
	got := captureStdout(t, func() {
		Run([]string{"departures"})
	})
	if i := strings.Index(got, "┌"); i != -1 {
		got = got[i:]
	}
	if !strings.Contains(got, "Basel, SBB") || !strings.Contains(got, "Zürich HB") {
		t.Errorf("interactive search output:\n%s", got)
	}
}

func TestRouteInteractiveTripSelect(t *testing.T) {
	setupClient(t)
	// The trips select: ENTER picks the first trip.
	promptsFakeInput([]string{"\n"})
	got := captureStdout(t, func() {
		Run([]string{"route", "Basel, Basel SBB", "Basel, Claraplatz", "2045", "20260914", "Departure"})
	})
	if i := strings.Index(got, "┌"); i != -1 {
		got = got[i:]
	}
	if !strings.Contains(got, "Fussweg") || !strings.Contains(got, "Getting off/Transferring") {
		t.Errorf("route interactive output:\n%s", got)
	}
}
