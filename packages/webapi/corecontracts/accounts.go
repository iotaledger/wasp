package corecontracts

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/webapi/common"
)

func GetTotalAssets(ch chain.Chain, blockIndexOrTrieRoot string) (isc.CoinBalances, error) {
	ret, err := common.CallView(ch, accounts.ViewTotalAssets.Message(), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return accounts.ViewTotalAssets.DecodeOutput(ret)
}

func GetAccountBalance(ch chain.Chain, agentID isc.AgentID, blockIndexOrTrieRoot string) (isc.CoinBalances, error) {
	ret, err := common.CallView(ch, accounts.ViewBalance.Message(&agentID), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return accounts.ViewBalance.DecodeOutput(ret)
}

func GetAccountObjects(ch chain.Chain, agentID isc.AgentID, blockIndexOrTrieRoot string) ([]isc.IotaObject, error) {
	ret, err := common.CallView(ch, accounts.ViewAccountObjects.Message(&agentID), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return accounts.ViewAccountObjects.DecodeOutput(ret)
}

func GetAccountNonce(ch chain.Chain, agentID isc.AgentID, blockIndexOrTrieRoot string) (uint64, error) {
	ret, err := common.CallView(ch, accounts.ViewGetAccountNonce.Message(&agentID), blockIndexOrTrieRoot)
	if err != nil {
		return 0, err
	}
	return accounts.ViewGetAccountNonce.DecodeOutput(ret)
}
