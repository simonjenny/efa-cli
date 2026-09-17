package i18n

import (
	"testing"
)

func TestSetLang(t *testing.T) {
	if err := SetLang("de"); err != nil {
		t.Fatalf("SetLang(de): %v", err)
	}
	if Lang() != "de" {
		t.Errorf("Lang() = %q, want de", Lang())
	}
	if err := SetLang("EN"); err != nil {
		t.Fatalf("SetLang(EN): %v", err)
	}
	if Lang() != "en" {
		t.Errorf("Lang() = %q, want en", Lang())
	}
	if err := SetLang("fr"); err == nil {
		t.Error("SetLang(fr) expected an error")
	}
	if Lang() != "en" {
		t.Errorf("Lang() = %q, want en after failed set", Lang())
	}
	_ = SetLang(DefaultLang)
}

func TestT(t *testing.T) {
	_ = SetLang("de")
	if got := T("summary.usage"); got != "VERWENDUNG:" {
		t.Errorf(`T("summary.usage") = %q`, got)
	}
	if got := T("route.trip_label", "Basel SBB", "20:45", "1", "2"); got != "Abfahrt ab Basel SBB um 20:45 mit 1 Umstieg(en). Dauer: 2h" {
		t.Errorf("T with args = %q", got)
	}
	_ = SetLang("en")
	if got := T("summary.usage"); got != "USAGE:" {
		t.Errorf(`T("summary.usage") en = %q`, got)
	}
	if got := T("missing.key"); got != "missing.key" {
		t.Errorf("missing key = %q, want the key itself", got)
	}
	_ = SetLang(DefaultLang)
}

func TestTranslationKeyParity(t *testing.T) {
	for k := range translationsDE {
		if _, ok := translationsEN[k]; !ok {
			t.Errorf("key %q exists in DE but not in EN", k)
		}
	}
	for k := range translationsEN {
		if _, ok := translationsDE[k]; !ok {
			t.Errorf("key %q exists in EN but not in DE", k)
		}
	}
}

func TestSupported(t *testing.T) {
	got := Supported()
	if len(got) != 2 || got[0] != "de" || got[1] != "en" {
		t.Errorf("Supported() = %v", got)
	}
}

func TestDetectEnvLang(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{name: "unset", env: map[string]string{}, want: "de"},
		{name: "german swiss", env: map[string]string{"LANG": "de_CH.UTF-8"}, want: "de"},
		{name: "german plain", env: map[string]string{"LANG": "de"}, want: "de"},
		{name: "german dash", env: map[string]string{"LANG": "de-CH"}, want: "de"},
		{name: "uppercase", env: map[string]string{"LANG": "DE_de"}, want: "de"},
		{name: "english", env: map[string]string{"LANG": "en_US.UTF-8"}, want: "en"},
		{name: "c locale", env: map[string]string{"LANG": "C"}, want: "de"},
		{name: "posix locale", env: map[string]string{"LANG": "POSIX"}, want: "de"},
		{name: "unsupported", env: map[string]string{"LANG": "fr_FR.UTF-8"}, want: "de"},
		{name: "lc_messages fallback", env: map[string]string{"LC_MESSAGES": "en_GB.UTF-8"}, want: "en"},
		{name: "precedence", env: map[string]string{"LANG": "en_US", "LC_ALL": "de_DE.UTF-8"}, want: "de"},
		{name: "language wins", env: map[string]string{"LANG": "de_DE", "LANGUAGE": "en"}, want: "en"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range []string{"LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG"} {
				t.Setenv(k, "")
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			if got := DetectEnvLang(); got != tt.want {
				t.Errorf("DetectEnvLang() = %q, want %q", got, tt.want)
			}
		})
	}
}
