package hashverifier

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/google/uuid"
)

// CrossNodeMessage is used to define a message claiming to be from another node. Note that marshalling into this
// structure does NOT automatically mean it is secure. You MUST call Validate on this struct if you are accepting it.
type CrossNodeMessage struct {
	// Hostname is the hostname of the node claiming to have sent this message.
	Hostname string `json:"h"`

	// Signature defines the PGP armored signature for the base64 encoded version of the sha-256 hashed JSON.
	Signature string `json:"s"`

	// Data is the data that the node wants to send. This MUST contain a salt key and the value SHOULD contain unique data.
	Data json.RawMessage `json:"d"`
}

// UserFacingError specified an error that the user can see. Should be sent as 400.
type UserFacingError struct {
	message string
}

// Error implements the error interface.
func (u UserFacingError) Error() string {
	return u.message
}

// Validate is used to validate that a message actually came from the node specified. Consensus cannot be zero.
func (msg CrossNodeMessage) Validate(ctx context.Context, consensus uint, informers, trusted []string) error {
	// Check if consensus is zero. It should not be.
	if consensus == 0 {
		return errors.New("consensus cannot be zero")
	}

	// Make sure that data contains the salt key.
	type saltBody struct {
		Salt string `json:"salt"`
	}
	var b saltBody
	err := json.Unmarshal(msg.Data, &b)
	if err != nil || b.Salt == "" {
		return UserFacingError{message: "salt not in json payload"}
	}

	// Make sure the hostname doesn't contain \n and is not blank.
	if strings.Contains(msg.Hostname, "\n") || msg.Hostname == "" {
		return UserFacingError{message: "hostname is not found"}
	}

	// Make sure the signature is not blank.
	if msg.Signature == "" {
		return UserFacingError{message: "signature is not found"}
	}

	// Hash the JSON body.
	sha := sha256.New()
	_, _ = sha.Write(msg.Data)
	hash := sha.Sum(nil)
	hashStr := base64.StdEncoding.EncodeToString(hash)

	// Process the hash blob.
	valid := NewClient(nil, informers, trusted, "", consensus).ProcessHashBlob(
		ctx, msg.Hostname+"\n"+hashStr+"\n"+msg.Signature, nil)
	if !valid {
		return UserFacingError{message: "cannot verify the message came from the hostname specified"}
	}
	return nil
}

// NewCrossNodeMessage is used to make a new cross node message. It will automatically inject the salt, PGP sign the value,
// and build the structure.
func NewCrossNodeMessage(data any, hostname string, pgpPriv *crypto.KeyRing) (CrossNodeMessage, error) {
	// Marshal the JSON.
	b, err := json.Marshal(data)
	if err != nil {
		return CrossNodeMessage{}, err
	}

	// Check this is an object.
	if b[0] != '{' {
		return CrossNodeMessage{}, errors.New("data is not a object")
	}

	// Trim the first character from this data and inject the salt.
	b = b[1:]
	randUuid, err := uuid.NewRandom()
	if err != nil {
		return CrossNodeMessage{}, err
	}
	jsonStart := []byte(`{"salt":"` + randUuid.String() + `",`)
	b = append(jsonStart, b...)

	// Hash the JSON body.
	sha := sha256.New()
	_, _ = sha.Write(b)
	hash := sha.Sum(nil)
	hashStr := base64.StdEncoding.EncodeToString(hash)

	// PGP sign the string.
	sig, err := helper.SignCleartextMessage(pgpPriv, hashStr)
	if err != nil {
		return CrossNodeMessage{}, err
	}

	// Return the structure.
	return CrossNodeMessage{
		Hostname:  hostname,
		Signature: sig,
		Data:      b,
	}, nil
}
