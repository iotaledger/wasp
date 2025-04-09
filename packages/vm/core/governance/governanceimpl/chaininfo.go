// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

// getChainInfo view returns general info about the chain: chain ID, chain admin ID, limits and default fees
func getChainInfo(ctx isc.SandboxView) *isc.ChainInfo {
	state := governance.NewStateReaderFromSandbox(ctx)
	return state.GetChainInfo(ctx.ChainID())
}
