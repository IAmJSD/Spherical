package backend

import (
	"strings"

	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

// UserInfo defines information about the user.
type UserInfo struct {
	FirstName      string  `gorethink:"firstName"`
	LastName       string  `gorethink:"lastName"`
	Email          *string `gorethink:"-"` // This value should be manually set to avoid repeating stuff.
	EmailConfirmed bool    `gorethink:"emailConfirmed"`
	Description    string  `gorethink:"description"`
	PhoneHash      *string `gorethink:"phoneHash"`
	TwoFactor      bool    `gorethink:"twoFactor"`
	ProfilePicture string  `gorethink:"profilePicture"`
	PublicKey      string  `gorethink:"publicKey"`
}

// User is the base for a user.
type User struct {
	Info     *UserInfo `rethinkdb:"email,reference" rethinkdb_ref:"id"`
	Password string    `gorethink:"password"`
	Email    string    `gorethink:"id,omitempty"`
	Tokens   []string  `gorethink:"tokens"`
}

// GetUserByEmail gets the user by their e-mail address.
func GetUserByEmail(Email string) *User {
	res, err := r.Table("users").Get(strings.ToLower(Email)).Run(RethinkConnection)
	if err != nil {
		return nil
	}
	if res.IsNil() {
		return nil
	}
	var user User
	err = res.One(&user)
	if err != nil {
		panic(err)
	}
	return &user
}
