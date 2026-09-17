package cli

import (
	"context"
	"net"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/simonjenny/efa-cli/internal/carbon"
	"github.com/simonjenny/efa-cli/internal/efa"
	"github.com/simonjenny/efa-cli/internal/i18n"
	"github.com/simonjenny/efa-cli/internal/jsonx"
)

// runMcp starts an MCP server that exposes departures, messages, route and
// stopinfo as MCP tools over Streamable HTTP.
func runMcp(p *parsedArgs) int {
	ip := p.value("ip")
	if ip == "" {
		ip = "127.0.0.1"
	}
	port := p.value("port")
	if port == "" {
		port = "8090"
	}
	addr := net.JoinHostPort(ip, port)

	server := newMCPServer()
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{Stateless: true})

	outf("efa-cli MCP server listening on http://%s\n", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		out(err.Error() + "\n")
		return 1
	}
	return 0
}

func newMCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: AppName, Version: Version}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "departures",
		Description: "Create a departure schedule for a specific bus stop.",
	}, departuresTool)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "messages",
		Description: "Show current information, disruptions and alerts (BVB network, german).",
	}, messagesTool)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "route",
		Description: "Plan a trip from point A to point B.",
	}, routeTool)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "stopinfo",
		Description: "Show information for a stop.",
	}, stopinfoTool)
	return server
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}

func errorResult(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: message}}}
}

type departuresArgs struct {
	Stop  string `json:"stop" jsonschema:"stop name to show departures for"`
	Limit string `json:"limit,omitempty" jsonschema:"maximum number of departures to return (default 10)"`
	Gid   bool   `json:"gid,omitempty" jsonschema:"set to true if stop is already a GID"`
}

func departuresTool(_ context.Context, _ *mcp.CallToolRequest, args departuresArgs) (*mcp.CallToolResult, any, error) {
	if args.Stop == "" {
		return errorResult(i18n.T("error.required")), nil, nil
	}
	limit := "10"
	if args.Limit != "" {
		limit = args.Limit
	}

	client := newClient()
	client.UseSpinner = false
	departures, err := client.Abfahrt(args.Stop, limit, args.Gid)
	if err != nil {
		return errorResult(i18n.T("error.fetch_failed")), nil, nil
	}

	for _, d := range departures {
		obj, ok := d.(*jsonx.Ordered)
		if !ok {
			continue
		}
		t := departureTime(obj)
		realtime := carbon.FloorMinutes(carbon.DiffInMinutes(carbon.Now(), t))
		obj.Set("realtime", parseFloat(realtime))
	}
	return textResult(jsonx.EncodePHP(departures)), nil, nil
}

type messagesArgs struct{}

func messagesTool(_ context.Context, _ *mcp.CallToolRequest, _ messagesArgs) (*mcp.CallToolResult, any, error) {
	client := newClient()
	client.UseSpinner = false
	meldungen, err := client.Meldungen()
	if err != nil {
		return errorResult(i18n.T("error.fetch_failed")), nil, nil
	}
	return textResult(jsonx.EncodePHP(meldungen)), nil, nil
}

type stopinfoArgs struct {
	Stop string `json:"stop" jsonschema:"stop name to look up"`
}

func stopinfoTool(_ context.Context, _ *mcp.CallToolRequest, args stopinfoArgs) (*mcp.CallToolResult, any, error) {
	if args.Stop == "" {
		return errorResult(i18n.T("error.required")), nil, nil
	}
	client := newClient()
	client.UseSpinner = false
	haltestelle, err := client.Haltestelle(args.Stop)
	if err != nil {
		return errorResult(i18n.T("error.fetch_failed")), nil, nil
	}
	if haltestelle == nil {
		return textResult(jsonx.EncodePHP(false)), nil, nil
	}
	return textResult(jsonx.EncodePHP(haltestelle)), nil, nil
}

type routeArgs struct {
	Start       string `json:"start" jsonschema:"starting stop name"`
	Destination string `json:"destination" jsonschema:"destination stop name"`
	Time        string `json:"time,omitempty" jsonschema:"time in HH:MM format (default: now)"`
	Date        string `json:"date,omitempty" jsonschema:"date in DD.MM.YYYY format (default: today)"`
	Mode        string `json:"mode,omitempty" jsonschema:"Departure or Arrival (default: Departure)"`
}

func routeTool(_ context.Context, _ *mcp.CallToolRequest, args routeArgs) (*mcp.CallToolResult, any, error) {
	if args.Start == "" || args.Destination == "" {
		return errorResult(i18n.T("error.required")), nil, nil
	}

	now := carbon.Now()
	timeValue := args.Time
	if timeValue == "" {
		timeValue = carbon.FormatTimeHi(now)
	}
	dateValue := args.Date
	if dateValue == "" {
		dateValue = carbon.FormatDateDMY(now)
	}
	mode := args.Mode
	if mode == "" {
		mode = "Departure"
	}

	parsedDate, err := carbon.Parse(dateValue)
	if err != nil {
		return errorResult(i18n.T("error.fetch_failed")), nil, nil
	}
	parsedTime, err := carbon.Parse(timeValue)
	if err != nil {
		return errorResult(i18n.T("error.fetch_failed")), nil, nil
	}

	client := newClient()
	client.UseSpinner = false
	routes, err := client.Route(efa.RouteArgs{
		Date:        carbon.FormatDateYmd(parsedDate),
		Time:        carbon.FormatTimeHiNumeric(parsedTime),
		Start:       args.Start,
		Destination: args.Destination,
		Mode:        mode,
	})
	if err != nil {
		return errorResult(i18n.T("error.fetch_failed")), nil, nil
	}
	return textResult(jsonx.EncodePHP(routes)), nil, nil
}
