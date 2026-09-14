package split

import "testing"

func TestWordWrapGoldenVectors(t *testing.T) {
	cases := []struct {
		in, want string
		width    int
	}{
		{"aaaa bbbbb", "aaaa\nbbbbb", 6},
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 10},
		{" aaaa bbbbb", " aaaa\nbbbbb", 6},
		{"aaaa  bbbbb", "aaaa \nbbbbb", 6},
		{"ab\ncd", "ab\ncd", 2},
		{"hello world foo", "hello\nworld foo", 10},
		{"aaaaa bbbbb ccccc", "aaaaa\nbbbbb\nccccc", 5},
		{"", "", 5},
		{"  abcde  abcde", " \nabcde\n abcde", 5},
		{"x y z", "x\ny\nz", 1},
		{"aaaa bbbbb ccc", "aaaa\nbbbbb\nccc", 8},
		{"one two three four five", "one two\nthree\nfour five", 9},
		{"Über die Haltestellen Basel, Bahnhof SBB", "Über die Haltestellen Basel, Bahnhof\nSBB", 40},
	}
	for _, c := range cases {
		got := WordWrap(c.in, c.width, "\n")
		if got != c.want {
			t.Errorf("WordWrap(%q, %d) = %q, want %q", c.in, c.width, got, c.want)
		}
	}
}

func TestText(t *testing.T) {
	in := "line one\nline two longer than seventy four characters is quite long indeed yes"
	got := Text(in)
	want := "line one\n\nline two longer than seventy four characters is quite long indeed yes"
	if got != want {
		t.Errorf("Text() = %q, want %q", got, want)
	}
	// Wrapping at 74 characters.
	in = "aaaa\n" + "b" + "\n" + "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa bbbbb ccccc ddddd"
	got = Text(in)
	want = "aaaa\n\nb\n\naaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\nbbbbb ccccc ddddd"
	if got != want {
		t.Errorf("Text() wrapped = %q, want %q", got, want)
	}
	// Existing newlines are doubled.
	if Text("a\nb") != "a\n\nb" {
		t.Error("newline doubling failed")
	}
	if Text("") != "" {
		t.Error("empty input should stay empty")
	}
}
