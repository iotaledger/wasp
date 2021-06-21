package model

import (
	"encoding/base64"
	"encoding/json"
)

// Bytes is the base64 representation of a byte slice.
// It is intended to be a replacement for any []byte attribute in JSON models.
// Normally it shouldn't be necessary since the standard json package already
// handles []byte data, but this makes sure that the swagger documentation
// shows examples correctly (instead of as if a []byte was json-encoded as an
// array of ints).
type Bytes string

func NewBytes(data []byte) Bytes {
	return Bytes(base64.StdEncoding.EncodeToString(data))
}

func (b Bytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(b))
}

func (b *Bytes) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	// verify encoding
	_, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	*b = Bytes(s)
	return err
}

func (b Bytes) Bytes() []byte {
	data, err := base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		// encoding should be already verified
		panic(err)
	}
	return data
}
