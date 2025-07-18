package crypto

import (
	"errors"
	"fmt"
	"io"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"golang.org/x/crypto/chacha20poly1305"

	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

// errors returned by the package
var (
	ErrNotCanonical       = errors.New("not canonical")
	ErrSmallOrder         = errors.New("small order")
	ErrInvalidInputLength = errors.New("invalid input length")
	ErrDecryptionFailed   = errors.New("decryption failed")
	ErrVerificationFailed = errors.New("verification failed")
)

const (
	// AEADKeySize denotes the size of the AEAD keys in bytes.
	AEADKeySize = chacha20poly1305.KeySize
	// AEADOverhead denotes the number of additional bytes required.
	AEADOverhead = chacha20poly1305.Overhead
)

// Share represents a private share of the secret.
type Share = share.PriShare

// Commits represents the Feldman VSS commitments.
type Commits []kyber.Point

// MarshalTo encodes the receiver into binary and writes it to w.
func (c Commits) MarshalTo(w io.Writer) (int, error) {
	ww := rwutil.NewWriter(w)
	counter := rwutil.NewWriteCounter(ww)
	for _, p := range c {
		ww.WriteFromFunc(p.MarshalTo)
	}
	return counter.Count(), ww.Err
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (c Commits) MarshalBinary() ([]byte, error) {
	ww := rwutil.NewBytesWriter()
	ww.WriteFromFunc(c.MarshalTo)
	return ww.Bytes(), ww.Err
}

// SecretLen returns the length of Secret in bytes.
func SecretLen(g kyber.Group) int { return g.PointLen() }

// Secret computes and returns the shared ephemeral secret.
func Secret(g kyber.Group, remotePublic kyber.Point, ownPrivate kyber.Scalar) []byte {
	dh := g.Point().Mul(ownPrivate, remotePublic)
	data, err := dh.MarshalBinary()
	if err != nil {
		panic(fmt.Errorf("cannot marshal a shared secret: %v", err))
	}
	return data
}

// ShareLen returns the length of an encrypted share in bytes.
func ShareLen(g kyber.Group) int { return g.ScalarLen() + AEADOverhead }

// DecryptShare decrypts and validates the encrypted share with the given index using the given secret.
// An error is returned if no valid share could be decrypted.
func DecryptShare(g kyber.Group, deal *Deal, index int, secret []byte) (*share.PriShare, error) {
	if len(secret) != SecretLen(g) {
		return nil, ErrInvalidInputLength
	}

	salt, _ := deal.Commits.MarshalBinary()
	aead := newAEAD(secret, salt, contextInfo(index))
	v := g.Scalar()
	if err := decryptScalar(v, aead, deal.Shares[index]); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrDecryptionFailed, err.Error())
	}
	s := &share.PriShare{I: index, V: v}
	if !share.NewPubPoly(g, nil, deal.Commits).Check(s) {
		return nil, ErrVerificationFailed
	}
	return s, nil
}

// InterpolateShare interpolates a new private share for index i.
func InterpolateShare(g kyber.Group, shares []*Share, n, i int) (*Share, error) {
	poly, err := share.RecoverPriPoly(g, shares, threshold(n), n)
	if err != nil {
		return nil, err
	}
	return poly.Eval(i), nil
}
