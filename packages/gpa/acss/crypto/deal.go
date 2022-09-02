package crypto

import (
	"bytes"
	"fmt"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/suites"
)

// Deal contains the information distributed by the dealer.
type Deal struct {
	Commits Commits     // Feldman VSS commitments
	PubKey  kyber.Point // ephemeral public key used to encrypt the shares
	Shares  [][]byte    // encrypted shares of all peers
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (d *Deal) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	if _, err := d.Commits.MarshalTo(&buf); err != nil {
		return nil, err
	}
	if _, err := d.PubKey.MarshalTo(&buf); err != nil {
		return nil, err
	}
	for _, s := range d.Shares {
		buf.Write(s)
	}
	return buf.Bytes(), nil
}

// DealLen returns the length of Deal in bytes.
func DealLen(g kyber.Group, n int) int {
	// t commitments, ephemeral public key, n encrypted shares
	return threshold(n)*g.PointLen() + g.PointLen() + n*ShareLen(g)
}

// NewDeal creates data necessary to distribute scalar to the peers.
// It returns the commitments C, public key pk_d and the encrypted shares Z.
func NewDeal(suite suites.Suite, pubKeys []kyber.Point, scalar kyber.Scalar) *Deal {
	var deal Deal
	n := len(pubKeys)

	// generate Feldman commitments
	poly := share.NewPriPoly(suite, threshold(n), scalar, suite.RandomStream())
	_, deal.Commits = poly.Commit(nil).Info()

	// generate ephemeral keypair
	sk := suite.Scalar().Pick(suite.RandomStream())
	deal.PubKey = suite.Point().Mul(sk, nil)

	// generate a private share for each peer
	priShares := poly.Shares(n)

	salt, err := deal.Commits.MarshalBinary()
	if err != nil {
		panic(err)
	}

	// encrypt the shares for each public key
	deal.Shares = make([][]byte, n)
	for i, pubkey := range pubKeys {
		// compute shared DH secret
		secret := Secret(suite, pubkey, sk)
		// encrypt with that secret
		aead := newAEAD(secret, salt, contextInfo(i))
		deal.Shares[i] = encryptScalar(priShares[i].V, aead)
	}
	return &deal
}

// DealUnmarshalBinary parses and verifies a deal.
// If an error is returned, the data is invalid and cannot be used by any peer.
// Otherwise, it returns the commitments C, public key pk_d and the encrypted shares.
func DealUnmarshalBinary(g kyber.Group, n int, data []byte) (*Deal, error) {
	if len(data) != DealLen(g, n) {
		return nil, ErrInvalidInputLength
	}
	var deal Deal
	buf := bytes.NewBuffer(data)

	// load all commitments
	deal.Commits = make(Commits, threshold(n))
	for i := range deal.Commits {
		c := g.Point()
		if _, err := PointUnmarshalFrom(c, buf); err != nil {
			return nil, fmt.Errorf("invalid commitment %d: %w", i, err)
		}
		deal.Commits[i] = c
	}

	// load the public key
	deal.PubKey = g.Point()
	if _, err := PointUnmarshalFrom(deal.PubKey, buf); err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	// load all encrypted shares
	shareLen := ShareLen(g)
	deal.Shares = make([][]byte, n)
	for i := range deal.Shares {
		deal.Shares[i] = make([]byte, shareLen)
		if _, err := buf.Read(deal.Shares[i]); err != nil {
			return nil, fmt.Errorf("invalid share %d: %w", i, err)
		}
	}

	return &deal, nil
}
