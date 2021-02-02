package coretypes

import (
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/mr-tron/base58"
)

type Utils interface {
	HashBlake2b(data []byte) hashing.HashValue
	HashSha3(data []byte) hashing.HashValue
	EncodeBase58(data []byte) string
	DecodeBase58(s string) ([]byte, error)
	Hname(s string) Hname
	ValidED25519Signature(data []byte, pubKey []byte, signature []byte) bool
}

type utilImpl struct{}

func NewUtils() Utils {
	return utilImpl{}
}

func (u utilImpl) HashBlake2b(data []byte) hashing.HashValue {
	return hashing.HashDataBlake2b(data)
}

func (u utilImpl) HashSha3(data []byte) hashing.HashValue {
	return hashing.HashSha3(data)
}

func (u utilImpl) EncodeBase58(data []byte) string {
	return base58.Encode(data)
}

func (u utilImpl) DecodeBase58(s string) ([]byte, error) {
	return base58.Decode(s)
}

func (u utilImpl) Hname(s string) Hname {
	return Hn(s)
}

func (u utilImpl) ValidED25519Signature(data []byte, pubKey []byte, signature []byte) bool {
	pk, _, err := ed25519.PublicKeyFromBytes(pubKey)
	if err != nil {
		return false
	}
	sig, _, err := ed25519.SignatureFromBytes(signature)
	if err != nil {
		return false
	}
	return pk.VerifySignature(data, sig)
}
