// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry/chainrecord"
	"github.com/iotaledger/wasp/packages/registry/committee_record"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

type Provider func() *Impl

// BlobCache is an access to the cache of big binary objects
type BlobCache interface {
	GetBlob(h hashing.HashValue) ([]byte, bool, error)
	HasBlob(h hashing.HashValue) (bool, error)
	// PutBlob ttl s TimeToLive, expiration time in Unix nanoseconds
	PutBlob(data []byte, ttl ...time.Duration) (hashing.HashValue, error)
}

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
	GetCommitteeRecord(addr ledgerstate.Address) (*committee_record.CommitteeRecord, error)
	SaveCommitteeRecord(rec *committee_record.CommitteeRecord) error
}

// ChainRecordRegistryProvider stands for a partial registry interface, needed for this package.
// It should be implemented by in the chainrecord package
type ChainRecordRegistryProvider interface {
	GetChainRecordByChainID(chainID *coretypes.ChainID) (*chainrecord.ChainRecord, error)
	GetChainRecords() ([]*chainrecord.ChainRecord, error)
	UpdateChainRecord(chainID *coretypes.ChainID, f func(*chainrecord.ChainRecord) bool) (*chainrecord.ChainRecord, error)
	ActivateChainRecord(chainID *coretypes.ChainID) (*chainrecord.ChainRecord, error)
	DeactivateChainRecord(chainID *coretypes.ChainID) (*chainrecord.ChainRecord, error)
}
