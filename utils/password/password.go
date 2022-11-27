package password

import (
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

// HashAndSalt is used to hash and salt a password with bcrypt.
func HashAndSalt(password string) string {
	b, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	return base64.StdEncoding.EncodeToString(b)
}

// Validate is used to validate the specified password hash is the same as the password.
func Validate(hash, password string) bool {
	hashB, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false
	}
	return bcrypt.CompareHashAndPassword(hashB, []byte(password)) == nil
}
