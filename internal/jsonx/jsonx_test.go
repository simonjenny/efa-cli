package jsonx

import (
	"encoding/json"
	"math"
	"testing"
)

func TestEncodePHPObjectKeyOrder(t *testing.T) {
	o := NewOrdered()
	o.Set("b", float64(1))
	o.Set("a", "Zürich")
	got := EncodePHP(o)
	want := `{
    "b": 1,
    "a": "Z\u00fcrich"
}`
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestEncodePHPEmpty(t *testing.T) {
	if got := EncodePHP(NewOrdered()); got != "{}" {
		t.Errorf("empty object: got %q", got)
	}
	if got := EncodePHP([]any{}); got != "[]" {
		t.Errorf("empty array: got %q", got)
	}
	if got := EncodePHP(nil); got != "null" {
		t.Errorf("null: got %q", got)
	}
	if got := EncodePHP(false); got != "false" {
		t.Errorf("false: got %q", got)
	}
}

func TestEncodePHPNumbers(t *testing.T) {
	cases := []struct {
		in   json.Number
		want string
	}{
		{"5968115.0", "5968115"},
		{"113.775", "113.775"},
		{"7.50", "7.5"},
		{"123", "123"},
		{"-5", "-5"},
		{"1e5", "100000"},
		{"1e21", "1.0e+21"},
		{"1.5e-5", "1.5e-5"},
		{"0.0", "0"},
		{"9223372036854775807", "9223372036854775807"},
	}
	for _, c := range cases {
		got := EncodePHP(c.in)
		if got != c.want {
			t.Errorf("EncodePHP(%s) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestEncodePHPFloats(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{2.0, "2"},
		{0.5, "0.5"},
		{math.Copysign(0, -1), "-0"},
		{5968115.0, "5968115"},
		{7.8333333333333, "7.8333333333333"},
		{-12.166666666667, "-12.166666666667"},
		{1e21, "1.0e+21"},
	}
	for _, c := range cases {
		got := EncodePHP(c.in)
		if got != c.want {
			t.Errorf("EncodePHP(%v) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestEscapePHPString(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{`Zürich`, `Z\u00fcrich`},
		{`🚌`, `\ud83d\ude8c`},
		{`<b>&`, `<b>&`},
		{"a\nb", `a\nb`},
		{"\t", `\t`},
		{`quote"back\`, `quote\"back\\`},
		{"\x01", `\u0001`},
		{"ascii", "ascii"},
	}
	for _, c := range cases {
		got := escapePHPString(c.in)
		if got != c.want {
			t.Errorf("escapePHPString(%q) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestDecode(t *testing.T) {
	data := []byte(`{"a": 1, "b": [true, null, "x"], "c": {"d": 1.5}}`)
	v, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	o := v.(*Ordered)
	if len(o.keys) != 3 || o.keys[0] != "a" || o.keys[1] != "b" || o.keys[2] != "c" {
		t.Fatalf("key order: %v", o.keys)
	}
	if num := o.values["a"].(json.Number); num.String() != "1" {
		t.Errorf("a = %v", num)
	}
	arr := o.values["b"].([]any)
	if len(arr) != 3 || arr[0] != true || arr[1] != nil || arr[2] != "x" {
		t.Errorf("b = %v", arr)
	}
	inner := o.values["c"].(*Ordered)
	if num := inner.values["d"].(json.Number); num.String() != "1.5" {
		t.Errorf("c.d = %v", num)
	}
}

func TestDecodeInvalid(t *testing.T) {
	for _, bad := range []string{"", "not json", "{", "[]x"} {
		if v, err := Decode([]byte(bad)); err == nil && v != nil {
			t.Errorf("Decode(%q) should fail, got %v", bad, v)
		}
	}
}

func TestPath(t *testing.T) {
	data := []byte(`{"stopFinder": {"points": {"point": {"name": "Basel, SBB"}}}}`)
	v, _ := Decode(data)
	pt := Path(v, "stopFinder", "points", "point")
	o := pt.(*Ordered)
	if name, _ := o.Get("name"); name != "Basel, SBB" {
		t.Errorf("name = %v", name)
	}
	if Path(v, "missing") != nil {
		t.Error("missing path should be nil")
	}
	if Path(v, "stopFinder", "points", "point", "ref", "gid") != nil {
		t.Error("missing nested path should be nil")
	}
}

func TestDecodeEmptyObjectIsArray(t *testing.T) {
	// PHP json_decode('[]') is an empty array; json_encode([]) -> "[]".
	v, err := Decode([]byte("[]"))
	if err != nil {
		t.Fatal(err)
	}
	arr := v.([]any)
	if arr == nil || len(arr) != 0 {
		t.Errorf("expected empty non-nil array")
	}
	if EncodePHP(arr) != "[]" {
		t.Errorf("empty array should encode to []")
	}
}

func TestNestedPrettyPrint(t *testing.T) {
	data := []byte(`{"a": [{"b": 1}, {"c": 2}]}`)
	v, _ := Decode(data)
	got := EncodePHP(v)
	want := `{
    "a": [
        {
            "b": 1
        },
        {
            "c": 2
        }
    ]
}`
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}
