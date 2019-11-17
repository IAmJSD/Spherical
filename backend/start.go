package backend

import (
	"log"
	"os"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

// RethinkConnection defines the RethinkDB connection.
var RethinkConnection *r.Session

// Start is the first function which will be called.
func Start() {
	// Creates the fasthttp router.
	router := fasthttprouter.New()

	// Loads in the frontend.
	FrontendLoader(router)

	// Loads in routes.
	RouteLoader(router)

	// Creates the RethinkDB connection.
	RethinkHost := os.Getenv("RETHINK_HOST")
	if RethinkHost == "" {
		RethinkHost = "127.0.0.1:28015"
	}
	RethinkPass := os.Getenv("RETHINK_PASSWORD")
	RethinkUser := os.Getenv("RETHINK_USER")
	if RethinkUser == "" {
		RethinkUser = "admin"
	}
	s, err := r.Connect(r.ConnectOpts{
		Address:  RethinkHost,
		Password: RethinkPass,
		Username: RethinkUser,
		Database: "spherical",
	})
	if err != nil {
		panic(err)
	}
	RethinkConnection = s

	// Loads the server.
	println("Spherical is loading into port 8000!")
	log.Fatal(fasthttp.ListenAndServe(":8000", router.Handler))
}
