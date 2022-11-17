package i18n

import (
	"net/http"
	"strings"

	"github.com/jakemakesstuff/spherical/config"
)

// GetWithRequest is used to get an i18n string with the http request.
func GetWithRequest(r *http.Request, s string) string {
	// Get the locale string.
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

	// Call the internationalisation function.
	return parseLocaleString(s, locale)
}
