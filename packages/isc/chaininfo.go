// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import (
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// ChainInfo is an API structure containing the main parameters of the chain
type ChainInfo struct {
	ChainID        ChainID
	ChainOwnerID   AgentID
	GasFeePolicy   *gas.FeePolicy
	GasLimits      *gas.Limits
	CustomMetadata []byte
}
