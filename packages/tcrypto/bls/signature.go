package bls

import (
	"bytes"
	"slices"

	"github.com/btcsuite/btcd/btcutil/base58"

	"github.com/iotaledger/hive.go/ierrors"
)

// region Signature ////////////////////////////////////////////////////////////////////////////////////////////////////

// Signature is the type of a raw BLS signature.
type Signature [SignatureSize]byte

// SignatureFromBytes unmarshals a Signature from a sequence of bytes.
func SignatureFromBytes(b []byte) (signature Signature, err error) {
	reader := bytes.NewReader(b)

	if signature, err = SignatureFromReader(reader); err != nil {
		err = ierrors.Wrap(err, "failed to parse Signature from Reader")

		return
	}
	return
}

// SignatureFromBase58EncodedString creates a Signature from a base58 encoded string.
func SignatureFromBase58EncodedString(base58EncodedString string) (signature Signature, err error) {
	bytes := base58.Decode(base58EncodedString)
	if len(bytes) == 0 {
		err = ierrors.Wrapf(ErrBase58DecodeFailed, "error while decoding base58 encoded Signature: %s", base58EncodedString)

		return
	}

	if signature, err = SignatureFromBytes(bytes); err != nil {
		err = ierrors.Wrap(err, "failed to parse Signature from bytes")

		return
	}

	return
}

// SignatureFromReader unmarshals a Signature using a MarshalUtil (for easier unmarshalling).
func SignatureFromReader(reader *bytes.Reader) (signature Signature, err error) {
	buffer := make([]byte, SignatureSize)

	n, err := reader.Read(buffer)
	if err != nil {
		err = ierrors.Wrapf(ErrParseBytesFailed, "failed to read signature bytes: %w", err)
		return
	}

	if n != SignatureSize {
		err = ierrors.Wrapf(ErrParseBytesFailed, "failed to read Signature length: %d", n)
		return
	}

	copy(signature[:], buffer)

	return
}

// Bytes returns a marshaled version of the Signature.
func (s Signature) Bytes() []byte {
	return s[:]
}

// Base58 returns a base58 encoded version of the Signature.
func (s Signature) Base58() string {
	return base58.Encode(s.Bytes())
}

// String returns a human-readable version of the signature.
func (s Signature) String() string {
	return s.Base58()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region SignatureWithPublicKey ///////////////////////////////////////////////////////////////////////////////////////

// SignatureWithPublicKey is a combination of a PublicKey and a Signature that is required to perform operations like
// Signature- and PublicKey-aggregations.
type SignatureWithPublicKey struct {
	PublicKey PublicKey
	Signature Signature
}

// NewSignatureWithPublicKey is the constructor for SignatureWithPublicKey objects.
func NewSignatureWithPublicKey(publicKey PublicKey, signature Signature) SignatureWithPublicKey {
	return SignatureWithPublicKey{
		PublicKey: publicKey,
		Signature: signature,
	}
}

// SignatureWithPublicKeyFromBytes unmarshals a SignatureWithPublicKey from a sequence of bytes.
func SignatureWithPublicKeyFromBytes(b []byte) (signatureWithPublicKey SignatureWithPublicKey, err error) {
	reader := bytes.NewReader(b)

	if signatureWithPublicKey, err = SignatureWithPublicKeyFromReader(reader); err != nil {
		err = ierrors.Wrap(err, "failed to parse SignatureWithPublicKey from Reader")
		return
	}

	return
}

// SignatureWithPublicKeyFromBase58EncodedString creates a SignatureWithPublicKey from a base58 encoded string.
func SignatureWithPublicKeyFromBase58EncodedString(base58EncodedString string) (signatureWithPublicKey SignatureWithPublicKey, err error) {
	bytes := base58.Decode(base58EncodedString)
	if len(bytes) == 0 {
		err = ierrors.Wrapf(ErrBase58DecodeFailed, "error while decoding base58 encoded SignatureWithPublicKey: %s", base58EncodedString)

		return
	}

	if signatureWithPublicKey, err = SignatureWithPublicKeyFromBytes(bytes); err != nil {
		err = ierrors.Wrap(err, "failed to parse SignatureWithPublicKey from bytes")

		return
	}

	return
}

// SignatureWithPublicKeyFromReader unmarshals a SignatureWithPublicKey using a Reader (for easier unmarshalling).
func SignatureWithPublicKeyFromReader(reader *bytes.Reader) (signatureWithPublicKey SignatureWithPublicKey, err error) {
	if signatureWithPublicKey.PublicKey, err = PublicKeyFromReader(reader); err != nil {
		err = ierrors.Wrap(err, "failed to parse PublicKey from Reader")

		return
	}

	if signatureWithPublicKey.Signature, err = SignatureFromReader(reader); err != nil {
		err = ierrors.Wrap(err, "failed to parse Signature from Reader")

		return
	}

	return
}

// IsValid returns true if the signature is correct for the given data.
func (s SignatureWithPublicKey) IsValid(data []byte) bool {
	return s.PublicKey.SignatureValid(data, s.Signature)
}

// Bytes returns the signature in bytes.
func (s SignatureWithPublicKey) Bytes() []byte {
	return slices.Concat(s.PublicKey.Bytes(), s.Signature.Bytes())
}

// Encode returns the signature in bytes.
func (s SignatureWithPublicKey) Encode() ([]byte, error) {
	return s.Bytes(), nil
}

// Encode returns the signature in bytes.
func (s *SignatureWithPublicKey) Decode(b []byte) error {
	decoded, err := SignatureWithPublicKeyFromBytes(b)
	if err != nil {
		return err
	}

	s.PublicKey = decoded.PublicKey
	s.Signature = decoded.Signature

	return nil
}

// Base58 returns a base58 encoded version of the SignatureWithPublicKey.
func (s SignatureWithPublicKey) Base58() string {
	return base58.Encode(s.Bytes())
}

// String returns a human-readable version of the SignatureWithPublicKey (base58 encoded).
func (s SignatureWithPublicKey) String() string {
	return base58.Encode(s.Bytes())
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
