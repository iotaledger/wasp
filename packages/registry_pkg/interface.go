// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry_pkg

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/registry_pkg/chainrecord"
)

// ChainRecordRegistryProvider stands for a partial registry interface, needed for this package.
// It should be implemented by in the chainrecord package
type ChainRecordRegistryProvider interface {
	GetChainRecordByChainID(chainID *coretypes.ChainID) (*chainrecord.ChainRecord, error)
	GetChainRecords() ([]*chainrecord.ChainRecord, error)
	UpdateChainRecord(chainID *coretypes.ChainID, f func(*chainrecord.ChainRecord) bool) (*chainrecord.ChainRecord, error)
	ActivateChainRecord(chainID *coretypes.ChainID) (*chainrecord.ChainRecord, error)
	DeactivateChainRecord(chainID *coretypes.ChainID) (*chainrecord.ChainRecord, error)
	SaveChainRecord(rec *chainrecord.ChainRecord) error
}
