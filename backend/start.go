package backend

import (
	"log"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

// Start is the first function which will be called.
func Start() {
	// Creates the fasthttp router.
	router := fasthttprouter.New()

	// Loads in the frontend.
	FrontendLoader(router)

	// Loads in routes.
	RouteLoader(router)

	// Loads the server.
	println("Spherical is loading into port 8000!")
	log.Fatal(fasthttp.ListenAndServe("0.0.0.0:8000", router.Handler))
}
