package cryptolib

import (
	"crypto/ed25519"
	"errors"
)

const SignatureSize = ed25519.SignatureSize

func SignatureFromBytes(bytes []byte) (result [SignatureSize]byte, consumedBytes int, err error) {
	if len(bytes) < SignatureSize {
		err = errors.New("bytes too short")
		return
	}

	copy(result[:SignatureSize], bytes)
	consumedBytes = SignatureSize

	return
}
