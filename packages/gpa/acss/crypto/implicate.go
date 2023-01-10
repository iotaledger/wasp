package crypto

import (
	"bytes"
	"crypto/sha512"
	"fmt"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
)

// ImplicateLen returns the length of Implicate in bytes.
func ImplicateLen(g kyber.Group) int {
	return SecretLen(g) + g.ScalarLen() + 2*g.PointLen()
}

// Implicate returns the secret as well as a proof of correctness.
// The proof is a NIZK that sk∗G=pk ∧ sk∗pk_d=secret.
func Implicate(suite suites.Suite, dealerPublic kyber.Point, ownPrivate kyber.Scalar) []byte {
	var buf bytes.Buffer
	buf.Write(Secret(suite, dealerPublic, ownPrivate))

	s, R1, R2 := dleqProof(suite, nil, dealerPublic, ownPrivate)
	if _, err := s.MarshalTo(&buf); err != nil {
		panic(err)
	}
	if _, err := R1.MarshalTo(&buf); err != nil {
		panic(err)
	}
	if _, err := R2.MarshalTo(&buf); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// CheckImplicate verifies whether data is a correct implicate from peer.
// It returns the secret which can then be used to decrypt the corresponding share.
func CheckImplicate(g kyber.Group, dealerPublic, peerPublic kyber.Point, data []byte) ([]byte, error) {
	if len(data) != ImplicateLen(g) {
		return nil, ErrInvalidInputLength
	}
	buf := bytes.NewBuffer(data)

	K := g.Point()
	if _, err := PointUnmarshalFrom(K, buf); err != nil {
		return nil, fmt.Errorf("invalid shared key: %w", err)
	}

	s := g.Scalar()
	if _, err := ScalarUnmarshalFrom(s, buf); err != nil {
		return nil, fmt.Errorf("invalid proof: %w", err)
	}
	R1 := g.Point()
	if _, err := PointUnmarshalFrom(R1, buf); err != nil {
		return nil, fmt.Errorf("invalid proof: %w", err)
	}
	R2 := g.Point()
	if _, err := PointUnmarshalFrom(R2, buf); err != nil {
		return nil, fmt.Errorf("invalid proof: %w", err)
	}

	if !dleqVerify(g, nil, dealerPublic, peerPublic, K, s, R1, R2) {
		return nil, ErrVerificationFailed
	}
	secret, _ := K.MarshalBinary()
	return secret, nil
}

func dleqProof(suite suites.Suite, G kyber.Point, H kyber.Point, secret kyber.Scalar) (kyber.Scalar, kyber.Point, kyber.Point) { //nolint:gocritic
	// compute the corresponding public keys
	P1 := suite.Point().Mul(secret, G)
	P2 := suite.Point().Mul(secret, H)

	// commitment
	r := suite.Scalar().Pick(suite.RandomStream())
	R1 := suite.Point().Mul(r, G)
	R2 := suite.Point().Mul(r, H)

	// challenge hash
	h := sha512.New()
	if _, err := P1.MarshalTo(h); err != nil {
		panic(fmt.Errorf("cannot marshal P1: %w", err))
	}
	if _, err := P2.MarshalTo(h); err != nil {
		panic(fmt.Errorf("cannot marshal P2: %w", err))
	}
	if _, err := R1.MarshalTo(h); err != nil {
		panic(fmt.Errorf("cannot marshal R1: %w", err))
	}
	if _, err := R2.MarshalTo(h); err != nil {
		panic(fmt.Errorf("cannot marshal R2: %w", err))
	}
	c := suite.Scalar().SetBytes(h.Sum(nil))

	// response
	s := suite.Scalar()
	s.Mul(secret, c).Add(s, r)

	return s, R1, R2
}

func dleqVerify(g kyber.Group, G kyber.Point, H kyber.Point, P1 kyber.Point, P2 kyber.Point, s kyber.Scalar, R1 kyber.Point, R2 kyber.Point) bool { //nolint:gocritic
	// challenge hash
	h := sha512.New()
	if _, err := P1.MarshalTo(h); err != nil {
		panic(fmt.Errorf("cannot marshal P1: %w", err))
	}
	if _, err := P2.MarshalTo(h); err != nil {
		panic(fmt.Errorf("cannot marshal P2: %w", err))
	}
	if _, err := R1.MarshalTo(h); err != nil {
		panic(fmt.Errorf("cannot marshal R1: %w", err))
	}
	if _, err := R2.MarshalTo(h); err != nil {
		panic(fmt.Errorf("cannot marshal R2: %w", err))
	}
	c := g.Scalar().SetBytes(h.Sum(nil))

	P := g.Point()

	// s * G == c * P1 + R1
	P = P.Mul(c, P1).Add(P, R1)
	if !g.Point().Mul(s, G).Equal(P) {
		return false
	}

	// s * H == c * P2 + R2
	P = P.Mul(c, P2).Add(P, R2)
	return g.Point().Mul(s, H).Equal(P)
}
