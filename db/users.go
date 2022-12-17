package db

import (
	"context"
	"errors"
	"runtime"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jakemakesstuff/spherical/errhandler"
	"github.com/jakemakesstuff/spherical/utils/httperrors"
	"github.com/jakemakesstuff/spherical/utils/password"
	"github.com/jakemakesstuff/spherical/utils/regexes"
)

// UserFlag is used to define a flag (or set of flags) on a user.
type UserFlag uint64

const (
	// UserFlagOwner is used to define a server owner.
	UserFlagOwner UserFlag = 1 << iota
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

// ErrInvalidCredentials is used to define an error thrown by CheckPassword.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrNotPasswordAuth is used to define an error thrown by CheckPassword.
var ErrNotPasswordAuth = errors.New("user is not password authenticated")

// AuthenticateUserByPassword is used to authenticate a user by their password. If there is an error, it will return either:
// a) Any other error - ignore the rest of this if this is the case.
// b) A half-authentication token - this is a special type of token that can be used to authenticate the user with a 2FA
// code. If this is the case, the error will be of type httperrors.HalfAuthentication and contain more information.
// c) ErrInvalidCredentials - this is returned if the credentials are invalid.
// d) ErrNotPasswordAuth - this is returned if the user is not using password authentication.
// !! PLEASE READ THE ABOVE CAREFULLY - THIS IS AUTHENTICATION, WE CANNOT AFFORD TO MAKE ERRORS !!
func AuthenticateUserByPassword(ctx context.Context, usernameOrEmail, passwordInput string) (uint64, error) {
	// Get the user ID and password.
	query := "SELECT password_authentication_users.user_id, password_authentication_users.password FROM " +
		"users LEFT JOIN password_authentication_users ON users.user_id = password_authentication_users.user_id " +
		"WHERE users.username = $1 OR users.email = $1"
	var userIdPtr *uint64
	var passwordHash *string
	err := dbConn().QueryRow(ctx, query, usernameOrEmail).Scan(&userIdPtr, &passwordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Turn the error into an invalid credentials error.
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	// Check if the user is using password authentication.
	if userIdPtr == nil {
		return 0, ErrNotPasswordAuth
	}

	// Update last_login_attempt on the user.
	userId := *userIdPtr
	_, err = dbConn().Exec(ctx, "UPDATE users SET last_login_attempt = NOW() WHERE user_id = $1", userId)
	if err != nil {
		errhandler.Process(err, "db/users/AuthenticateUserByPassword", map[string]string{
			"action": "update last_login_attempt, invalid auth pathway",
		})
	}

	// Validate the password.
	hashValid := password.Validate(*passwordHash, passwordInput)
	if !hashValid {
		return 0, ErrInvalidCredentials
	}

	// We now know the password is valid. Let's check if the user has 2FA enabled.
	query = "SELECT totp_token IS NOT NULL FROM users_mfa WHERE user_id = $1"
	totpEnabled := false
	err = dbConn().QueryRow(ctx, query, userId).Scan(&totpEnabled)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	// Handle if the user has 2FA enabled.
	if totpEnabled {
		// Generate a half-authentication token.
		token, err := BuildHalfToken(ctx, userId)
		if err != nil {
			return 0, err
		}

		// Return the half-authentication error.
		return 0, httperrors.HalfAuthentication{
			SupportedMethods: []string{"totp", "recovery"},
			HalfToken:        token,
		}
	}

	// Return the user ID.
	return userId, nil
}
