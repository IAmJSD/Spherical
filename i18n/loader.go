package i18n

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/jakemakesstuff/spherical/config"
	"github.com/robfig/gettext-go/gettext/po"
)

//go:embed locales/**/*
var locales embed.FS

var allowedFsChars = regexp.MustCompile("^[a-zA-Z0-9_]+$")

var (
	cache     = map[string]*po.File{}
	cacheLock = sync.RWMutex{}
)

var localsOverride fs.FS

// TODO: dev mode
func cachedPoGet(fp string) *po.File {
	if localsOverride == nil {
		// Check the cache with the filename key.
		cacheLock.RLock()
		poCache := cache[fp]
		cacheLock.RUnlock()
		if poCache != nil {
			return poCache
		}
	}

	// Read the file.
	var useFs fs.FS
	if localsOverride == nil {
		useFs = locales
	} else {
		useFs = localsOverride
	}
	f, err := useFs.Open(fp)
	if err != nil {
		return nil
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil
	}

	// Check if the file is in the cache if this is dev.
	if localsOverride != nil {
		cacheLock.RLock()
		poCache := cache[string(b)]
		cacheLock.RUnlock()
		if poCache != nil {
			return poCache
		}
	}

	// Parse the PO file.
	p, err := po.LoadData(b)
	if err != nil {
		return nil
	}

	// Write to the cache.
	cacheLock.Lock()
	if localsOverride == nil {
		cache[fp] = p
	} else {
		cache[string(b)] = p
	}
	cacheLock.Unlock()

	// Return the PO file.
	return p
}

// Validates the locale. If this fails, returns the system default.
// If this is blank, returns en_US.
func validateLocale(locale string) string {
	// Check it is <something>_<caps something>. We can be usefully lax here,
	// but we have to be safe.
	s := strings.SplitN(locale, "_", 2)
	if len(s) != 2 {
		// Try a dash. Technically uncompliant to our needs, but fine enough.
		s = strings.SplitN(locale, "-", 2)
		if len(s) != 2 {
			// Fuck.
			s = nil
		}
	}

	// Check if the parts are okay.
	if len(s) == 2 {
		if allowedFsChars.MatchString(s[0]) && allowedFsChars.MatchString(s[1]) {
			return strings.ToLower(s[0]) + "_" + strings.ToUpper(s[1])
		}
	}

	// Try to figure out what locale to use from the system.
	configLocale := config.Config().Locale
	if configLocale == "" {
		// Default to en_US if everything else gets fucked.
		configLocale = "en_US"
	}
	return configLocale
}

func parseLocaleString(s string, preferredLocale string) string {
	// Get the locale.
	preferredLocale = validateLocale(preferredLocale)
	localesChecked := []string{preferredLocale}
	if preferredLocale != "en_US" {
		localesChecked = append(localesChecked, preferredLocale)
	}

	// Split the string by the colon.
	split := strings.SplitN(s, ":", 2)
	if len(split) != 2 {
		// Locale string is badly formatted. Assume it is not one.
		return s
	}

	// Split by the slash into parts and make sure that each part is valid.
	validatedParts := []string{}
	for _, v := range strings.Split(split[0], "/") {
		if v == "" {
			continue
		}
		if !allowedFsChars.MatchString(v) {
			return "!! INVALID LOCALE PATH - UNSUPPORTED FS CHARS !!"
		}
		validatedParts = append(validatedParts, v)
	}

	// Go through each locale and check.
	for _, locale := range localesChecked {
		// Read the file.
		safePath := "locales/" + locale + "/" +
			strings.Join(validatedParts, "/") + ".po"
		poFile := cachedPoGet(safePath)

		// Read the string.
		var value string
		for _, v := range poFile.Messages {
			if v.MsgId == split[1] {
				// This is the right one.
				value = v.MsgStr
				break
			}
		}
		if value != "" {
			return value
		}
	}
	return split[1]
}

var once uintptr

// Setup is used to initialize the i18n engine. Should only be called once.
func Setup(dev bool) {
	if dev {
		s, err := os.Stat("i18n")
		if err == nil && s.IsDir() {
			if atomic.SwapUintptr(&once, 1) == 0 {
				fmt.Println("[i18n] dev mode enabled - using local fs for locales folder!")
			}
			localsOverride = os.DirFS("i18n")
			return
		}
		if atomic.SwapUintptr(&once, 1) == 0 {
			fmt.Println("[i18n] i18n folder is not in cwd - live updates are off!")
		}
	}
}
