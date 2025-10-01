package bls

import (
	"bytes"

	"github.com/btcsuite/btcd/btcutil/base58"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/bdn"

	"github.com/iotaledger/hive.go/ierrors"
)

// PrivateKey is the type of BLS private keys.
type PrivateKey struct {
	Scalar kyber.Scalar
}

// PrivateKeyFromBytes creates a PrivateKey from the given bytes.
func PrivateKeyFromBytes(b []byte) (privateKey PrivateKey, err error) {
	buffer := bytes.NewReader(b)

	privateKey, err = PrivateKeyFromMarshalUtil(buffer)
	if err != nil {
		err = ierrors.Wrap(err, "failed to parse PublicKey from MarshalUtil")
		return PrivateKey{}, err
	}
	return privateKey, nil
}

// PrivateKeyFromMarshalUtil unmarshals a PrivateKey using a MarshalUtil (for easier unmarshalling).
func PrivateKeyFromMarshalUtil(reader *bytes.Reader) (privateKey PrivateKey, err error) {
	privateKeyBytes := make([]byte, PrivateKeySize)

	n, err := reader.Read(privateKeyBytes)
	if err != nil {
		err = ierrors.Wrapf(ErrParseBytesFailed, "failed to read PrivateKey bytes: %w", err)
		return PrivateKey{}, err
	}

	if n != PrivateKeySize {
		err = ierrors.Wrapf(ErrParseBytesFailed, "failed to read PrivateKey length: %d", n)
		return PrivateKey{}, err
	}

	if err = privateKey.Scalar.UnmarshalBinary(privateKeyBytes); err != nil {
		err = ierrors.Wrapf(ErrParseBytesFailed, "failed to unmarshal PrivateKey: %w", err)

		return PrivateKey{}, err
	}

	return privateKey, nil
}

// PrivateKeyFromRandomness generates a new random PrivateKey.
func PrivateKeyFromRandomness() (privateKey PrivateKey) {
	privateKey.Scalar, _ = bdn.NewKeyPair(blsSuite, randomness)
	return privateKey
}

// PublicKey returns the PublicKey corresponding to the PrivateKey.
func (p PrivateKey) PublicKey() PublicKey {
	return PublicKey{
		Point: blsSuite.G2().Point().Mul(p.Scalar, nil),
	}
}

// Sign signs the message and returns a SignatureWithPublicKey.
func (p PrivateKey) Sign(data []byte) (signatureWithPublicKey SignatureWithPublicKey, err error) {
	sig, err := bdn.Sign(blsSuite, p.Scalar, data)
	if err != nil {
		err = ierrors.Wrapf(ErrBLSFailed, "failed to sign data: %w", err)
		return SignatureWithPublicKey{}, err
	}

	signatureWithPublicKey.PublicKey = p.PublicKey()
	copy(signatureWithPublicKey.Signature[:], sig)

	return signatureWithPublicKey, nil
}

// Bytes returns a marshaled version of the PrivateKey.
func (p PrivateKey) Bytes() (bytes []byte) {
	bytes, err := p.Scalar.MarshalBinary()
	if err != nil {
		panic(err)
	}

	return bytes
}

// Base58 returns a base58 encoded version of the PrivateKey.
func (p PrivateKey) Base58() string {
	return base58.Encode(p.Bytes())
}

// String returns a human-readable version of the PrivateKey (base58 encoded).
func (p PrivateKey) String() string {
	return p.Base58()
}
