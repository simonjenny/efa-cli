package efa

import "testing"

func TestGuzzleQueryEncode(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Basler Verkehrs-Betriebe (BVB)", "Basler%20Verkehrs-Betriebe%20(BVB)"},
		{"Basel, Basel SBB", "Basel,%20Basel%20SBB"},
		{"%Basel%20SBB%", "%Basel%20SBB%25"},
		{"WGS84[DD.ddddd]", "WGS84%5BDD.ddddd%5D"},
		{"a=b&c/d:e?f", "a=b&c/d:e?f"},
		{"Zürich", "Z%C3%BCrich"},
		{"Basel, Claraplatz", "Basel,%20Claraplatz"},
		{"Departure", "Departure"},
		{"20260914", "20260914"},
		{"20:45", "20:45"},
		{"%xy", "%25xy"},
		{"a b%20c", "a%20b%20c"},
	}
	for _, c := range cases {
		got := guzzleQueryEncode(c.in)
		if got != c.want {
			t.Errorf("guzzleQueryEncode(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBuildURL(t *testing.T) {
	cases := []struct {
		endpoint string
		params   []queryParam
		want     string
	}{
		{
			"XML_ADDINFO_REQUEST",
			[]queryParam{
				{key: "filterProviderCode", value: "Basler Verkehrs-Betriebe (BVB)"},
				{key: "outputFormat", value: "JSON"},
			},
			"https://www.efa-bw.de/bvb3/XML_ADDINFO_REQUEST?filterProviderCode=Basler%20Verkehrs-Betriebe%20(BVB)&outputFormat=JSON",
		},
		{
			"XSLT_STOPFINDER_REQUEST",
			[]queryParam{
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
				{key: "name_sf", value: "Basel, Basel SBB"},
			},
			"https://www.efa-bw.de/bvb3/XSLT_STOPFINDER_REQUEST?language=de&outputFormat=JSON&coordOutputFormat=WGS84%5BDD.ddddd%5D&itdLPxx_usage=origin&useLocalityMainStop=true&SpEncId=0&locationServerActive=1&stateless=1&type_sf=any&anyObjFilter_sf=2&anyMaxSizeHitList=1&name_sf=Basel,%20Basel%20SBB",
		},
		{
			"XML_DM_REQUEST",
			[]queryParam{
				{key: "laguage", value: "de"},
				{key: "typeInfo_dm", value: "stopID"},
				{key: "deleteAssignedStops_dm", value: "1"},
				{key: "useRealtime", value: "1"},
				{key: "mode", value: "direct"},
				{key: "outputFormat", value: "rapidJSON"},
				{key: "limit", value: "10"},
				{key: "nameInfo_dm", value: "ch:23005:300"},
			},
			"https://www.efa-bw.de/bvb3/XML_DM_REQUEST?laguage=de&typeInfo_dm=stopID&deleteAssignedStops_dm=1&useRealtime=1&mode=direct&outputFormat=rapidJSON&limit=10&nameInfo_dm=ch:23005:300",
		},
		{
			"XSLT_TRIP_REQUEST2",
			[]queryParam{
				{key: "itdDate", value: "20260914"},
				{key: "itdTime", value: "2045"},
				{key: "language", value: "de"},
				{key: "sessionID", value: "0"},
				{key: "outputFormat", value: "JSON"},
				{key: "type_origin", value: "stop"},
				{key: "name_origin", value: "Basel, Basel SBB"},
				{key: "type_destination", value: "stop"},
				{key: "name_destination", value: "Basel, Claraplatz"},
				{key: "itdTripDateTimeDepArr", value: "Departure"},
			},
			"https://www.efa-bw.de/bvb3/XSLT_TRIP_REQUEST2?itdDate=20260914&itdTime=2045&language=de&sessionID=0&outputFormat=JSON&type_origin=stop&name_origin=Basel,%20Basel%20SBB&type_destination=stop&name_destination=Basel,%20Claraplatz&itdTripDateTimeDepArr=Departure",
		},
	}
	for _, c := range cases {
		got := buildURL("https://www.efa-bw.de/bvb3/", c.endpoint, c.params)
		if got != c.want {
			t.Errorf("buildURL:\n got: %s\nwant: %s", got, c.want)
		}
	}
}
