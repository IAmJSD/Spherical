package main

import (
	"bytes"
	"strings"

	"github.com/jakemakesstuff/spherical/hashverifier"
	"github.com/valyala/fasthttp"
)

var (
	verifyPath = []byte("/verify")
	post       = []byte("POST")
	options    = []byte("OPTIONS")

	trueB  = []byte("true")
	falseB = []byte("false")
)

func postHandler(ctx *fasthttp.RequestCtx, client *hashverifier.Client) {
	// Get the body.
	body := string(ctx.PostBody())
	if len(body) == 0 {
		ctx.Response.SetStatusCode(400)
		ctx.Response.SetBodyString("Not found.")
		return
	}

	// Get the response data.
	xSkip := strings.Split(string(ctx.Request.Header.Peek("X-Skip")), ",")
	for i, v := range xSkip {
		xSkip[i] = strings.TrimSpace(strings.ToLower(v))
	}
	resp := client.ProcessHashBlob(ctx, body, xSkip)
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetStatusCode(204)
	if resp {
		ctx.Response.SetBody(trueB)
	} else {
		ctx.Response.SetBody(falseB)
	}
}

func handler(client *hashverifier.Client) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// CHeck if this is /verify.
		if bytes.Equal(verifyPath, ctx.Path()) {
			// If this is OPTIONS, send the headers.
			if bytes.Equal(options, ctx.Method()) {
				ctx.Response.Header.Set("Allow", "OPTIONS, POST")
				ctx.Response.Header.Set("Cache-Control", "max-age=604800")
				ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
				ctx.Response.SetStatusCode(204)
				return
			}

			// If this is a POST, do the request.
			if bytes.Equal(post, ctx.Method()) {
				postHandler(ctx, client)
				return
			}
		}

		// Return the not found.
		ctx.Response.SetStatusCode(404)
		ctx.Response.SetBodyString("Not found.")
	}
}
