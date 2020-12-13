// Package dks provides functions to operate on distributed key shares.
// Distributed key shares are usually generated using a DKG procedure.
// See the `dkg` package for the generation part. This package provides
// a way to use them.
package dks

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"bytes"
	"io"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/tcrypto/tbdn"
	"github.com/iotaledger/wasp/packages/util"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"go.dedis.ch/kyber/v3/util/key"
)

type Suite interface {
	kyber.Group
	pairing.Suite
	key.Suite
}

// RegistryProvider stands for a partial registry interface, needed for this package.
// It should be implemented by registry.impl
type RegistryProvider interface {
	SaveDKShare(dkShare *DKShare) error
	LoadDKShare(chainID *coretypes.ChainID) (*DKShare, error)
}

// DKShare stands for the information stored on
// a node as a result of the DKG procedure.
type DKShare struct {
	ChainID       *coretypes.ChainID
	Address       *address.Address
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
	var err error
	//
	// Derive the ChainID.
	var pubBytes []byte
	if pubBytes, err = sharedPublic.MarshalBinary(); err != nil {
		return nil, err
	}
	var sharedAddress = address.FromBLSPubKey(pubBytes)
	var chainID coretypes.ChainID
	if chainID, err = coretypes.NewChainIDFromBytes(sharedAddress.Bytes()); err != nil {
		return nil, err
	}
	//
	// Construct the DKShare.
	dkShare := DKShare{
		ChainID:       &chainID,
		Address:       &sharedAddress,
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
func (s *DKShare) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	s.Write(&buf)
	return buf.Bytes(), nil
}

// Write returns byte representation of this struct.
func (s *DKShare) Write(w io.Writer) error {
	var err error
	if err = s.ChainID.Write(w); err != nil {
		return err
	}
	if err = util.WriteBytes16(w, s.Address.Bytes()); err != nil {
		return err
	}
	if err = util.WriteUint16(w, *s.Index); err != nil { // It must be not nil here.
		return err
	}
	if err = util.WriteUint16(w, s.N); err != nil {
		return err
	}
	if err = util.WriteUint16(w, s.T); err != nil {
		return err
	}
	if err = util.WriteMarshaled(w, s.SharedPublic); err != nil {
		return err
	}
	for i := uint16(0); i < s.N-1; i++ {
		if err = util.WriteMarshaled(w, s.PublicCommits[i]); err != nil {
			return err
		}
	}
	for i := uint16(0); i < s.N; i++ {
		if err = util.WriteMarshaled(w, s.PublicShares[i]); err != nil {
			return err
		}
	}
	if err = util.WriteMarshaled(w, s.PrivateShare); err != nil {
		return err
	}
	return nil
}

func (s *DKShare) Read(r io.Reader) error {
	var err error
	chainID := coretypes.NilChainID // Make a copy.
	if err = chainID.Read(r); err != nil {
		return err
	}
	s.ChainID = &chainID
	var addr address.Address
	var addrBytes []byte
	if addrBytes, err = util.ReadBytes16(r); err != nil {
		return err
	}
	if addr, _, err = address.FromBytes(addrBytes); err != nil {
		return err
	}
	s.Address = &addr
	var index uint16
	if err = util.ReadUint16(r, &index); err != nil {
		return err
	}
	s.Index = &index
	if err = util.ReadUint16(r, &s.N); err != nil {
		return err
	}
	if err = util.ReadUint16(r, &s.T); err != nil {
		return err
	}
	s.SharedPublic = s.suite.Point()
	if err = util.ReadMarshaled(r, s.SharedPublic); err != nil {
		return err
	}
	s.PublicCommits = make([]kyber.Point, s.N-1)
	for i := uint16(0); i < s.N-1; i++ {
		s.PublicCommits[i] = s.suite.Point()
		if err = util.ReadMarshaled(r, s.PublicCommits[i]); err != nil {
			return err
		}
	}
	s.PublicShares = make([]kyber.Point, s.N)
	for i := uint16(0); i < s.N; i++ {
		s.PublicShares[i] = s.suite.Point()
		if err = util.ReadMarshaled(r, s.PublicShares[i]); err != nil {
			return err
		}
	}
	s.PrivateShare = s.suite.Scalar()
	if err = util.ReadMarshaled(r, s.PrivateShare); err != nil {
		return err
	}
	return nil
}

// SignShare signs the data with the own key share.
// returns SigShare, which contains signature and the index
func (s *DKShare) SignShare(data []byte) (tbdn.SigShare, error) {
	priShare := share.PriShare{
		I: int(*s.Index),
		V: s.PrivateShare,
	}
	return tbdn.Sign(s.suite, &priShare, data)
}

// VerifySigShare verifies the signature of a particular share.
func (s *DKShare) VerifySigShare(data []byte, sigshare tbdn.SigShare) error {
	idx, err := sigshare.Index()
	if err != nil || idx >= int(s.N) || idx < 0 {
		return err
	}
	return bdn.Verify(s.suite, s.PublicShares[idx], data, sigshare.Value())
}

// RecoverFullSignature generates (recovers) master signature from partial sigshares.
// returns signature as defined in the value Tangle
func (s *DKShare) RecoverFullSignature(sigShares [][]byte, data []byte) (signaturescheme.Signature, error) {
	pubPoly := share.NewPubPoly(s.suite, nil, s.PublicCommits)
	recoveredSignature, err := tbdn.Recover(s.suite, pubPoly, data, sigShares, int(s.T), int(s.N))
	if err != nil {
		return nil, err
	}
	pubKeyBin, err := s.SharedPublic.MarshalBinary()
	if err != nil {
		return nil, err
	}
	finalSignature := signaturescheme.NewBLSSignature(pubKeyBin, recoveredSignature)

	if finalSignature.Address().String() != s.Address.String() {
		panic("finalSignature.ChainID() != op.dkShare.ChainID")
	}
	return finalSignature, nil
}
