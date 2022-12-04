package db

import (
	"context"
	"errors"
	"runtime"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jakemakesstuff/spherical/utils/password"
	"github.com/jakemakesstuff/spherical/utils/regexes"
)

// UserFlag is used to define a flag (or set of flags) on a user.
type UserFlag uint64

const (
	// UserFlagOwner is used to define a server owner.
	UserFlagOwner = 1 << iota
)

// UserPasswordCreateOpts is the options required to create a user with a password.
type UserPasswordCreateOpts struct {
	Username  string
	Email     string
	Password  string
	Confirmed bool
	Flags     UserFlag
}

// CreatePasswordUserError is used to define an error thrown by CreatePasswordUser.
type CreatePasswordUserError struct {
	err     error
	poValue string
}

// Error implements the error interface.
func (p CreatePasswordUserError) Error() string { return p.err.Error() }

// Unwrap returns the underlying error.
func (p CreatePasswordUserError) Unwrap() error { return p.err }

// POValue is used to get the PO value of the error.
func (p CreatePasswordUserError) POValue() string { return p.poValue }

// CreatePasswordUser is used to create a user with a password.
func CreatePasswordUser(ctx context.Context, opts UserPasswordCreateOpts) (userId uint64, err error) {
	// Let's forget the plain text password as fast as possible.
	if !regexes.Password.MatchString(opts.Password) {
		return 0, CreatePasswordUserError{
			err:     errors.New("password is not valid"),
			poValue: "db/users:invalid_password",
		}
	}
	opts.Password = password.HashAndSalt(opts.Password)
	runtime.GC()

	// Check the username and email.
	if !regexes.Username.MatchString(opts.Username) {
		return 0, CreatePasswordUserError{
			err:     errors.New("username is not valid"),
			poValue: "db/users:invalid_username",
		}
	}
	if !regexes.Email.MatchString(opts.Email) {
		return 0, CreatePasswordUserError{
			err:     errors.New("email address is not valid"),
			poValue: "db/users:invalid_email",
		}
	}

	// Create the user ID.
	userId = uint64(snowGen.Generate())

	// Defines the queries to create the user and password in the DB.
	q1 := "INSERT INTO users (user_id, username, email, confirmed, flags) VALUES ($1, $2, $3, $4, $5)"
	q2 := "INSERT INTO password_authentication_users (user_id, password) VALUES ($1, $2)"
	batch := &pgx.Batch{}
	batch.Queue(q1, userId, opts.Username, opts.Email, opts.Confirmed, opts.Flags)
	batch.Queue(q2, userId, opts.Password)
	res := dbConn().SendBatch(ctx, batch)
	defer res.Close()
	if _, err = res.Exec(); err != nil {
		p, ok := err.(*pgconn.PgError)
		if ok && p.Code == "23502" {
			return 0, CreatePasswordUserError{
				err:     err,
				poValue: "db/users:user_or_email_used",
			}
		}
		return 0, CreatePasswordUserError{
			err:     err,
			poValue: "db/users:insert_fail",
		}
	}
	if _, err = res.Exec(); err != nil {
		return 0, CreatePasswordUserError{
			err:     err,
			poValue: "db/users:insert_fail",
		}
	}
	return userId, err
}

// HasOwner is used to check if the server has an owner. THIS IS VERY SLOW AND
// SHOULD ONLY BE RAN DURING SETUP! It intentionally doesn't return an error
// because errors are picked up later in setup.
func HasOwner(ctx context.Context) bool {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE flags & 1 != 0)"
	val := false
	_ = dbConn().QueryRow(ctx, query).Scan(&val)
	return val
}

// ScanUserFromToken is used to scan a user from a token.
func ScanUserFromToken(ctx context.Context, token string, user any) error {
	return pgxscan.Get(
		ctx, dbConn(), user,
		"SELECT * FROM users WHERE user_id = (SELECT user_id FROM sessions WHERE token = $1)", token)
}
