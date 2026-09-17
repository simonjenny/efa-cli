// Package i18n provides lightweight translations for the user facing
// strings of efa-cli. German is the default language, English the second.
package i18n

import (
	"fmt"
	"os"
	"strings"
)

// DefaultLang is used when no locale can be detected and nothing is set.
const DefaultLang = "de"

// supported lists the languages with translations.
var supported = []string{"de", "en"}

var current = DefaultLang

// SetLang switches the current language. Unknown languages return an error
// and leave the current language unchanged.
func SetLang(lang string) error {
	lang = normalize(lang)
	for _, l := range supported {
		if l == lang {
			current = lang
			return nil
		}
	}
	return fmt.Errorf("language %q is not supported", lang)
}

// Lang returns the currently active language.
func Lang() string {
	return current
}

// Supported returns the languages with translations.
func Supported() []string {
	return append([]string{}, supported...)
}

// T translates key and formats the result with args. When the key is
// missing in the current language, the English translation is used as a
// fallback, and finally the key itself.
func T(key string, args ...any) string {
	text, ok := translations[current][key]
	if !ok {
		text = translations["en"][key]
	}
	if text == "" {
		return key
	}
	if len(args) == 0 {
		return text
	}
	return fmt.Sprintf(text, args...)
}

// DetectEnvLang reads the standard locale environment variables
// (LANGUAGE, LC_ALL, LC_MESSAGES, LANG) and maps them to a supported
// language, e.g. "de_CH.UTF-8" -> "de". Unset locales and the neutral
// "C"/"POSIX" locales fall back to the default language.
func DetectEnvLang() string {
	for _, name := range []string{"LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG"} {
		v := os.Getenv(name)
		if v == "" || v == "C" || v == "POSIX" {
			continue
		}
		lang := normalize(v)
		for _, l := range supported {
			if l == lang {
				return l
			}
		}
	}
	return DefaultLang
}

// normalize maps a locale like "de_CH.UTF-8", "de-CH" or "De" to the
// language code "de".
func normalize(locale string) string {
	lang := strings.ToLower(locale)
	if i := strings.IndexAny(lang, "_-"); i != -1 {
		lang = lang[:i]
	}
	if i := strings.IndexByte(lang, '.'); i != -1 {
		lang = lang[:i]
	}
	return strings.TrimSpace(lang)
}
