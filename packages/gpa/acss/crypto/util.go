package crypto

import (
	"crypto/cipher"
	"crypto/sha256"
	"fmt"
	"io"

	"go.dedis.ch/kyber/v3"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

// threshold returns the threshold for n.
func threshold(n int) int {
	return (n-1)/3 + 1 // threshold is fixed ⌊n/3⌋+1
}

// newAEAD creates a new AEAD cipher based on secret and info.
func newAEAD(secret, salt, info []byte) cipher.AEAD {
	h := hkdf.New(sha256.New, secret, salt, info)
	key := make([]byte, AEADKeySize)
	if _, err := io.ReadFull(h, key); err != nil {
		panic(err)
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		panic(err)
	}
	return aead
}

// encryptScalar encrypts a scalar and returns the cipher text.
func encryptScalar(s kyber.Scalar, aead cipher.AEAD) []byte {
	sBytes, err := s.MarshalBinary()
	if err != nil {
		panic(err)
	}
	nonce := make([]byte, aead.NonceSize())
	return aead.Seal(nil, nonce, sBytes, nil)
}

// decryptScalar decrypts cipherText and sets the corresponding scalar dst.
func decryptScalar(dst kyber.Scalar, aead cipher.AEAD, cipherText []byte) error {
	nonce := make([]byte, aead.NonceSize())
	plaintext, err := aead.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return err
	}
	if err := dst.UnmarshalBinary(plaintext); err != nil {
		return err
	}
	return nil
}

func contextInfo(index int) []byte {
	return []byte(fmt.Sprintf("ACSS-%d", index))
}
