// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"github.com/iotaledger/hive.go/crypto/bls"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/tbls"
)

// NOTE: this is a temporar mock of DKShare
// TODO: remove this when DKShare is migrated to iotago
type mockedDKShare struct {
	Node *mockedNode
	Log  *logger.Logger
}

var _ tcrypto.DKShare = &mockedDKShare{}

func NewMockedDKShare(node *mockedNode) *mockedDKShare {
	return &mockedDKShare{
		Node: node,
		Log:  node.Log.Named("dks"),
	}
}

func (*mockedDKShare) GetAddress() iotago.Address             { panic("TODO") }
func (*mockedDKShare) GetIndex() *uint16                      { panic("TODO") }
func (*mockedDKShare) GetN() uint16                           { panic("TODO") }
func (*mockedDKShare) GetT() uint16                           { panic("TODO") }
func (*mockedDKShare) GetNodePubKeys() []*cryptolib.PublicKey { panic("TODO") }
func (*mockedDKShare) GetSharedPublic() kyber.Point           { panic("TODO") }
func (*mockedDKShare) GetPublicShares() []kyber.Point         { panic("TODO") }
func (*mockedDKShare) SetPublicShares(ps []kyber.Point)       { panic("TODO") }
func (*mockedDKShare) GetPrivateShare() kyber.Scalar          { panic("TODO") }

func (*mockedDKShare) Bytes() []byte                                            { panic("TODO") }
func (*mockedDKShare) VerifySigShare(data []byte, sigshare tbls.SigShare) error { panic("TODO") }
func (*mockedDKShare) VerifyMasterSignature(data, signature []byte) error       { panic("TODO") }
func (*mockedDKShare) SignShare(data []byte) (tbls.SigShare, error)             { panic("TODO") }
func (*mockedDKShare) RecoverFullSignature(sigShares [][]byte, data []byte) (*bls.SignatureWithPublicKey, error) {
	panic("TODO")
}
