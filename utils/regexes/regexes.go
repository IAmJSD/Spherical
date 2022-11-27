package regexes

import "regexp"

var (
	// Username is the regex for a valid Spherical username.
	Username = regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z0-9_-]{3,31}$")

	// Password is the regex for a valid Spherical password.
	Password = regexp.MustCompile("^[^ ].{5,71}$")

	// Email is the regex for a valid Spherical e-mail.
	Email = regexp.MustCompile("^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$")
)
