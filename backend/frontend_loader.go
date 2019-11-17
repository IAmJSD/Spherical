package backend

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

// FrontendLoader loads all files from the frontend and makes routes for all of them.
func FrontendLoader(router *fasthttprouter.Router) {
	files, err := ioutil.ReadDir("./frontend/dist")
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		FileLoaded, err := ioutil.ReadFile("./frontend/dist/" + f.Name())
		if err != nil {
			panic(err)
		}
		fsplit := strings.Split(f.Name(), ".")
		ext := fsplit[len(fsplit)-1]
		var m string
		if ext == "html" {
			m = "text/html; charset=utf-8"
		} else if ext == "css" {
			m = "text/css; charset=utf-8"
		} else if ext == "js" {
			m = "text/javascript; charset=utf-8"
		} else {
			m = http.DetectContentType(FileLoaded)
		}
		router.GET("/"+f.Name(), func(ctx *fasthttp.RequestCtx) {
			ctx.SetStatusCode(200)
			ctx.SetContentType(m)
			ctx.Write(FileLoaded)
		})
	}
}
