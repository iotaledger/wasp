package cryptolib

import (
	"crypto/ed25519"
	"crypto/sha512"
	"io"

	// We need to use this package to have access to low-level edwards25519 operations.
	//
	// Excerpt from the docs:
	// https://pkg.go.dev/crypto/ed25519/internal/edwards25519?utm_source=godoc
	//
	// However, developers who do need to interact with low-level edwards25519
	// operations can use filippo.io/edwards25519,
	// an extended version of this package repackaged as an importable module.
	"filippo.io/edwards25519"

	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

const SignatureSize = ed25519.SignatureSize

// Signature defines an Ed25519 signature.
type Signature struct {
	// The signature schema (0 == ED25519)
	signatureScheme byte

	// The public key used to verify the given signature.
	publicKey *PublicKey
	// The signature.
	signature [SignatureSize]byte
}

var (
	_ rwutil.IoReader = &Signature{}
	_ rwutil.IoWriter = &Signature{}
)

func NewEmptySignature() *Signature {
	return &Signature{}
}

func NewDummySignature(publicKey *PublicKey) *Signature {
	return &Signature{publicKey: publicKey}
}

func NewSignature(publicKey *PublicKey, signature []byte) *Signature {
	result := Signature{
		publicKey: publicKey,
	}
	// copy(result.publicKey[:], publicKey.AsBytes())
	copy(result.signature[:], signature)
	return &result
}

// TODO: is it really needed?
func (s *Signature) GetPublicKey() *PublicKey {
	return s.publicKey
}

/*// TODO: is it really needed?
func (s *Signature) GetSignature() [SignatureSize]byte {
	return s.signature
}*/

// Validate reports whether sig is a valid signature of message by publicKey.
// It uses precisely-specified validation criteria (ZIP 215) suitable for use in consensus-critical contexts.
// It is compatible with the particular validation rules around edge cases described in IOTA protocol RFC-0028.
func (s *Signature) Validate(message []byte) bool {
	publicKey := s.publicKey.AsBytes()
	if s.signature[63]&224 != 0 {
		return false
	}

	// ZIP215: this works because SetBytes does not check that encodings are canonical
	A, err := new(edwards25519.Point).SetBytes(publicKey)
	if err != nil {
		return false
	}
	A.Negate(A)

	h := sha512.New()
	h.Write(s.signature[:32])
	h.Write(publicKey)
	h.Write(message)
	var digest [64]byte
	h.Sum(digest[:0])

	hReduced, err := new(edwards25519.Scalar).SetUniformBytes(digest[:])
	if err != nil {
		panic(err)
	}

	// ZIP215: this works because SetBytes does not check that encodings are canonical
	checkR, err := new(edwards25519.Point).SetBytes(s.signature[:32])
	if err != nil {
		return false
	}

	// https://tools.ietf.org/html/rfc8032#section-5.1.7 requires that s be in
	// the range [0, order) in order to prevent signature malleability
	scalar, err := new(edwards25519.Scalar).SetCanonicalBytes(s.signature[32:])
	if err != nil {
		return false
	}

	R := new(edwards25519.Point).VarTimeDoubleScalarBaseMult(hReduced, A, scalar)

	// ZIP215: We want to check [8](R - checkR) == 0
	p := new(edwards25519.Point).Subtract(R, checkR)     // p = R - checkR
	p.Add(p, p)                                          // p = [2]p
	p.Add(p, p)                                          // p = [4]p
	p.Add(p, p)                                          // p = [8]p
	return p.Equal(edwards25519.NewIdentityPoint()) == 1 // p == 0
}

func (s *Signature) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	s.signatureScheme = rr.ReadByte()
	s.publicKey = NewEmptyPublicKey()
	rr.Read(s.publicKey)
	signature := rr.ReadBytes()
	copy(s.signature[:], signature)
	return rr.Err
}

func (s *Signature) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(s.signatureScheme)
	ww.Write(s.publicKey)
	ww.WriteBytes(s.signature[:])
	return ww.Err
}

func (s *Signature) AsSuiSignature() *iotasigner.Signature {
	result := &iotasigner.Signature{
		Ed25519Signature: iotasigner.NewEd25519Signature(s.publicKey.AsBytes(), s.signature[:]),
	}
	return result
}
