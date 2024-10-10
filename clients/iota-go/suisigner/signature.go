package suisigner

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"errors"
)

type Signature struct {
	*Ed25519Signature
	*Secp256k1Signature
	*Secp256r1Signature
}

const (
	SizeEd25519Signature = ed25519.PublicKeySize + ed25519.SignatureSize + 1
)

type Secp256k1Signature struct {
	Signature []byte // secp256k1.pubKey + Secp256k1Signature + 1
}

type Secp256r1Signature struct {
	Signature []byte // secp256k1.pubKey + Secp256k1Signature + 1
}

type Ed25519Signature struct {
	Signature [SizeEd25519Signature]byte
}

func (s Signature) Bytes() []byte {
	switch {
	case s.Ed25519Signature != nil:
		return s.Ed25519Signature.Signature[:]
	case s.Secp256k1Signature != nil:
		return s.Secp256k1Signature.Signature[:]
	case s.Secp256r1Signature != nil:
		return s.Secp256r1Signature.Signature[:]
	default:
		return nil
	}
}

func (s Signature) MarshalJSON() ([]byte, error) {
	switch {
	case s.Ed25519Signature != nil:
		return json.Marshal(s.Ed25519Signature.Signature[:])
	case s.Secp256k1Signature != nil:
		return json.Marshal(s.Secp256k1Signature.Signature[:])
	case s.Secp256r1Signature != nil:
		return json.Marshal(s.Secp256r1Signature.Signature[:])
	default:
		return nil, errors.New("nil signature")
	}
}

func (s *Signature) UnmarshalJSON(data []byte) error {
	var signature []byte
	err := json.Unmarshal(data, &signature)
	if err != nil {
		return err
	}
	switch signature[0] {
	case 0:
		if len(signature) != ed25519.PublicKeySize+ed25519.SignatureSize+1 {
			return errors.New("invalid ed25519 signature")
		}
		var signatureBytes [ed25519.PublicKeySize + ed25519.SignatureSize + 1]byte
		copy(signatureBytes[:], signature)
		s.Ed25519Signature = &Ed25519Signature{
			Signature: signatureBytes,
		}
	default:
		return errors.New("not supported signature")
	}
	return nil
}

func NewEd25519Signature(publicKey ed25519.PublicKey, signedMessage []byte) *Ed25519Signature {
	sigBuffer := bytes.NewBuffer([]byte{})
	sigBuffer.WriteByte(byte(KeySchemeFlagEd25519))
	sigBuffer.Write(signedMessage)
	sigBuffer.Write(publicKey)

	return &Ed25519Signature{
		Signature: [SizeEd25519Signature]byte(sigBuffer.Bytes()),
	}
}