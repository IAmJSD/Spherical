package defaulttrusted

import (
	_ "embed"
	"strings"
)

//go:embed defaults.txt
var defaults string

// New returns a new slice with defaults contents normalised.
func New() []string {
	s := strings.Split(defaults, "\n")
	a := make([]string, 0, len(s))
	for _, v := range s {
		v = strings.TrimSpace(v)
		if v != "" && !strings.HasPrefix(v, "#") {
			a = append(a, v)
		}
	}
	return a
}
