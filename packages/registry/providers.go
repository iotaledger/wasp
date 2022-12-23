// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

type NodeIdentityProvider interface {
	NodeIdentity() *cryptolib.KeyPair
	NodePublicKey() *cryptolib.PublicKey
}

type DKShareRegistryProvider interface {
	SaveDKShare(dkShare tcrypto.DKShare) error
	LoadDKShare(sharedAddress iotago.Address) (tcrypto.DKShare, error)
}

type ChainRecordRegistryProvider interface {
	Events() *ChainRecordRegistryEvents
	ChainRecord(chainID isc.ChainID) (*ChainRecord, error)
	ChainRecords() ([]*ChainRecord, error)
	ForEachActiveChainRecord(consumer func(*ChainRecord) bool) error
	AddChainRecord(chainRecord *ChainRecord) error
	DeleteChainRecord(chainID isc.ChainID) error
	UpdateChainRecord(chainID isc.ChainID, f func(*ChainRecord) bool) (*ChainRecord, error)
	ActivateChainRecord(chainID isc.ChainID) (*ChainRecord, error)
	DeactivateChainRecord(chainID isc.ChainID) (*ChainRecord, error)
}

type TrustedPeersRegistryProvider interface {
	IsTrustedPeer(pubKey *cryptolib.PublicKey) error
	TrustPeer(pubKey *cryptolib.PublicKey, netID string) (*peering.TrustedPeer, error)
	DistrustPeer(pubKey *cryptolib.PublicKey) (*peering.TrustedPeer, error)
	TrustedPeers() ([]*peering.TrustedPeer, error)
}
