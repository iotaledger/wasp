// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tcrypto

import (
	"encoding/json"
	"errors"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/sign/tbls"

	"github.com/iotaledger/hive.go/crypto/bls"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/onchangemap"
	"github.com/iotaledger/wasp/packages/util"
)

// It also covers the dss.SecretShare.
type SecretShare interface {
	PriShare() *share.PriShare
	Commitments() []kyber.Point
	NodeCount() int // N -- Total number of parties.
	MaxFaulty() int // F -- Maximal number of faulty nodes.
	Threshold() int // T -- Minimal number of participants to recover a secret, T = N - F.
}

// DKShare stands for the information stored on
// a node as a result of the DKG procedure.
type DKShare interface {
	json.Marshaler
	json.Unmarshaler
	ID() *util.ComparableAddress
	Clone() onchangemap.Item[string, *util.ComparableAddress]
	Bytes() []byte
	GetAddress() iotago.Address
	GetIndex() *uint16
	GetN() uint16
	GetT() uint16
	GetNodePubKeys() []*cryptolib.PublicKey
	GetSharedPublic() *cryptolib.PublicKey
	SetPublicShares(edPublicShares []kyber.Point, blsPublicShares []kyber.Point)
	//
	// Schnorr based crypto (for L1 signatures).
	DSSPublicShares() []kyber.Point
	DSSSharedPublic() kyber.Point
	DSSSignShare(data []byte, nonce SecretShare) (*dss.PartialSig, error)
	DSSVerifySigShare(data []byte, sigshare *dss.PartialSig) error
	DSSRecoverMasterSignature(sigShares []*dss.PartialSig, data []byte, nonce SecretShare) ([]byte, error)
	DSSVerifyMasterSignature(data, signature []byte) error
	DSS() SecretShare
	//
	// BLS based crypto (for randomness only.)
	BLSThreshold() uint16
	BLSSharedPublic() kyber.Point
	BLSPublicShares() []kyber.Point
	BLSSignShare(data []byte) (tbls.SigShare, error)
	BLSVerifySigShare(data []byte, sigshare tbls.SigShare) error
	BLSRecoverMasterSignature(sigShares [][]byte, data []byte) (*bls.SignatureWithPublicKey, error)
	BLSVerifyMasterSignature(data, signature []byte) error
	BLSSign(data []byte) ([]byte, error)                        // Non-threshold variant.
	BLSVerify(signer kyber.Point, data, signature []byte) error // Non-threshold variant.
	BLSCommits() *share.PubPoly                                 // TODO: Abstract the BLS signing part to some interface and keep the keys inside.
	BLSPriShare() *share.PriShare                               // TODO: Abstract the BLS signing part to some interface and keep the keys inside.
	//
	// For tests only.
	AssignNodePubKeys(nodePubKeys []*cryptolib.PublicKey)
	AssignCommonData(dks DKShare)
	ClearCommonData()
}

var ErrDKShareNotFound = errors.New("dkShare not found")
