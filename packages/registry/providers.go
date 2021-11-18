// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

type Provider func() *Impl

type NodeIdentityProvider interface {
	GetNodeIdentity() (*ed25519.KeyPair, error)
	GetNodePublicKey() (*ed25519.PublicKey, error)
}

// PeerNetworkConfigProvider access to node and chain configuration: a list of netIDs of potential peers
type PeerNetworkConfigProvider interface { // TODO: KP: Remove or redesign.
	OwnNetID() string
	PeeringPort() int
	Neighbors() []string
	String() string
}

// DKShareRegistryProvider stands for a partial registry interface, needed for this package.
// It should be implemented by registry.impl
type DKShareRegistryProvider interface {
	SaveDKShare(dkShare *tcrypto.DKShare) error
	LoadDKShare(sharedAddress ledgerstate.Address) (*tcrypto.DKShare, error)
}

type CommitteeRegistryProvider interface {
	GetCommitteeRecord(addr ledgerstate.Address) (*CommitteeRecord, error)
	SaveCommitteeRecord(rec *CommitteeRecord) error
}

// ChainRecordRegistryProvider stands for a partial registry interface, needed for this package.
type ChainRecordRegistryProvider interface {
	GetChainRecordByChainID(chainID *iscp.ChainID) (*ChainRecord, error)
	GetChainRecords() ([]*ChainRecord, error)
	UpdateChainRecord(chainID *iscp.ChainID, f func(*ChainRecord) bool) (*ChainRecord, error)
	ActivateChainRecord(chainID *iscp.ChainID) (*ChainRecord, error)
	DeactivateChainRecord(chainID *iscp.ChainID) (*ChainRecord, error)
}
