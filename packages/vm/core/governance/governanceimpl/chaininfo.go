// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

// getChainInfo view returns general info about the chain: chain ID, chain owner ID, limits and default fees
func getChainInfo(ctx isc.SandboxView) *isc.ChainInfo {
	return lo.Must(governance.GetChainInfo(ctx.StateR(), ctx.ChainID()))
}
