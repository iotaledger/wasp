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

	if len(info.PublicURL) > 0 {
		ret.Set(governance.VarPublicURL, codec.EncodeString(info.PublicURL))
	}

	if len(info.Metadata.EVMJsonRPCURL) > 0 {
		ret.Set(governance.VarMetadataEVMJsonRPCURL, codec.EncodeString(info.Metadata.EVMJsonRPCURL))
	}

	if len(info.Metadata.EVMWebSocketURL) > 0 {
		ret.Set(governance.VarMetadataEVMWebSocketURL, codec.EncodeString(info.Metadata.EVMWebSocketURL))
	}

	if len(info.Metadata.ChainName) > 0 {
		ret.Set(governance.VarMetadataChainName, codec.EncodeString(info.Metadata.ChainName))
	}

	if len(info.Metadata.ChainDescription) > 0 {
		ret.Set(governance.VarMetadataChainDescription, codec.EncodeString(info.Metadata.ChainDescription))
	}

	if len(info.Metadata.ChainOwnerEmail) > 0 {
		ret.Set(governance.VarMetadataChainOwnerEmail, codec.EncodeString(info.Metadata.ChainOwnerEmail))
	}

	if len(info.Metadata.ChainWebsite) > 0 {
		ret.Set(governance.VarMetadataChainWebsite, codec.EncodeString(info.Metadata.ChainWebsite))
	}

	return ret
}
