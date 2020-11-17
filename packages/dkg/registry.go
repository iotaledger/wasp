package dkg

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"bytes"
	"io"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/util"
	"go.dedis.ch/kyber/v3"
)

// RegistryProvider stands for a partial registry interface, needed for this package.
// It should be implemented by registry.impl
type RegistryProvider interface {
	SaveDKShare(dkShare *DKShare) error
	LoadDKShare(chainID *coretypes.ChainID) (*DKShare, error)
}

// DKShare stands for the information stored on
// a node as a result of the DKG procedure.
type DKShare struct {
	ChainID      coretypes.ChainID
	Index        uint32
	N            uint32
	T            uint32
	SharedPublic kyber.Point
	PrivateShare kyber.Scalar
	suite        kyber.Group // Transient, only needed for un-marshaling.
}

// NewDKShare creates new share of the key.
func NewDKShare(
	index uint32,
	n uint32,
	t uint32,
	sharedPublic kyber.Point,
	privateShare kyber.Scalar,
	version byte, // address.VersionED25519 = 1 | address.VersionBLS = 2
	suite kyber.Group,
) (*DKShare, error) {
	var err error
	var pubBytes []byte
	if pubBytes, err = sharedPublic.MarshalBinary(); err != nil {
		return nil, err
	}
	var sharedAddress address.Address
	switch version {
	case address.VersionED25519:
		var edPubKey ed25519.PublicKey
		if edPubKey, _, err = ed25519.PublicKeyFromBytes(pubBytes); err != nil {
			return nil, err
		}
		sharedAddress = address.FromED25519PubKey(edPubKey)
	case address.VersionBLS:
		sharedAddress = address.FromBLSPubKey(pubBytes)
	}
	var chainID coretypes.ChainID
	if chainID, err = coretypes.NewChainIDFromBytes(sharedAddress.Bytes()); err != nil {
		return nil, err
	}
	dkShare := DKShare{
		ChainID:      chainID,
		Index:        index,
		N:            n,
		T:            t,
		SharedPublic: sharedPublic,
		PrivateShare: privateShare,
	}
	return &dkShare, nil
}

// DKShareFromBytes reads DKShare from bytes.
func DKShareFromBytes(buf []byte, suite kyber.Group) (*DKShare, error) {
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
	if err = util.WriteUint32(w, s.Index); err != nil {
		return err
	}
	if err = util.WriteUint32(w, s.N); err != nil {
		return err
	}
	if err = util.WriteUint32(w, s.T); err != nil {
		return err
	}
	if err = writeMarshaled(w, s.SharedPublic); err != nil {
		return err
	}
	if err = writeMarshaled(w, s.PrivateShare); err != nil {
		return err
	}
	return nil
}

func (s *DKShare) Read(r io.Reader) error {
	var err error
	if err = s.ChainID.Read(r); err != nil {
		return err
	}
	if err = util.ReadUint32(r, &s.Index); err != nil {
		return err
	}
	if err = util.ReadUint32(r, &s.N); err != nil {
		return err
	}
	if err = util.ReadUint32(r, &s.T); err != nil {
		return err
	}
	s.SharedPublic = s.suite.Point()
	if err = readMarshaled(r, s.SharedPublic); err != nil {
		return err
	}
	s.PrivateShare = s.suite.Scalar()
	if err = readMarshaled(r, s.PrivateShare); err != nil {
		return err
	}
	return nil
}
