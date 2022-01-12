// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"errors"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

type Provider func() *Impl

type NodeIdentityProvider interface {
	GetNodeIdentity() (*cryptolib.KeyPair, error)
	GetNodePublicKey() (*cryptolib.PublicKey, error)
}

// DKShareRegistryProvider stands for a partial registry interface, needed for this package.
// It should be implemented by registry.impl
type DKShareRegistryProvider interface {
	SaveDKShare(dkShare *tcrypto.DKShare) error
	LoadDKShare(sharedAddress iotago.Address) (*tcrypto.DKShare, error)
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
