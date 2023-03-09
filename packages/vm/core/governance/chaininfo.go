// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// ChainInfo is an API structure containing the main properties of the chain
type ChainInfo struct {
	ChainID        isc.ChainID
	ChainOwnerID   isc.AgentID
	GasFeePolicy   *gas.FeePolicy
	CustomMetadata []byte
}
