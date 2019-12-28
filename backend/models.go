package backend

import (
	"strings"

	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

// UserInfo defines information about the user.
type UserInfo struct {
	ID string `gorethink:"id,omitempty"`
	FirstName      string  `gorethink:"firstName"`
	LastName       string  `gorethink:"lastName"`
	Email          string `gorethink:"email"`
	EmailConfirmed bool    `gorethink:"emailConfirmed"`
	Description    string  `gorethink:"description"`
	PhoneHash      *string `gorethink:"phoneHash"`
	TwoFactor      bool    `gorethink:"twoFactor"`
	ProfilePicture string  `gorethink:"profilePicture"`
	CreatedAt      int     `gorethink:"createdAt"`
}

// User is the base for a user.
type User struct {
	Info                *UserInfo `rethinkdb:"userId,reference" rethinkdb_ref:"id"`
	UserID string `gorethink:"userId"`
	Password            string    `gorethink:"password"`
	Email               string    `gorethink:"id,omitempty"`
	Tokens              []string  `gorethink:"tokens"`
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

// GetUserByToken gets the user by their token.
func GetUserByToken(Token string) *User {
	res, err := r.Table("users").GetAllByIndex("tokens", Token).Run(RethinkConnection)
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

// GetUserByID gets the user by their ID.
func GetUserByID(ID string) *User {
	res, err := r.Table("user_info").Get(ID).Run(RethinkConnection)
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
	return GetUserByEmail(user.UserID)
}
