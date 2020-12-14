// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tcrypto

import "github.com/iotaledger/wasp/packages/coretypes"

// RegistryProvider stands for a partial registry interface, needed for this package.
// It should be implemented by registry.impl
type RegistryProvider interface {
	SaveDKShare(dkShare *DKShare) error
	LoadDKShare(chainID *coretypes.ChainID) (*DKShare, error)
}
