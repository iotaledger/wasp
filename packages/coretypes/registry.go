// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"time"

	"github.com/iotaledger/wasp/packages/registry/chainrecord"
	"github.com/iotaledger/wasp/packages/registry/committee_record"

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
// It should be implemented by in the chainrecord package
type ChainRecordRegistryProvider interface {
	GetChainRecordByChainID(chainID *ledgerstate.AliasAddress) (*chainrecord.ChainRecord, error)
	GetChainRecords() ([]*chainrecord.ChainRecord, error)
	UpdateChainRecord(chainID *ledgerstate.AliasAddress, f func(*chainrecord.ChainRecord) bool) (*chainrecord.ChainRecord, error)
	ActivateChainRecord(chainID *ledgerstate.AliasAddress) (*chainrecord.ChainRecord, error)
	DeactivateChainRecord(chainID *ledgerstate.AliasAddress) (*chainrecord.ChainRecord, error)
	SaveChainRecord(rec *chainrecord.ChainRecord) error
}
