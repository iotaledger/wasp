package cryptolib

import (
	"crypto/ed25519"
	"errors"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

const SignatureSize = ed25519.SignatureSize

func SignatureFromBytes(bytes []byte) (result [SignatureSize]byte, err error) {
	if len(bytes) < SignatureSize {
		err = errors.New("bytes too short")
		return
	}
	copy(result[:], bytes)
	return
}

func IsVariantKeyPairValid(variantKeyPair VariantKeyPair) bool {
	if variantKeyPair == nil {
		return false
	}

	return !variantKeyPair.IsNil()
}

// TODO ADDRESS: standard implementation?
func EncodeHex(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return hexutil.Encode(b)
}

// TODO ADDRESS: standard implementation?
// DecodeHex decodes the given hex string to bytes. It expects the 0x prefix.
func DecodeHex(s string) ([]byte, error) {
	b, err := hexutil.Decode(s)
	if err != nil {
		if err == hexutil.ErrEmptyString {
			return []byte{}, nil
		}
		return nil, err
	}
	return b, nil
}
