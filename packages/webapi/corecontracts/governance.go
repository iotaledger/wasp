// Package corecontracts provides methods for the webapi to interact with core contracts
package corecontracts

import (
	"github.com/iotaledger/wasp/v2/packages/chain"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/webapi/common"
)

func GetChainAdmin(ch chain.Chain, blockIndexOrTrieRoot string) (isc.AgentID, error) {
	ret, err := common.CallView(ch, governance.ViewGetChainAdmin.Message(), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return governance.ViewGetChainAdmin.DecodeOutput(ret)
}

func GetChainInfo(ch chain.Chain, blockIndexOrTrieRoot string) (*isc.ChainInfo, error) {
	ret, err := common.CallView(ch, governance.ViewGetChainInfo.Message(), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return governance.ViewGetChainInfo.DecodeOutput(ret)
}
