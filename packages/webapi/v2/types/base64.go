package types

import (
	"encoding/base64"
	"encoding/json"
)

// Base64 is the base64 representation of a byte slice.
// It is intended to be a replacement for any []byte attribute in JSON models.
// Normally it shouldn't be necessary since the standard json package already
// handles []byte data, but this makes sure that the swagger documentation
// shows examples correctly (instead of as if a []byte was json-encoded as an
// array of ints).
type Base64 string

func NewBase64(data []byte) Base64 {
	return Base64(base64.StdEncoding.EncodeToString(data))
}

func (b Base64) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(b))
}

func (b *Base64) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	// verify encoding
	_, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	*b = Base64(s)
	return err
}

func (b Base64) Bytes() []byte {
	data, err := base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		// encoding should be already verified
		panic(err)
	}
	return data
}
