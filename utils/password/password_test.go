package password

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPassword(t *testing.T) {
	value := HashAndSalt("password123")
	assert.NotEqual(t, "password123", value)
	assert.True(t, Validate(value, "password123"))
	assert.False(t, Validate(
		base64.StdEncoding.EncodeToString([]byte("random_shit")),
		"password123"))
}
