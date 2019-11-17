package backend

import (
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

// RouteLoader loads in all of the routes.
func RouteLoader(router *fasthttprouter.Router) {
	router.GET("/", func(ctx *fasthttp.RequestCtx) {
		SendBase(map[string]interface{}{
			"Title":       "Spherical",
			"Description": "Spherical is an encrypted and secure alternative to platforms such as Facebook.",
		}, ctx)
	})
	router.GET("/api/v1/user/auth", UserAuth)
	router.GET("/api/v1/user/private_key", UserPrivateKey)
	router.GET("/api/v1/user/profile", UserProfile)
}
