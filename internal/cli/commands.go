package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/simonjenny/efa-cli/internal/carbon"
	"github.com/simonjenny/efa-cli/internal/efa"
	"github.com/simonjenny/efa-cli/internal/htmltext"
	"github.com/simonjenny/efa-cli/internal/jsonx"
	"github.com/simonjenny/efa-cli/internal/prompts"
	"github.com/simonjenny/efa-cli/internal/split"
)

// newClient creates the EFA client.
var newClient = func() *efa.Client {
	return efa.NewClient()
}

// promptSearch runs a stop search prompt (Laravel Prompts search).
func promptSearch(client *efa.Client, label, placeholder string) (string, error) {
	client.UseSpinner = false
	value, err := prompts.Search(label, placeholder, func(term string) []prompts.SelectOption {
		if term == "" {
			return nil
		}
		options, err := client.Haltestellen("%" + term + "%")
		if err != nil {
			return nil
		}
		var out []prompts.SelectOption
		for _, o := range options {
			out = append(out, prompts.SelectOption{Key: o.Key, Label: o.Label})
		}
		return out
	}, "")
	if err != nil {
		return "", err
	}
	return value, nil
}

// handlePromptError prints the error like Symfony renders a
// NonInteractiveValidationException.
func handlePromptError(err error) int {
	if _, ok := err.(*prompts.NonInteractiveValidationError); ok {
		out("Required.\n")
		return 1
	}
	out(err.Error() + "\n")
	return 1
}

// ---------------------------------------------------------------------------
// departures
// ---------------------------------------------------------------------------

func runDepartures(p *parsedArgs) int {
	client := newClient()

	stop := p.arg("stop")
	if stop == "" {
		value, err := promptSearch(client, "Witch Stop do you want to see the departures for?", "Basel, Claraplatz")
		if err != nil {
			return handlePromptError(err)
		}
		stop = value
	}

	jsonMode := p.has("json")
	limit := "10"
	if p.has("limit") {
		limit = p.value("limit")
	}

	client.UseSpinner = !jsonMode
	departures, err := client.Abfahrt(stop, limit, p.has("gid"))
	if err != nil {
		out("Fetching Data... failed\n")
		return 1
	}

	if jsonMode {
		for _, d := range departures {
			obj, ok := d.(*jsonx.Ordered)
			if !ok {
				continue
			}
			t := departureTime(obj)
			realtime := carbon.FloorMinutes(carbon.DiffInMinutes(carbon.Now(), t))
			obj.Set("realtime", parseFloat(realtime))
		}
		out(jsonx.EncodePHP(departures) + "\n")
		return 0
	}

	var rows [][]string
	for _, d := range departures {
		obj, ok := d.(*jsonx.Ordered)
		if !ok {
			continue
		}
		product := stringOf(jsonx.Path(obj, "transportation", "product", "name"))
		typ := "🚊"
		if product == "Bus" {
			typ = "🚌"
		}
		number := stringOf(jsonx.Path(obj, "transportation", "number"))
		destination := stringOf(jsonx.Path(obj, "transportation", "destination", "name"))
		dep := departureTime(obj)
		rows = append(rows, []string{
			typ,
			number,
			destination,
			carbon.DiffForHumans(dep, carbon.Now()),
		})
	}
	prompts.DisplayTable([]string{"", "Nr.", "Destination", "Departure"}, rows)
	return 0
}

// departureTime resolves the effective departure time
// (departureTimeEstimated ?? departureTimePlanned).
func departureTime(obj *jsonx.Ordered) time.Time {
	t := stringOf(jsonx.Path(obj, "departureTimeEstimated"))
	if t == "" {
		t = stringOf(jsonx.Path(obj, "departureTimePlanned"))
	}
	parsed, err := carbon.Parse(t)
	if err != nil {
		return time.Now()
	}
	return parsed
}

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

func stringOf(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// ---------------------------------------------------------------------------
// messages
// ---------------------------------------------------------------------------

func runMessages(p *parsedArgs) int {
	return messages(p.has("json"))
}

func messages(jsonMode bool) int {
	client := newClient()
	client.UseSpinner = !jsonMode
	meldungen, err := client.Meldungen()
	if err != nil {
		out("Fetching Data... failed\n")
		return 1
	}

	if jsonMode {
		out(jsonx.EncodePHP(meldungen) + "\n")
		return 0
	}

	if len(meldungen) == 0 {
		return 0
	}

	options := make([]prompts.SelectOption, 0, len(meldungen))
	for i, m := range meldungen {
		obj, ok := m.(*jsonx.Ordered)
		if !ok {
			continue
		}
		text := stringOf(jsonx.Path(obj, "infoLink", "infoLinkText"))
		options = append(options, prompts.SelectOption{Key: "_" + strconv.Itoa(i), Label: text})
	}

	client.UseSpinner = false
	selected, err := prompts.Select("Which alert would you like to read?", options, 0, "")
	if err != nil {
		return handlePromptError(err)
	}
	client.UseSpinner = true

	id, err := strconv.Atoi(strings.TrimPrefix(selected, "_"))
	if err != nil {
		return 1
	}
	if id >= len(meldungen) {
		return 1
	}
	obj, ok := meldungen[id].(*jsonx.Ordered)
	if !ok {
		return 1
	}

	lastObj := lastMeldung(meldungen)
	header := stringOf(jsonx.Path(lastObj, "infoLink", "infoLinkText"))

	content := stringOf(jsonx.Path(obj, "infoLink", "content"))
	text := htmltext.ToText(content)
	wrapped := split.Text(text)

	prompts.DisplayTable([]string{header}, [][]string{{wrapped}})
	return 0
}

func lastMeldung(meldungen []any) *jsonx.Ordered {
	for i := len(meldungen) - 1; i >= 0; i-- {
		if o, ok := meldungen[i].(*jsonx.Ordered); ok {
			return o
		}
	}
	return jsonx.NewOrdered()
}

// ---------------------------------------------------------------------------
// stopinfo
// ---------------------------------------------------------------------------

func runStopinfo(p *parsedArgs) int {
	client := newClient()

	stop := p.arg("stop")
	if stop == "" {
		value, err := promptSearch(client, "Witch stop do you want to show information for?", "Basel, Basel SBB")
		if err != nil {
			return handlePromptError(err)
		}
		stop = value
	}

	jsonMode := p.has("json")
	client.UseSpinner = !jsonMode
	haltestelle, err := client.Haltestelle(stop)
	if err != nil {
		out("Fetching Data... failed\n")
		return 1
	}

	if jsonMode {
		if haltestelle == nil {
			out(jsonx.EncodePHP(false) + "\n")
		} else {
			out(jsonx.EncodePHP(haltestelle) + "\n")
		}
		return 0
	}

	name := ""
	id := ""
	gid := ""
	coords := ""
	infos := any(nil)
	if haltestelle != nil {
		name = stringOf(jsonx.Path(haltestelle, "name"))
		id = stringOf(jsonx.Path(haltestelle, "ref", "id"))
		gid = stringOf(jsonx.Path(haltestelle, "ref", "gid"))
		coords = stringOf(jsonx.Path(haltestelle, "ref", "coords"))
		infos, _ = haltestelle.Get("infos")
	}

	geo := strings.Split(coords, ",")
	geo0, geo1 := "", ""
	if len(geo) > 0 {
		geo0 = geo[0]
	}
	if len(geo) > 1 {
		geo1 = geo[1]
	}

	rows := [][]string{
		{"EFA Stop ID", id},
		{"GID", gid},
		{"Coordinates", coords},
		{"Google Maps", "https://www.google.com/maps/search/?api=1&query=" + geo1 + "," + geo0},
		{"Web Departure Monitor", "https://dfi.bvb.ch/?point=" + id},
		{"", ""},
	}

	if infos != nil {
		var text strings.Builder
		switch list := infos.(type) {
		case []any:
			for _, info := range list {
				if o, ok := info.(*jsonx.Ordered); ok {
					text.WriteString(stringOf(jsonx.Path(o, "infoLinkText")))
					text.WriteString("\n")
				}
			}
		case *jsonx.Ordered:
			text.WriteString(stringOf(jsonx.Path(list, "infoLinkText")))
			text.WriteString("\n")
		}
		text.WriteString("\n")
		text.WriteString("Detailed information available at http://info.bvb.ch (German only)")
		rows = append(rows, []string{"Info", text.String()})
	}

	prompts.DisplayTable([]string{name, ""}, rows)
	return 0
}

// ---------------------------------------------------------------------------
// route
// ---------------------------------------------------------------------------

func runRoute(p *parsedArgs) int {
	client := newClient()

	now := carbon.Now()

	start := p.arg("start")
	if start == "" {
		value, err := promptSearch(client, "Starting Stop:", "Basel, Basel SBB")
		if err != nil {
			return handlePromptError(err)
		}
		start = value
	}

	destination := p.arg("destination")
	if destination == "" {
		value, err := promptSearch(client, "Destination Stop:", "Basel, Claraplatz")
		if err != nil {
			return handlePromptError(err)
		}
		destination = value
	}

	timeValue := p.arg("time")
	if timeValue == "" {
		client.UseSpinner = false
		value, err := prompts.Text("At wich time would you like to departure/arrive ?", "", carbon.FormatTimeHi(now), "")
		if err != nil {
			return handlePromptError(err)
		}
		timeValue = value
	}

	dateValue := p.arg("date")
	if dateValue == "" {
		client.UseSpinner = false
		value, err := prompts.Text("At wich date would you like to departure/arrive ?", "", carbon.FormatDateDMY(now), "")
		if err != nil {
			return handlePromptError(err)
		}
		dateValue = value
	}

	mode := p.arg("mode")
	if mode == "" {
		client.UseSpinner = false
		value, err := prompts.Select("Do you want the trips for", []prompts.SelectOption{
			{Key: "Departure", Label: "Departure"},
			{Key: "Arrival", Label: "Arrival"},
		}, 0, "")
		if err != nil {
			return handlePromptError(err)
		}
		mode = value
	}

	jsonMode := p.has("json")
	client.UseSpinner = !jsonMode
	routes, err := client.Route(efa.RouteArgs{
		Date:        dateValue,
		Time:        timeValue,
		Start:       start,
		Destination: destination,
		Mode:        mode,
	})
	if err != nil {
		out("Fetching Data... failed\n")
		return 1
	}

	if jsonMode {
		out(jsonx.EncodePHP(routes) + "\n")
		return 0
	}

	trips := any(nil)
	if routesObj, ok := routes.(*jsonx.Ordered); ok {
		trips, _ = routesObj.Get("trips")
	}
	tripsList, ok := trips.([]any)
	if !ok || len(tripsList) == 0 {
		prompts.Error("There are no trips available!")
		return 0
	}

	tripOptions := make([]prompts.SelectOption, 0, len(tripsList))
	for i, t := range tripsList {
		obj, ok := t.(*jsonx.Ordered)
		if !ok {
			continue
		}
		legs, _ := obj.Get("legs")
		legsList, ok := legs.([]any)
		if !ok || len(legsList) == 0 {
			continue
		}
		firstLeg, ok := legsList[0].(*jsonx.Ordered)
		if !ok {
			continue
		}
		from := stringOf(jsonx.Path(firstLeg, "points", "0", "name"))
		at := stringOf(jsonx.Path(firstLeg, "points", "0", "dateTime", "time"))
		interchange := stringOf(jsonx.Path(obj, "interchange"))
		duration := stringOf(jsonx.Path(obj, "duration"))
		label := fmt.Sprintf("Departing from %s at %s with %s transfer(s). Duration: %sh", from, at, interchange, duration)
		tripOptions = append(tripOptions, prompts.SelectOption{Key: "_" + strconv.Itoa(i), Label: label})
	}

	client.UseSpinner = false
	selected, err := prompts.Select("Available Trips:", tripOptions, 0, "")
	if err != nil {
		return handlePromptError(err)
	}
	client.UseSpinner = true

	id, err := strconv.Atoi(strings.TrimPrefix(selected, "_"))
	if err != nil || id >= len(tripsList) {
		return 1
	}
	trip, ok := tripsList[id].(*jsonx.Ordered)
	if !ok {
		return 1
	}

	legs, _ := trip.Get("legs")
	legsList, ok := legs.([]any)
	if !ok {
		return 0
	}

	var rows [][]string
	meldungen := false
	for _, leg := range legsList {
		legObj, ok := leg.(*jsonx.Ordered)
		if !ok {
			continue
		}
		infos := jsonx.Path(legObj, "infos", "info")
		meldungen = infos != nil

		modeProduct := stringOf(jsonx.Path(legObj, "mode", "product"))
		modeNumber := stringOf(jsonx.Path(legObj, "mode", "number"))
		board := stringOf(jsonx.Path(legObj, "points", "0", "name"))
		depart := stringOf(jsonx.Path(legObj, "points", "0", "dateTime", "time"))
		arrive := stringOf(jsonx.Path(legObj, "points", "1", "dateTime", "time"))
		getOff := stringOf(jsonx.Path(legObj, "points", "1", "name"))

		productWord := strings.Fields(modeProduct)
		product := ""
		if len(productWord) > 0 {
			product = productWord[0]
		}

		duration := ""
		if depart != "" && arrive != "" {
			t0, err0 := carbon.Parse(depart)
			t1, err1 := carbon.Parse(arrive)
			if err0 == nil && err1 == nil {
				duration = carbon.FloatString(carbon.DiffInMinutes(t0, t1)) + "'"
			}
		}

		rows = append(rows, []string{
			product,
			modeNumber,
			board,
			depart + "h",
			arrive + "h",
			duration,
			getOff,
		})
	}

	prompts.DisplayTable([]string{"", "Nr.", "Boarding", "Departure", "Arrival", "Duration", "Getting off/Transferring"}, rows)

	if meldungen {
		confirmed, err := prompts.Confirm("there are alerts for this route? Would you like to see them now?", true)
		if err != nil {
			return handlePromptError(err)
		}
		if confirmed {
			return messages(false)
		}
	}

	return 0
}
