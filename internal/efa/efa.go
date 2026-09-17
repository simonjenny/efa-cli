// Package efa replicates the App\Helpers\Efa class of the original
// efa-cli: the HTTP requests to the EFA journey planner API with
// Guzzle-compatible query encoding.
package efa

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/simonjenny/efa-cli/internal/i18n"
	"github.com/simonjenny/efa-cli/internal/jsonx"
	"github.com/simonjenny/efa-cli/internal/prompts"
)

const baseURL = "https://www.efa-bw.de/bvb3/"

// Client mirrors the App\Helpers\Efa class.
type Client struct {
	httpClient *http.Client
	// BaseURL overrides the default API base URL (used by tests).
	BaseURL string
	// UseSpinner controls whether requests show the "Fetching Data..."
	// spinner (mirrors the json config flag: no spinner in JSON mode).
	UseSpinner bool
}

// NewClient creates a client with the default HTTP settings.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		BaseURL:    baseURL,
	}
}

// guzzleQueryEncode percent-encodes a query value the way Guzzle 7.9
// (PSR-7 Uri::filterQuery) does: unreserved characters, reserved
// characters and valid percent sequences stay as-is; everything else is
// percent-encoded with uppercase hex digits.
func guzzleQueryEncode(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '%' {
			if i+2 < len(s) && isHex(s[i+1]) && isHex(s[i+2]) {
				b.WriteByte(c)
			} else {
				b.WriteString("%25")
			}
			continue
		}
		if isQueryAllowed(c) {
			b.WriteByte(c)
			continue
		}
		fmt.Fprintf(&b, "%%%02X", c)
	}
	return b.String()
}

func isQueryAllowed(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return true
	}
	switch c {
	case '-', '_', '.', '~', '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', '%', ':', '@', '/', '?':
		return true
	}
	return false
}

func isHex(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

type queryParam struct {
	key   string
	value string
}

// buildURL constructs the request URL with the exact query parameter
// order used by the PHP application.
func buildURL(base string, endpoint string, params []queryParam) string {
	var b strings.Builder
	b.WriteString(base)
	b.WriteString(endpoint)
	b.WriteByte('?')
	for i, p := range params {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(p.key)
		b.WriteByte('=')
		b.WriteString(guzzleQueryEncode(p.value))
	}
	return b.String()
}

// get performs the HTTP GET, decoding the JSON response. Invalid or
// empty responses decode to an empty array, matching PHP's
// response->object() ?? [].
func (c *Client) get(endpoint string, params []queryParam) (any, error) {
	url := buildURL(c.BaseURL, endpoint, params)
	if c.UseSpinner {
		return prompts.SpinResult(i18n.T("spinner.fetching"), func() (any, error) {
			return c.doGet(url)
		})
	}
	return c.doGet(url)
}

func (c *Client) doGet(url string) (any, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return []any{}, nil
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "GuzzleHttp/7")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return []any{}, nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []any{}, nil
	}
	v, err := jsonx.Decode(body)
	if err != nil {
		return []any{}, nil
	}
	return v, nil
}

// Option is a stop search result (GID + name).
type Option struct {
	Key   string
	Label string
}

// Meldungen returns the BVB travel information messages, mirroring
// Efa::meldungen().
func (c *Client) Meldungen() ([]any, error) {
	v, err := c.get("XML_ADDINFO_REQUEST", []queryParam{
		{key: "filterProviderCode", value: "Basler Verkehrs-Betriebe (BVB)"},
		{key: "outputFormat", value: "JSON"},
	})
	if err != nil {
		return nil, err
	}
	return toAnySlice(jsonx.Path(v, "additionalInformation", "travelInformations", "travelInformation")), nil
}

// Haltestelle resolves a single stop, mirroring Efa::haltestelle().
// It returns nil when no stop could be resolved (PHP: false).
func (c *Client) Haltestelle(name string) (*jsonx.Ordered, error) {
	v, err := c.get("XSLT_STOPFINDER_REQUEST", []queryParam{
		{key: "language", value: "de"},
		{key: "outputFormat", value: "JSON"},
		{key: "coordOutputFormat", value: "WGS84[DD.ddddd]"},
		{key: "itdLPxx_usage", value: "origin"},
		{key: "useLocalityMainStop", value: "true"},
		{key: "SpEncId", value: "0"},
		{key: "locationServerActive", value: "1"},
		{key: "stateless", value: "1"},
		{key: "type_sf", value: "any"},
		{key: "anyObjFilter_sf", value: "2"},
		{key: "anyMaxSizeHitList", value: "1"},
		{key: "name_sf", value: name},
	})
	if err != nil {
		return nil, err
	}
	point := jsonx.Path(v, "stopFinder", "points", "point")
	o, ok := point.(*jsonx.Ordered)
	if !ok {
		return nil, nil
	}
	return o, nil
}

// Haltestellen searches for stops, mirroring Efa::haltestellen().
func (c *Client) Haltestellen(search string) ([]Option, error) {
	v, err := c.get("XSLT_STOPFINDER_REQUEST", []queryParam{
		{key: "language", value: "de"},
		{key: "outputFormat", value: "JSON"},
		{key: "coordOutputFormat", value: "WGS84[DD.ddddd]"},
		{key: "itdLPxx_usage", value: "origin"},
		{key: "useLocalityMainStop", value: "true"},
		{key: "SpEncId", value: "0"},
		{key: "locationServerActive", value: "1"},
		{key: "stateless", value: "1"},
		{key: "type_sf", value: "any"},
		{key: "anyObjFilter_sf", value: "2"},
		{key: "anyMaxSizeHitList", value: "50"},
		{key: "name_sf", value: search},
	})
	if err != nil {
		return nil, err
	}
	var options []Option
	switch points := jsonx.Path(v, "stopFinder", "points").(type) {
	case []any:
		for _, p := range points {
			if point, ok := p.(*jsonx.Ordered); ok {
				ref := jsonx.Path(point, "ref")
				if refObj, ok := ref.(*jsonx.Ordered); ok {
					gid, _ := refObj.Get("gid")
					name, _ := point.Get("name")
					options = append(options, Option{
						Key:   stringOf(gid),
						Label: stringOf(name),
					})
				}
			}
		}
	case *jsonx.Ordered:
		ref := jsonx.Path(points, "point", "ref")
		if refObj, ok := ref.(*jsonx.Ordered); ok {
			gid, _ := refObj.Get("gid")
			name, _ := jsonx.Path(points, "point").(*jsonx.Ordered).Get("name")
			options = append(options, Option{
				Key:   stringOf(gid),
				Label: stringOf(name),
			})
		}
	}
	return options, nil
}

// RouteArgs mirrors the route command arguments.
type RouteArgs struct {
	Date        string
	Time        string
	Start       string
	Destination string
	Mode        string
}

// Route plans a trip, mirroring Efa::route().
func (c *Client) Route(args RouteArgs) (any, error) {
	return c.get("XSLT_TRIP_REQUEST2", []queryParam{
		{key: "itdDate", value: args.Date},
		{key: "itdTime", value: args.Time},
		{key: "language", value: "de"},
		{key: "sessionID", value: "0"},
		{key: "outputFormat", value: "JSON"},
		{key: "type_origin", value: "stop"},
		{key: "name_origin", value: args.Start},
		{key: "type_destination", value: "stop"},
		{key: "name_destination", value: args.Destination},
		{key: "itdTripDateTimeDepArr", value: depArrCode(args.Mode)},
	})
}

// depArrCode maps the user-facing "Departure"/"Arrival" mode to the short
// code the EFA API expects for itdTripDateTimeDepArr.
func depArrCode(mode string) string {
	if mode == "Arrival" {
		return "arr"
	}
	return "dep"
}

// Abfahrt returns the departures for a stop, mirroring Efa::abfahrt().
// The limit is passed through as a string, exactly like the PHP
// application passes the raw option value into the URL.
func (c *Client) Abfahrt(stop string, limit string, isGID bool) ([]any, error) {
	gid := stop
	if !isGID {
		haltestelle, err := c.Haltestelle(stop)
		if err != nil {
			return nil, err
		}
		if haltestelle != nil {
			ref := jsonx.Path(haltestelle, "ref")
			if refObj, ok := ref.(*jsonx.Ordered); ok {
				g, _ := refObj.Get("gid")
				gid = stringOf(g)
			} else {
				gid = ""
			}
		} else {
			gid = ""
		}
	}
	v, err := c.get("XML_DM_REQUEST", []queryParam{
		{key: "laguage", value: "de"},
		{key: "typeInfo_dm", value: "stopID"},
		{key: "deleteAssignedStops_dm", value: "1"},
		{key: "useRealtime", value: "1"},
		{key: "mode", value: "direct"},
		{key: "outputFormat", value: "rapidJSON"},
		{key: "limit", value: limit},
		{key: "nameInfo_dm", value: gid},
	})
	if err != nil {
		return nil, err
	}
	return toAnySlice(jsonx.Path(v, "stopEvents")), nil
}

func toAnySlice(v any) []any {
	if arr, ok := v.([]any); ok {
		return arr
	}
	return []any{}
}

func stringOf(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
