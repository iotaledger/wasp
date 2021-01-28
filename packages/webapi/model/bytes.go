package model

import (
	"encoding/base64"
	"encoding/json"
)

// Bytes is the base64 representation of a byte slice
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
