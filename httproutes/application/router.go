package application

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jakemakesstuff/spherical/config"
	"github.com/jakemakesstuff/spherical/html"
	"github.com/jakemakesstuff/spherical/httproutes/application/apiv1"
	"github.com/jakemakesstuff/spherical/httproutes/shared"
	"github.com/jakemakesstuff/spherical/i18n"
)

func genericHtmlWriterHn(w http.ResponseWriter, r *http.Request) {
	c := config.Config()
	title := c.ServerName
	description := i18n.GetWithRequest(r, "httproutes/application/router:description")
	html.Handler(
		title, description, false, nil, http.StatusOK)(w, r)
}

// Router is used to return this packages router.
func Router(dev bool) http.Handler {
	r := mux.NewRouter()

	r.Path("/").HandlerFunc(genericHtmlWriterHn)
	r.Path("/app").HandlerFunc(genericHtmlWriterHn)

	r.PathPrefix("/setup").HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		http.Redirect(writer, req, "/", http.StatusPermanentRedirect)
	})
	r.Path("/spherical.pub").HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		pubKeyArmored := req.Context().Value("pubKeyArmored").(string)
		writer.Header().Set("Content-Type", "application/pgp-keys")
		writer.Header().Set("Content-Length", strconv.Itoa(len(pubKeyArmored)))
		_, _ = writer.Write([]byte(pubKeyArmored))
	})
	apiv1.Router(r.PathPrefix("/api/v1").Subrouter())

	shared.Router(r, dev)

	return r
}
