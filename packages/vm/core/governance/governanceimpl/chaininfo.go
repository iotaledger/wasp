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
	ret.Set(governance.ParamChainID, codec.EncodeChainID(info.ChainID))
	ret.Set(governance.VarChainOwnerID, codec.EncodeAgentID(info.ChainOwnerID))
	ret.Set(governance.VarGasFeePolicyBytes, info.GasFeePolicy.Bytes())
	ret.Set(governance.VarGasLimitsBytes, info.GasLimits.Bytes())
	if len(info.CustomMetadata) > 0 {
		ret.Set(governance.VarCustomMetadata, info.CustomMetadata)
	}
	return ret
}
