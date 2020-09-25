// wrapper package for BLS threshold cryptography used in the Wasp node
// TODO DKG protocol must be rewritten because currently it is not 100% secure
package tcrypto

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/wasp/packages/tcrypto/tbdn"
	"github.com/pkg/errors"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/bdn"
)

// DKShare represents distributed key share for (T,N) threshold signatures based on BLS
// Structure is a partial share owned by the node to participate in the
// committee. The only 'priKey' part is secret, the rest is public
type DKShare struct {
	// interface for the BN256 bilinear pairing for the underlying BLS cryptography
	Suite *bn256.Suite
	// size of the committee
	N uint16
	// threshold factor, 1 <= T <= N
	T uint16
	// peer index of the owner of this share in he committee
	// all N peers are indexed 0..N-1
	Index uint16
	// BLS address represented by the set of shares. It is used as a key to find the DKShare
	// all nodes in the committee have DKShare records with same address
	// Addresses is blake2 hash of master public key prefixed with one byte of signature type
	Address *address.Address
	// partial public keys of all committee nodes for this DKS
	// may be used to identify and authenticate individual committee node
	PubKeys []kyber.Point // all public shares by peers
	// must be same same as PubKeys[Index]
	// TODO cleanup. remove redundant information, plus tests
	PubKeyOwn kyber.Point
	// public polynomial, recovered form public keys according to BDN
	PubPoly *share.PubPoly
	// public key from own private key. It corresponds to address
	PubKeyMaster kyber.Point
	// secret partial private key
	// it is a sum of private shares, generated during DKG
	// partial private key not known to anyone
	// TODO however owner can reconstruct master secret from the information gathered during the DKG
	// make that optional during DKG
	priKey kyber.Scalar
	// temporary fields used during DKG process
	// not used after
	// TODO refactor during cleanup, remove tmp fields from the permanent structure
	Aggregated bool              // true after DKG
	Committed  bool              // true after DKG
	PriShares  []*share.PriShare // nil after DKG
}

func ValidateDKSParams(t, n, index uint16) error {
	if n < 2 || t < 1 || t > n || index < 0 || index >= n {
		return fmt.Errorf("wrong DKG parameters: N = %d, T = %d, Index = %d", n, t, index)
	}
	if t < n/2+1 {
		// quorum t must be larger than half size in order to avoid more than one valid quorum
		// in committee.
		// For the DKG itself it is enough to have t >= 2
		return fmt.Errorf("wrong DKG parameters: for N = %d value T must be at least %d", n, n/2+1)
	}
	return nil
}

// NewRndDKShare creates empty structure
func NewRndDKShare(t, n, index uint16) (*DKShare, error) {
	if err := ValidateDKSParams(t, n, index); err != nil {
		return nil, err
	}
	suite := bn256.NewSuite()
	// create seed secret
	secret := suite.G1().Scalar().Pick(suite.RandomStream())
	// create random polynomial of degree t
	priPoly := share.NewPriPoly(suite.G2(), int(t), secret, suite.RandomStream())
	// create private shares of the random polynomial
	// with index n corresponds to p(n+1)
	shares := priPoly.Shares(int(n))
	ret := &DKShare{
		Suite:     suite,
		N:         n,
		T:         t,
		Index:     index,
		PriShares: shares,
	}
	return ret, nil
}

// AggregateDKS is a call in DKG process
func (ks *DKShare) AggregateDKS(priShares []kyber.Scalar) error {
	if ks.Aggregated {
		return errors.New("already Aggregated")
	}
	// aggregate (add up) secret shares
	ks.priKey = ks.Suite.G2().Scalar().Zero()
	for i, pshare := range priShares {
		if uint16(i) == ks.Index {
			ks.priKey = ks.priKey.Add(ks.priKey, ks.PriShares[ks.Index].V)
			continue
		}
		ks.priKey = ks.priKey.Add(ks.priKey, pshare)
	}
	// calculate own public key
	ks.PubKeyOwn = ks.Suite.G2().Point().Mul(ks.priKey, nil)
	ks.Aggregated = true
	return nil
}

// FinalizeDKS is a call in DKG process
func (ks *DKShare) FinalizeDKS(pubKeys []kyber.Point) error {
	if ks.Committed {
		return errors.New("already Committed")
	}
	ks.PubKeys = pubKeys
	var err error
	ks.PubPoly, err = RecoverPubPoly(ks.Suite, ks.PubKeys, ks.T, ks.N)
	if err != nil {
		return err
	}
	pubKeyMaster := ks.PubPoly.Commit()
	pubKeyBin, err := pubKeyMaster.MarshalBinary()
	if err != nil {
		return err
	}
	// calculate address, the permanent key ID
	a := address.FromBLSPubKey(pubKeyBin)
	ks.Address = &a

	ks.PriShares = nil // not needed anymore
	ks.Committed = true
	return nil
}

// SignShare signs the data with the own key share.
// returns SigShare, which contains signature and the index
func (ks *DKShare) SignShare(data []byte) (tbdn.SigShare, error) {
	priShare := share.PriShare{
		I: int(ks.Index),
		V: ks.priKey,
	}
	return tbdn.Sign(ks.Suite, &priShare, data)
}

// VerifyOwnSigShare is only used for assertions
func (ks *DKShare) VerifyOwnSigShare(data []byte, sigshare tbdn.SigShare) error {
	if !ks.Committed {
		return errors.New("key set is not committed")
	}
	idx, err := sigshare.Index()
	if err != nil || uint16(idx) != ks.Index {
		return err
	}
	return bdn.Verify(ks.Suite, ks.PubKeyOwn, data, sigshare[2:])
}

// VerifySigShare checks if partial signature (sigshare) of the data is valid
func (ks *DKShare) VerifySigShare(data []byte, sigshare tbdn.SigShare) error {
	if !ks.Committed {
		return errors.New("key set is not committed")
	}
	idx, err := sigshare.Index()
	if err != nil || idx >= int(ks.N) || idx < 0 {
		return err
	}
	return bdn.Verify(ks.Suite, ks.PubKeys[idx], data, sigshare.Value())
}

// VerifyMasterSignature checks signature against master public key
func (ks *DKShare) VerifyMasterSignature(data []byte, signature []byte) error {
	if !ks.Committed {
		return errors.New("key set is not committed")
	}
	return bdn.Verify(ks.Suite, ks.PubKeyMaster, data, signature)
}

var suiteLoc = bn256.NewSuite()

// VerifyWithPublicKey checks signature against arbitrary public key
func VerifyWithPublicKey(data, signature, pubKeyBin []byte) error {
	pubKey := suiteLoc.G2().Point()
	err := pubKey.UnmarshalBinary(pubKeyBin)
	if err != nil {
		return err
	}
	return bdn.Verify(suiteLoc, pubKey, data, signature)
}

// RecoverPubPoly recovers public polynomial from the partial public keys
func RecoverPubPoly(suite *bn256.Suite, pubKeys []kyber.Point, t, n uint16) (*share.PubPoly, error) {
	pubShares := make([]*share.PubShare, len(pubKeys))
	for i, v := range pubKeys {
		pubShares[i] = &share.PubShare{
			I: i,
			V: v,
		}
	}
	return share.RecoverPubPoly(suite.G2(), pubShares, int(t), int(n))
}

// RecoverFullSignature generates (recovers) master signature from partial sigshares.
// returns signature as defined in the value Tangle
func (ks *DKShare) RecoverFullSignature(sigShares [][]byte, data []byte) (signaturescheme.Signature, error) {
	recoveredSignature, err := tbdn.Recover(ks.Suite, ks.PubPoly, data, sigShares, int(ks.T), int(ks.N))
	if err != nil {
		return nil, err
	}
	pubKeyBin, err := ks.PubKeyMaster.MarshalBinary()
	if err != nil {
		return nil, err
	}
	finalSignature := signaturescheme.NewBLSSignature(pubKeyBin, recoveredSignature)

	if finalSignature.Address() != *ks.Address {
		panic("finalSignature.Addresses() != op.dkshare.Addresses")
	}
	return finalSignature, nil
}
