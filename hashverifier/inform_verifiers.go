package hashverifier

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// InformVerifiers is used to inform verifiers that a server is willing to verify a hash.
func InformVerifiers(originHostname, originHash, originPgpSig string, hashVerifierHostnames []string) {
	// Normalise the origin hostname.
	originHostname = strings.ToLower(originHostname)

	// Defines the payload that will be sent to hash verification servers.
	payload := originHostname + "\n" + originHash + "\n" + originPgpSig

	for _, hostname := range hashVerifierHostnames {
		// Make the URL.
		u, err := url.Parse("https://example.com/verify")
		if err != nil {
			// Not a valid hostname. Carry on.
			continue
		}
		u.Host = originHostname

		// Start a goroutine to handle hash verification.
		go func(hostname string) {
			// Get the context with a 10 second timeout.
			ctx, c := context.WithTimeout(context.Background(), time.Second*10)
			defer c()

			// Get the request and make it. Ignore any errors.
			req, err := http.NewRequestWithContext(ctx, "POST", u.String(), strings.NewReader(payload))
			if err != nil {
				// Hmmm odd, just return here.
				return
			}
			res, err := http.DefaultClient.Do(req)
			if err == nil {
				_ = res.Body.Close()
			}
		}(hostname)
	}
}
