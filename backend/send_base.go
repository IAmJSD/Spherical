package backend

import (
	"html/template"
	"io/ioutil"

	"github.com/valyala/fasthttp"
)

// Base defines the base.
var base, _ = ioutil.ReadFile("./base.html")

// SendBase sends the base template.
func SendBase(Args map[string]interface{}, ctx *fasthttp.RequestCtx) {
	template := template.Must(template.New("base").Parse(string(base)))
	ctx.SetContentType("text/html; charset=utf-8")
	_ = template.Execute(ctx, Args)
}
