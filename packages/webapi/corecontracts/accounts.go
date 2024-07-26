package corecontracts

import (
	"github.com/samber/lo"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/webapi/common"
)

func GetTotalAssets(ch chain.Chain, blockIndexOrTrieRoot string) (*isc.Assets, error) {
	ret, err := common.CallView(ch, accounts.ViewTotalAssets.Message(), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return accounts.ViewTotalAssets.Output.Decode(ret)
}

func GetAccountBalance(ch chain.Chain, agentID isc.AgentID, blockIndexOrTrieRoot string) (*isc.Assets, error) {
	ret, err := common.CallView(ch, accounts.ViewBalance.Message(&agentID), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return accounts.ViewTotalAssets.Output.Decode(ret)
}

func GetAccountNFTs(ch chain.Chain, agentID isc.AgentID, blockIndexOrTrieRoot string) ([]iotago.NFTID, error) {
	ret, err := common.CallView(ch, accounts.ViewAccountNFTs.Message(&agentID), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return accounts.ViewAccountNFTs.Output.Decode(ret)
}

func GetAccountFoundries(ch chain.Chain, agentID isc.AgentID, blockIndexOrTrieRoot string) ([]uint32, error) {
	ret, err := common.CallView(ch, accounts.ViewAccountFoundries.Message(&agentID), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	sns, err := accounts.ViewAccountFoundries.Output.Decode(ret)
	if err != nil {
		return nil, err
	}
	return lo.Keys(sns), nil
}

func GetAccountNonce(ch chain.Chain, agentID isc.AgentID, blockIndexOrTrieRoot string) (uint64, error) {
	ret, err := common.CallView(ch, accounts.ViewGetAccountNonce.Message(&agentID), blockIndexOrTrieRoot)
	if err != nil {
		return 0, err
	}
	return accounts.ViewGetAccountNonce.Output.Decode(ret)
}

func GetNFTData(ch chain.Chain, nftID iotago.NFTID, blockIndexOrTrieRoot string) (*isc.NFT, error) {
	ret, err := common.CallView(ch, accounts.ViewNFTData.Message(nftID), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return accounts.ViewNFTData.Output.Decode(ret)
}

func GetNativeTokenIDRegistry(ch chain.Chain, blockIndexOrTrieRoot string) ([]iotago.NativeTokenID, error) {
	ret, err := common.CallView(ch, accounts.ViewGetNativeTokenIDRegistry.Message(), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return accounts.ViewGetNativeTokenIDRegistry.Output.Decode(ret)
}

func GetFoundryOutput(ch chain.Chain, serialNumber uint32, blockIndexOrTrieRoot string) (*iotago.FoundryOutput, error) {
	ret, err := common.CallView(ch, accounts.ViewNativeToken.Message(serialNumber), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	out, err := accounts.ViewNativeToken.Output.Decode(ret)
	if err != nil {
		return nil, err
	}
	return out.(*iotago.FoundryOutput), nil
}
