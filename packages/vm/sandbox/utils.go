package sandbox

import (
	"github.com/iotaledger/hive.go/crypto/bls"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/mr-tron/base58"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"golang.org/x/xerrors"
)

type utilImpl struct {
	gas iscp.Gas
}

// needed separate implementation to resolve conflict between function names
type utilImplBLS struct {
	gas iscp.Gas
}

func NewUtils(gas iscp.Gas) iscp.Utils {
	return utilImpl{gas}
}

// ------ iscp.Utils() interface

func (u utilImpl) Base58() iscp.Base58 {
	return u
}

func (u utilImpl) Hashing() iscp.Hashing {
	return u
}

func (u utilImpl) ED25519() iscp.ED25519 {
	return u
}

func (u utilImpl) BLS() iscp.BLS {
	return utilImplBLS{u.gas}
}

// --- iscp.Base58 interface

func (u utilImpl) Decode(s string) ([]byte, error) {
	u.gas.Burn(gas.BurnCodeUtilsBase58Decode)
	return base58.Decode(s)
}

func (u utilImpl) Encode(data []byte) string {
	u.gas.Burn(gas.BurnCodeUtilsBase58Encode)
	return base58.Encode(data)
}

// --- iscp.Hashing interface

func (u utilImpl) Blake2b(data []byte) hashing.HashValue {
	u.gas.Burn(gas.BurnCodeUtilsHashingBlake2b)
	return hashing.HashDataBlake2b(data)
}

func (u utilImpl) Sha3(data []byte) hashing.HashValue {
	u.gas.Burn(gas.BurnCodeUtilsHashingSha3)
	return hashing.HashSha3(data)
}

func (u utilImpl) Hname(name string) iscp.Hname {
	u.gas.Burn(gas.BurnCodeUtilsHashingHname)
	return iscp.Hn(name)
}

// --- iscp.ED25519 interface

func (u utilImpl) ValidSignature(data []byte, pubKey []byte, signature []byte) bool {
	u.gas.Burn(gas.BurnCodeUtilsED25519ValidSig)
	pk, err := cryptolib.PublicKeyFromBytes(pubKey)
	if err != nil {
		return false
	}
	sig, _, err := cryptolib.SignatureFromBytes(signature)
	if err != nil {
		return false
	}
	return cryptolib.Verify(pk, data, sig[:])
}

func (u utilImpl) AddressFromPublicKey(pubKey []byte) (iotago.Address, error) {
	u.gas.Burn(gas.BurnCodeUtilsED25519AddrFromPubKey)
	return cryptolib.Ed25519AddressFromPubKey(pubKey), nil
}

// iscp.BLS interface
var suite = bn256.NewSuite()

func (u utilImplBLS) ValidSignature(data []byte, pubKeyBin []byte, signature []byte) bool {
	u.gas.Burn(gas.BurnCodeUtilsBLSValidSignature)
	pubKey := suite.G2().Point()
	var err error
	if err = pubKey.UnmarshalBinary(pubKeyBin); err != nil {
		return false
	}
	return bdn.Verify(suite, pubKey, data, signature) == nil
}

func (u utilImplBLS) AddressFromPublicKey(pubKeyBin []byte) (iotago.Address, error) {
	panic("deprecate BLS")
	// u.gas.Burn(gas.UtilsBLSAddressFromPublicKey)
	// pubKey := suite.G2().Point()
	// if err := pubKey.UnmarshalBinary(pubKeyBin); err != nil {
	// 	return nil, xerrors.Errorf("BLSUtil: wrong public key bytes")
	// }
	// return iotago.NewBLSAddress(pubKeyBin), nil
}

func (u utilImplBLS) AggregateBLSSignatures(pubKeysBin [][]byte, sigsBin [][]byte) ([]byte, []byte, error) {
	if len(sigsBin) == 0 || len(pubKeysBin) != len(sigsBin) {
		return nil, nil, xerrors.Errorf("BLSUtil: number of public keys must be equal to the number of signatures and not empty")
	}
	u.gas.Burn(gas.BurnCodeUtilsBLSAggregateBLS1P, uint64(len(sigsBin)))

	sigPubKey := make([]bls.SignatureWithPublicKey, len(pubKeysBin))
	for i := range pubKeysBin {
		pubKey, _, err := bls.PublicKeyFromBytes(pubKeysBin[i])
		if err != nil {
			return nil, nil, xerrors.Errorf("BLSUtil: wrong public key bytes: %v", err)
		}
		sig, _, err := bls.SignatureFromBytes(sigsBin[i])
		if err != nil {
			return nil, nil, xerrors.Errorf("BLSUtil: wrong signature bytes: %v", err)
		}
		sigPubKey[i] = bls.NewSignatureWithPublicKey(pubKey, sig)
	}

	aggregatedSignature, err := bls.AggregateSignatures(sigPubKey...)
	if err != nil {
		return nil, nil, xerrors.Errorf("BLSUtil: fialed aggregate signatures: %v", err)
	}

	return aggregatedSignature.PublicKey.Bytes(), aggregatedSignature.Signature.Bytes(), nil
}
