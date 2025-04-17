package corecontracts

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/webapi/common"
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
