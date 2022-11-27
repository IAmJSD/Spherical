package i18n

import (
	"container/list"
	"io/fs"
	"net/http"
	"strings"

	"github.com/jakemakesstuff/spherical/config"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

func getLocale(r *http.Request) string {
	var locale string
	for _, v := range r.Cookies() {
		if v.Name == "spherical_language_override" {
			locale = v.Value
			break
		}
	}
	if locale == "" {
		// If there is no language override, check if the config has the locale.
		locale = config.Config().Locale
		if locale == "" {
			// This is set to auto mode. Try to get from the browser.
			s := strings.SplitN(r.Header.Get("Accept-Language"), ",", 2)
			locale = s[0]
			if locale == "" {
				// Set to en_US. No language found.
				locale = "en_US"
			}
		}
	}
	return locale
}

// GetWithRequest is used to get an i18n string with the http request.
func GetWithRequest(r *http.Request, s string) string {
	locale := getLocale(r)
	return parseLocaleString(s, locale)
}

func recursivePos(f fs.FS, fp string, l *list.List) []string {
	start := l == nil
	if start {
		l = list.New()
	}
	file, err := f.Open(fp)
	if err == nil {
		s, err := file.Stat()
		if err == nil {
			if s.IsDir() {
				// List all files in this folder.
				dirs, _ := fs.ReadDir(f, fp)
				for _, v := range dirs {
					recursivePos(f, fp+"/"+v.Name(), l)
				}
			} else {
				// Tag this one along.
				l.PushBack(fp)
			}
		}
	}

	if start {
		s := make([]string, l.Len())
		i := 0
		for x := l.Front(); x != nil; x = x.Next() {
			s[i] = x.Value.(string)
			i++
		}
		return s
	}
	return nil
}

// GetAllFrontend is used to get all frontend PO strings.
func GetAllFrontend(r *http.Request) map[string]string {
	// Get the filesystem.
	var f fs.FS
	if localsOverride == nil {
		f = locales
	} else {
		f = localsOverride
	}

	// Go through each locale to get the keys mapped to values.
	langMapping := map[string]string{}
	locales := []string{validateLocale(getLocale(r))}
	if locales[0] != "en_US" {
		locales = append(locales, "en_US")
	}
	for _, locale := range locales {
		matches := recursivePos(f, "locales/"+locale+"/frontend", nil)
		for _, fp := range matches {
			// Load the PO file.
			poFile := cachedPoGet(fp)

			// Go through each phrase in there.
			for _, message := range poFile.Messages {
				frag := strings.SplitN(fp, "/", 3)[2]
				key := frag[:len(frag)-3] + ":" + message.MsgId
				if _, ok := langMapping[key]; !ok {
					langMapping[key] = message.MsgStr
				}
			}
		}
	}

	// Return all the phrases.
	return langMapping
}

// Locale is used to define a locale.
type Locale struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// GetLocales is used to get all the supported i18n locales and their names in their
// respective languages.
func GetLocales() []Locale {
	// Get the filesystem.
	var f fs.FS
	if localsOverride == nil {
		f = locales
	} else {
		f = localsOverride
	}

	// Look for the folders within locales.
	e, _ := fs.ReadDir(f, "locales")
	if e != nil {
		// Get all possible locale folders.
		k := make([]Locale, 0, len(e))
		for _, v := range e {
			if v.IsDir() {
				localeName := v.Name()
				t, err := language.Parse(localeName)
				if err == nil {
					k = append(k, Locale{
						Code: localeName,
						Name: display.Self.Name(t),
					})
				}
			}
		}
		return k
	}
	return []Locale{}
}
