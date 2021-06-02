// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"time"

	"github.com/iotaledger/wasp/packages/hashing"
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
