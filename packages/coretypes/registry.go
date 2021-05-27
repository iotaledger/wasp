// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"time"

	"github.com/iotaledger/wasp/packages/registry_pkg/chain_record"
	"github.com/iotaledger/wasp/packages/registry_pkg/committee_record"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/util/key"
)

// BlobCache is an access to the cache of big binary objects
type BlobCache interface {
	GetBlob(h hashing.HashValue) ([]byte, bool, error)
	HasBlob(h hashing.HashValue) (bool, error)
	// PutBlob ttl s TimeToLive, expiration time in Unix nanoseconds
	PutBlob(data []byte, ttl ...time.Duration) (hashing.HashValue, error)
}

// DKShareRegistryProvider stands for a partial registry interface, needed for this package.
// It should be implemented by registry.impl
type DKShareRegistryProvider interface {
	SaveDKShare(dkShare *tcrypto.DKShare) error
	LoadDKShare(sharedAddress ledgerstate.Address) (*tcrypto.DKShare, error)
}

type CommitteeRegistryProvider interface {
	GetCommitteeRecord(addr ledgerstate.Address) (*committee_record.CommitteeRecord, error)
	SaveCommitteeRecord(rec *committee_record.CommitteeRecord) error
}

// NodeIdentityProvider is a subset of the registry interface
// providing access to the persistent node identity information.
type NodeIdentityProvider interface {
	GetNodeIdentity() (*key.Pair, error)
	GetNodePublicKey() (kyber.Point, error)
}

// PeerNetworkConfigProvider access to node and chain configuration: a list of netIDs of potential peers
type PeerNetworkConfigProvider interface {
	OwnNetID() string
	PeeringPort() int
	Neighbors() []string
	String() string
}

// ChainRecordRegistryProvider stands for a partial registry interface, needed for this package.
// It should be implemented by in the chain_record package
type ChainRecordRegistryProvider interface {
	GetChainRecordByChainID(chainID *ledgerstate.AliasAddress) (*chain_record.ChainRecord, error)
	GetChainRecords() ([]*chain_record.ChainRecord, error)
	UpdateChainRecord(chainID *ledgerstate.AliasAddress, f func(*chain_record.ChainRecord) bool) (*chain_record.ChainRecord, error)
	ActivateChainRecord(chainID *ledgerstate.AliasAddress) (*chain_record.ChainRecord, error)
	DeactivateChainRecord(chainID *ledgerstate.AliasAddress) (*chain_record.ChainRecord, error)
	SaveChainRecord(rec *chain_record.ChainRecord) error
}
