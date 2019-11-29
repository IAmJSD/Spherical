package backend

import (
	"encoding/json"

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

// UserPrivateKey returns the users private key.
func UserPrivateKey(ctx *fasthttp.RequestCtx) {
	// Gets the user.
	Token := string(ctx.Request.Header.Peek("Token"))
	if Token == "" {
		ctx.Response.SetStatusCode(403)
		ctx.Response.Header.SetContentType("application/json")
		ctx.SetBody([]byte("null"))
		return
	}
	user := GetUserByToken(Token)
	if user == nil {
		ctx.Response.SetStatusCode(403)
		ctx.Response.Header.SetContentType("application/json")
		ctx.SetBody([]byte("null"))
		return
	}

	// Returns the users private key in encrypted form.
	ctx.Response.SetStatusCode(200)
	ctx.Response.Header.SetContentType("application/json")
	j, err := json.Marshal(&user.EncryptedPrivateKey)
	if err != nil {
		panic(err)
	}
	ctx.Response.SetBody(j)
}

// UserProfile gets a users profile.
func UserProfile(ctx *fasthttp.RequestCtx) {
	// Gets the user.
	Token := string(ctx.Request.Header.Peek("Token"))
	if Token == "" {
		ctx.Response.SetStatusCode(403)
		ctx.Response.Header.SetContentType("application/json")
		ctx.SetBody([]byte("null"))
		return
	}
	user := GetUserByToken(Token)
	if user == nil {
		ctx.Response.SetStatusCode(403)
		ctx.Response.Header.SetContentType("application/json")
		ctx.SetBody([]byte("null"))
		return
	}

	// Returns the users profile in encrypted form.
	ctx.Response.SetStatusCode(200)
	ctx.Response.Header.SetContentType("application/json")
	j, err := json.Marshal(user.Info)
	if err != nil {
		panic(err)
	}
	ctx.Response.SetBody(j)
}
