package model

import (
	"encoding/json"

	"github.com/iotaledger/wasp/packages/hashing"
)

// HashValue is the hex representation of a hashing.HashValue
type HashValue string

func NewHashValue(h hashing.HashValue) HashValue {
	return HashValue(h.String())
}

func (h HashValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(h))
}

func (h *HashValue) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	// verify encoding
	_, err := hashing.HashValueFromHex(s)
	if err != nil {
		return err
	}
	*h = HashValue(s)
	return nil
}

func (h HashValue) HashValue() hashing.HashValue {
	r, err := hashing.HashValueFromHex(string(h))
	if err != nil {
		// encoding should be already verified
		panic(err)
	}
	return r
}
