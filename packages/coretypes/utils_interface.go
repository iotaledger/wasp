package coretypes

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/mr-tron/base58"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign/bdn"
)

// Utils implement various utilities which are faster on host side than on wasm VM
// Implement deterministic stateless computations
type Utils interface {
	Base58Decode(s string) ([]byte, error)
	Base58Encode(data []byte) string
	HashBlake2b(data []byte) hashing.HashValue
	HashSha3(data []byte) hashing.HashValue
	Hname(s string) Hname
	ValidED25519Signature(data []byte, pubKey []byte, signature []byte) bool
	ValidBLSSignature(data []byte, pubKey []byte, signature []byte) bool
	AggregateBLSSignatures(pubKeysBin [][]byte, sigsBin [][]byte) ([]byte, []byte)
}

type utilImpl struct {
	log LogInterface
}

func NewUtils(log LogInterface) Utils {
	return utilImpl{log}
}

func (u utilImpl) Base58Decode(s string) ([]byte, error) {
	return base58.Decode(s)
}

func (u utilImpl) Base58Encode(data []byte) string {
	return base58.Encode(data)
}

func (u utilImpl) HashBlake2b(data []byte) hashing.HashValue {
	return hashing.HashDataBlake2b(data)
}

func (u utilImpl) HashSha3(data []byte) hashing.HashValue {
	return hashing.HashSha3(data)
}

func (u utilImpl) Hname(s string) Hname {
	return Hn(s)
}

func (u utilImpl) ValidED25519Signature(data []byte, pubKey []byte, signature []byte) bool {
	pk, _, err := ed25519.PublicKeyFromBytes(pubKey)
	if err != nil {
		u.log.Panicf("ValidED25519Signature: wrong public key bytes")
	}
	sig, _, err := ed25519.SignatureFromBytes(signature)
	if err != nil {
		u.log.Panicf("ValidED25519Signature: wrong signature bytes")
	}
	return pk.VerifySignature(data, sig)
}

var suite = bn256.NewSuite()

func (u utilImpl) ValidBLSSignature(data []byte, pubKeyBin []byte, sigBin []byte) bool {
	pubKey := suite.G2().Point()
	var err error
	if err = pubKey.UnmarshalBinary(pubKeyBin); err != nil {
		u.log.Panicf("ValidBLSSignature: wrong public key bytes")
	}
	err = bdn.Verify(suite, pubKey, data, sigBin)
	if err != nil {
		u.log.Infof("ValidBLSSignature: %v", err)
	}
	return err == nil
}

// AggregateBLSSignatures
// TODO: optimize redundant binary manipulation.
//   Implement more flexible access to parts of Signature
func (u utilImpl) AggregateBLSSignatures(pubKeysBin [][]byte, sigsBin [][]byte) ([]byte, []byte) {
	if len(sigsBin) == 0 || len(pubKeysBin) != len(sigsBin) {
		u.log.Panicf("number of publik keys must be equal to the number of signatures and not empty")
	}
	sigs := make([]signaturescheme.Signature, len(sigsBin))
	for i := range sigs {
		sigs[i] = signaturescheme.NewBLSSignature(pubKeysBin[i], sigsBin[i])
	}
	ret, err := signaturescheme.AggregateBLSSignatures(sigs...)
	if err != nil {
		u.log.Panicf("AggregateBLSSignatures: %v", err)
	}
	pubKeyBin := ret.Bytes()[1 : 1+signaturescheme.BLSPublicKeySize]
	sigBin := ret.Bytes()[1+signaturescheme.BLSPublicKeySize:]

	return pubKeyBin, sigBin
}
