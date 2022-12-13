package httproutes

import (
	"context"
	"net/http"

	"github.com/jakemakesstuff/spherical/config"
	"github.com/jakemakesstuff/spherical/httproutes/application"
	"github.com/jakemakesstuff/spherical/httproutes/oobe"
)

// SelectRouter is a function used to select the router based on the state of the config.
func SelectRouter(dev bool, pubKeyArmored string) http.Handler {
	appRouter := application.Router(dev)
	oobeRouter := oobe.Router(dev)
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx := context.WithValue(request.Context(), "dev", dev)
		ctx = context.WithValue(ctx, "pubKeyArmored", pubKeyArmored)
		request = request.WithContext(ctx)
		if config.Config().Setup {
			// Use the main router.
			appRouter.ServeHTTP(writer, request)
		} else {
			// Use the OOBE router.
			oobeRouter.ServeHTTP(writer, request)
		}
	})
}
