package htmltext

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoldenVectors(t *testing.T) {
	// Golden files generated with the real stevebauman/hypertext v1.1.2
	// (keepNewLines) on live EFA content.
	files, err := filepath.Glob("../../testdata/hypertext/content_*_text.txt")
	if err != nil || len(files) == 0 {
		t.Skip("golden files not available")
	}
	for _, f := range files {
		id := filepath.Base(f)
		id = id[len("content_") : len(id)-len("_text.txt")]
		in, err := os.ReadFile(filepath.Join("../../testdata/hypertext", "content_"+id+".html"))
		if err != nil {
			t.Fatalf("read input %s: %v", id, err)
		}
		want, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read golden %s: %v", id, err)
		}
		got := ToText(string(in))
		if got != string(want) {
			t.Errorf("ToText(%s) differs:\n got: %q\nwant: %q", id, got, string(want))
		}
	}
}

func TestBasic(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Hello", "Hello"},
		{"<p>Hello</p>", "Hello"},
		{"a<br>b", "ab"},
		{"a<br />b", "ab"},
		{"a&uuml;b", "aüb"},
		{"a&nbsp;b", "a b"},
		{"a <b>bold</b> c", "a bold c"},
		{"<div>x</div><div>y</div>", "x y"},
		{"a\n\nb", "a\nb"},
		{"a \n b", "a\nb"},
		{"a&#65;b", "aAb"},
		{"a&#x41;b", "aAb"},
		{"a<br>\n\nb", "a\nb"},
		{"a=20b", "a b"},
		{"a=\nb", "ab"},
		{"  padded  ", "padded"},
	}
	for _, c := range cases {
		got := ToText(c.in)
		if got != c.want {
			t.Errorf("ToText(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
