// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tcrypto

import (
	"github.com/iotaledger/hive.go/crypto/bls"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/sign/tbls"
)

// DKShare stands for the information stored on
// a node as a result of the DKG procedure.
type DKShare interface {
	Bytes() []byte
	GetAddress() iotago.Address
	GetIndex() *uint16
	GetN() uint16
	GetT() uint16
	GetNodePubKeys() []*cryptolib.PublicKey
	SetPublicShares(edPublicShares []kyber.Point, blsPublicShares []kyber.Point)
	//
	// Schnorr based crypto (for L1 signatures).
	GetSharedPublic() kyber.Point
	GetPublicShares() []kyber.Point
	// GetPrivateShare() kyber.Scalar // TODO: remove it.
	SignShare(data []byte) (*dss.PartialSig, error)
	VerifySigShare(data []byte, sigshare *dss.PartialSig) error
	RecoverMasterSignature(sigShares []*dss.PartialSig, data []byte) ([]byte, error)
	VerifyMasterSignature(data, signature []byte) error
	//
	// BLS based crypto (for randomness only.)
	BlsSharedPublic() kyber.Point
	BlsPublicShares() []kyber.Point
	BlsSignShare(data []byte) (tbls.SigShare, error)
	BlsVerifySigShare(data []byte, sigshare tbls.SigShare) error
	BlsRecoverMasterSignature(sigShares [][]byte, data []byte) (*bls.SignatureWithPublicKey, error)
	BlsVerifyMasterSignature(data, signature []byte) error
	//
	// For tests only.
	AssignNodePubKeys(nodePubKeys []*cryptolib.PublicKey)
	AssignCommonData(dks DKShare)
	ClearCommonData()
}
