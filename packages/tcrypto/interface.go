// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tcrypto

import (
	"github.com/iotaledger/hive.go/crypto/bls"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/tbls"
)

// DKShare stands for the information stored on
// a node as a result of the DKG procedure.
type DKShare interface {
	GetAddress() iotago.Address
	GetIndex() *uint16
	GetN() uint16
	GetT() uint16
	GetNodePubKeys() []*cryptolib.PublicKey
	GetSharedPublic() kyber.Point
	GetPublicShares() []kyber.Point
	SetPublicShares(ps []kyber.Point)
	GetPrivateShare() kyber.Scalar

	Bytes() []byte
	VerifySigShare(data []byte, sigshare tbls.SigShare) error
	VerifyMasterSignature(data, signature []byte) error
	SignShare(data []byte) (tbls.SigShare, error)
	RecoverFullSignature(sigShares [][]byte, data []byte) (*bls.SignatureWithPublicKey, error)
}
