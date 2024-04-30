// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

// getChainInfo view returns general info about the chain: chain ID, chain owner ID, limits and default fees
func getChainInfo(ctx isc.SandboxView) dict.Dict {
	info := governance.MustGetChainInfo(ctx.StateR(), ctx.ChainID())
	ret := dict.New()
	ret.Set(governance.ParamChainID, codec.ChainID.Encode(info.ChainID))
	ret.Set(governance.VarChainOwnerID, codec.AgentID.Encode(info.ChainOwnerID))
	ret.Set(governance.VarGasFeePolicyBytes, info.GasFeePolicy.Bytes())
	ret.Set(governance.VarGasLimitsBytes, info.GasLimits.Bytes())

	if info.PublicURL != "" {
		ret.Set(governance.VarPublicURL, codec.String.Encode(info.PublicURL))
	}

	ret.Set(governance.VarMetadata, info.Metadata.Bytes())

	return ret
}
