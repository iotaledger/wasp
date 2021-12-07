// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tcrypto

import (
	"bytes"
	"io"

	"github.com/iotaledger/hive.go/crypto/bls"
	iotago "github.com/iotaledger/iota.go/v3"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"go.dedis.ch/kyber/v3/sign/tbls"
	"golang.org/x/xerrors"
)

// DKShare stands for the information stored on
// a node as a result of the DKG procedure.
type DKShare struct {
	Address       iotago.Address
	Index         *uint16 // nil, if the current node is not a member of a group sharing the key.
	N             uint16
	T             uint16
	SharedPublic  kyber.Point
	PublicCommits []kyber.Point
	PublicShares  []kyber.Point
	PrivateShare  kyber.Scalar
	suite         Suite // Transient, only needed for un-marshaling.
}

// NewDKShare creates new share of the key.
func NewDKShare(
	index uint16,
	n uint16,
	t uint16,
	sharedPublic kyber.Point,
	publicCommits []kyber.Point,
	publicShares []kyber.Point,
	privateShare kyber.Scalar,
) (*DKShare, error) {
	//var err error
	//
	// Derive the ChainID.
	//var pubBytes []byte
	//if pubBytes, err = sharedPublic.MarshalBinary(); err != nil {
	//	return nil, err
	//}
	// TODO bls has been removed from IOTAGO, should this be moved to ed25519?
	//sharedAddress := iotago.BLSAddressFromPubKey(bls.PublicKeyFromBytes(pubBytes))
	var sharedAddress iotago.Address // TODO not correct

	//
	// Construct the DKShare.
	dkShare := DKShare{
		Address:       sharedAddress,
		Index:         &index,
		N:             n,
		T:             t,
		SharedPublic:  sharedPublic,
		PublicCommits: publicCommits,
		PublicShares:  publicShares,
		PrivateShare:  privateShare,
		// NOTE: suite is not stored here.
	}
	return &dkShare, nil
}

// DKShareFromBytes reads DKShare from bytes.
func DKShareFromBytes(buf []byte, suite Suite) (*DKShare, error) {
	r := bytes.NewReader(buf)
	s := DKShare{suite: suite}
	if err := s.Read(r); err != nil {
		return nil, err
	}
	return &s, nil
}

// Bytes returns byte representation of the share.
func (s *DKShare) Bytes() []byte {
	var buf bytes.Buffer
	if err := s.Write(&buf); err != nil {
		panic(xerrors.Errorf("DKShare.Bytes: %w", err))
	}
	return buf.Bytes()
}

// Write returns byte representation of this struct.
//nolint:gocritic
func (s *DKShare) Write(w io.Writer) error {
	panic("TODO implement")
	// var err error
	// if _, err = w.Write(s.Address.Bytes()); err != nil {
	// 	return err
	// }
	// if err = util.WriteUint16(w, *s.Index); err != nil { // It must be not nil here.
	// 	return err
	// }
	// if err = util.WriteUint16(w, s.N); err != nil {
	// 	return err
	// }
	// if err = util.WriteUint16(w, s.T); err != nil {
	// 	return err
	// }
	// if err = util.WriteMarshaled(w, s.SharedPublic); err != nil {
	// 	return err
	// }
	// if err = util.WriteUint16(w, uint16(len(s.PublicCommits))); err != nil {
	// 	return err
	// }
	// for i := 0; i < len(s.PublicCommits); i++ {
	// 	if err = util.WriteMarshaled(w, s.PublicCommits[i]); err != nil {
	// 		return err
	// 	}
	// }
	// if err = util.WriteUint16(w, uint16(len(s.PublicShares))); err != nil {
	// 	return err
	// }
	// for i := 0; i < len(s.PublicShares); i++ {
	// 	if err = util.WriteMarshaled(w, s.PublicShares[i]); err != nil {
	// 		return err
	// 	}
	// }
	// return util.WriteMarshaled(w, s.PrivateShare)
}

//nolint:gocritic
func (s *DKShare) Read(r io.Reader) error {
	panic("TODO implement")

	// var err error
	// var addrBytes [iotago.Ed25519AddressBytesLength]byte
	// var arrLen uint16
	// if n, err := r.Read(addrBytes[:]); err != nil || n != iotago.Ed25519AddressBytesLength {
	// 	return err
	// }
	// if s.Address, _, err = iotago.AddressFromBytes(addrBytes[:]); err != nil {
	// 	return err
	// }
	// var index uint16
	// if err = util.ReadUint16(r, &index); err != nil {
	// 	return err
	// }
	// s.Index = &index
	// if err = util.ReadUint16(r, &s.N); err != nil {
	// 	return err
	// }
	// if err = util.ReadUint16(r, &s.T); err != nil {
	// 	return err
	// }
	// s.SharedPublic = s.suite.G2().Point()
	// if err = util.ReadMarshaled(r, s.SharedPublic); err != nil {
	// 	return err
	// }
	// //
	// // PublicCommits
	// if err = util.ReadUint16(r, &arrLen); err != nil {
	// 	return err
	// }
	// s.PublicCommits = make([]kyber.Point, arrLen)
	// for i := uint16(0); i < arrLen; i++ {
	// 	s.PublicCommits[i] = s.suite.G2().Point()
	// 	if err = util.ReadMarshaled(r, s.PublicCommits[i]); err != nil {
	// 		return err
	// 	}
	// }
	// //
	// // PublicShares
	// if err = util.ReadUint16(r, &arrLen); err != nil {
	// 	return err
	// }
	// s.PublicShares = make([]kyber.Point, arrLen)
	// for i := uint16(0); i < arrLen; i++ {
	// 	s.PublicShares[i] = s.suite.G2().Point()
	// 	if err = util.ReadMarshaled(r, s.PublicShares[i]); err != nil {
	// 		return err
	// 	}
	// }
	// //
	// // Private share.
	// s.PrivateShare = s.suite.G2().Scalar()
	// return util.ReadMarshaled(r, s.PrivateShare)
}

// SignShare signs the data with the own key share.
// returns SigShare, which contains signature and the index
func (s *DKShare) SignShare(data []byte) (tbls.SigShare, error) {
	priShare := share.PriShare{
		I: int(*s.Index),
		V: s.PrivateShare,
	}
	return tbls.Sign(s.suite, &priShare, data)
}

// VerifySigShare verifies the signature of a particular share.
func (s *DKShare) VerifySigShare(data []byte, sigshare tbls.SigShare) error {
	idx, err := sigshare.Index()
	if err != nil || idx >= int(s.N) || idx < 0 {
		return err
	}
	return bdn.Verify(s.suite, s.PublicShares[idx], data, sigshare.Value())
}

// VerifyOwnSigShare is only used for assertions
// NOTE: Not used.
func (s *DKShare) VerifyOwnSigShare(data []byte, sigshare tbls.SigShare) error {
	idx, err := sigshare.Index()
	if err != nil || uint16(idx) != *s.Index {
		return err
	}
	return bdn.Verify(s.suite, s.PublicShares[idx], data, sigshare[2:])
}

// VerifyMasterSignature checks signature against master public key
// NOTE: Not used.
func (s *DKShare) VerifyMasterSignature(data, signature []byte) error {
	return bdn.Verify(s.suite, s.SharedPublic, data, signature)
}

// RecoverFullSignature generates (recovers) master signature from partial sigshares.
// returns signature as defined in the value Tangle
func (s *DKShare) RecoverFullSignature(sigShares [][]byte, data []byte) (*bls.SignatureWithPublicKey, error) {
	var err error
	var recoveredSignatureBin []byte
	if s.N > 1 {
		pubPoly := share.NewPubPoly(s.suite, nil, s.PublicCommits)
		recoveredSignatureBin, err = tbls.Recover(s.suite, pubPoly, data, sigShares, int(s.T), int(s.N))
		if err != nil {
			return nil, err
		}
	} else {
		singleSigShare := tbls.SigShare(sigShares[0])
		recoveredSignatureBin = singleSigShare.Value()
	}
	sig, _, err := bls.SignatureFromBytes(recoveredSignatureBin)
	if err != nil {
		return nil, err
	}
	ret := bls.NewSignatureWithPublicKey(bls.PublicKey{Point: s.SharedPublic}, sig)
	return &ret, nil
}
