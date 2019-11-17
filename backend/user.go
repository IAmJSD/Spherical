package backend

import (
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"golang.org/x/crypto/bcrypt"
	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

// UserAuth authenticates the user.
func UserAuth(ctx *fasthttp.RequestCtx) {
	// Gets the email and password.
	Email := string(ctx.Request.Header.Peek("Email"))
	Password := string(ctx.Request.Header.Peek("Password"))
	if Email == "" || Password == "" {
		ctx.Response.SetStatusCode(403)
		ctx.Response.Header.SetContentType("application/json")
		ctx.SetBody([]byte("null"))
		return
	}

	// Gets the user.
	user := GetUserByEmail(Email)
	if user == nil {
		ctx.Response.SetStatusCode(403)
		ctx.Response.Header.SetContentType("application/json")
		ctx.SetBody([]byte("null"))
		return
	}

	// Checks the password.
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(Password))
	if err != nil {
		ctx.Response.SetStatusCode(403)
		ctx.Response.Header.SetContentType("application/json")
		ctx.SetBody([]byte("null"))
		return
	}

	// Creates the auth token.
	AuthToken := uuid.Must(uuid.NewUUID()).String()
	user.Tokens = append(user.Tokens, AuthToken)
	_, err = r.Table("users").Get(user.Email).Update(map[string]interface{}{
		"tokens": user.Tokens,
	}).Run(RethinkConnection)
	if err != nil {
		panic(err)
	}

	// Returns the auth token.
	ctx.Response.SetStatusCode(403)
	ctx.Response.Header.SetContentType("application/json")
	ctx.SetBody([]byte("\"" + AuthToken + "\""))
}
