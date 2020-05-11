// Package kyber.sign.tbls adjusted for bdn
// All calls to bls were replaced to same calls from tbdn
// It doesn't make much difference, because all functions used here from bdn
// are equivalent to those  from bls
// However, for the sake of correctness

package tbdn

import (
	"bytes"
	"encoding/binary"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/bdn"
)

// sigShare encodes a threshold BDN signature share Si = i || v where the 2-byte
// big-endian value i corresponds to the share's index and v represents the
// share's value. The signature share Si is a point on curve G1.
type SigShare []byte

// Index returns the index i of the TBDN share Si.
func (s SigShare) Index() (int, error) {
	var index uint16
	buf := bytes.NewReader(s)
	err := binary.Read(buf, binary.BigEndian, &index)
	if err != nil {
		return -1, err
	}
	return int(index), nil
}

// Value returns the value v of the TBDN share Si.
func (s *SigShare) Value() []byte {
	return []byte(*s)[2:]
}

// Sign creates a threshold BDN signature Si = xi * H(m) on the given message m
// using the provided secret key share xi.
func Sign(suite pairing.Suite, private *share.PriShare, msg []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, uint16(private.I)); err != nil {
		return nil, err
	}
	s, err := bdn.Sign(suite, private.V, msg)
	if err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// VerifySigShare checks the given threshold BDN signature Si on the message m using
// the public key share Xi that is associated to the secret key share xi. This
// public key share Xi can be computed by evaluating the public sharing
// polynomial at the share's index i.
func Verify(suite pairing.Suite, public *share.PubPoly, msg, sig []byte) error {
	s := SigShare(sig)
	i, err := s.Index()
	if err != nil {
		return err
	}
	return bdn.Verify(suite, public.Eval(i).V, msg, s.Value())
}

// Recover reconstructs the full BDN signature S = x * H(m) from a threshold t
// of signature shares Si using Lagrange interpolation. The full signature S
// can be verified through the regular BDN verification routine using the
// shared public key X. The shared public key can be computed by evaluating the
// public sharing polynomial at index 0.
func Recover(suite pairing.Suite, public *share.PubPoly, msg []byte, sigs [][]byte, t, n int) ([]byte, error) {
	pubShares := make([]*share.PubShare, 0)
	for _, sig := range sigs {
		s := SigShare(sig)
		i, err := s.Index()
		if err != nil {
			return nil, err
		}
		if err = bdn.Verify(suite, public.Eval(i).V, msg, s.Value()); err != nil {
			return nil, err
		}
		point := suite.G1().Point()
		if err := point.UnmarshalBinary(s.Value()); err != nil {
			return nil, err
		}
		pubShares = append(pubShares, &share.PubShare{I: i, V: point})
		if len(pubShares) >= t {
			break
		}
	}
	commit, err := share.RecoverCommit(suite.G1(), pubShares, t, n)
	if err != nil {
		return nil, err
	}
	sig, err := commit.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return sig, nil
}
