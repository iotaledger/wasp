// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"errors"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

type Provider func() *Impl

// BlobCache is an access to the cache of big binary objects
type BlobCache interface { // TODO should blobcache be removed ?
	GetBlob(h hashing.HashValue) ([]byte, bool, error)
	HasBlob(h hashing.HashValue) (bool, error)
	// PutBlob ttl s TimeToLive, expiration time in Unix nanoseconds
	PutBlob(data []byte, ttl ...time.Duration) (hashing.HashValue, error)
}

type NodeIdentityProvider interface {
	GetNodeIdentity() (*ed25519.KeyPair, error)
	GetNodePublicKey() (*ed25519.PublicKey, error)
}

// DKShareRegistryProvider stands for a partial registry interface, needed for this package.
// It should be implemented by registry.impl
type DKShareRegistryProvider interface {
	SaveDKShare(dkShare *tcrypto.DKShare) error
	LoadDKShare(sharedAddress ledgerstate.Address) (*tcrypto.DKShare, error)
}

var ErrDKShareNotFound = errors.New("dkShare not found")

// ChainRecordRegistryProvider stands for a partial registry interface, needed for this package.
type ChainRecordRegistryProvider interface {
	GetChainRecordByChainID(chainID *iscp.ChainID) (*ChainRecord, error)
	GetChainRecords() ([]*ChainRecord, error)
	UpdateChainRecord(chainID *iscp.ChainID, f func(*ChainRecord) bool) (*ChainRecord, error)
	ActivateChainRecord(chainID *iscp.ChainID) (*ChainRecord, error)
	DeactivateChainRecord(chainID *iscp.ChainID) (*ChainRecord, error)
}
