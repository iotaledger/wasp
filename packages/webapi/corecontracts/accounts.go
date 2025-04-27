package corecontracts

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/webapi/common"
)

func GetTotalAssets(ch chain.Chain, blockIndexOrTrieRoot string) (isc.CoinBalances, error) {
	ret, err := common.CallView(ch, accounts.ViewTotalAssets.Message(), blockIndexOrTrieRoot)
	if err != nil {
		return isc.CoinBalances{}, err
	}
	return accounts.ViewTotalAssets.DecodeOutput(ret)
}

func GetAccountBalance(ch chain.Chain, agentID isc.AgentID, blockIndexOrTrieRoot string) (isc.CoinBalances, error) {
	ret, err := common.CallView(ch, accounts.ViewBalance.Message(&agentID), blockIndexOrTrieRoot)
	if err != nil {
		return isc.CoinBalances{}, err
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

func GetAccountFoundries(ch chain.Chain, agentID isc.AgentID, blockIndexOrTrieRoot string) ([]coin.Type, error) {
	panic("minting on L2 is currently unsupported")
}

func GetAccountNonce(ch chain.Chain, agentID isc.AgentID, blockIndexOrTrieRoot string) (uint64, error) {
	ret, err := common.CallView(ch, accounts.ViewGetAccountNonce.Message(&agentID), blockIndexOrTrieRoot)
	if err != nil {
		return 0, err
	}
	return accounts.ViewGetAccountNonce.DecodeOutput(ret)
}

func GetNativeTokenIDRegistry(ch chain.Chain, blockIndexOrTrieRoot string) ([]coin.Type, error) {
	panic("TODO")
	// ret, err := common.CallView(ch, accounts.ViewGetNativeTokenIDRegistry.Message(), blockIndexOrTrieRoot)
	// if err != nil {
	// 	return nil, err
	// }
	// return accounts.ViewGetNativeTokenIDRegistry.DecodeOutput(ret)
}
