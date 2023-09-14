package corecontracts

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/webapi/common"
)

func GetAccounts(ch chain.Chain, blockIndexOrTrieRoot string) ([]isc.AgentID, error) {
	accountIDs, err := common.CallView(ch, accounts.Contract.Hname(), accounts.ViewAccounts.Hname(), nil, blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}

	ret := make([]isc.AgentID, 0)
	for accountID := range accountIDs {
		agentID, err := accounts.AgentIDFromKey(accountID, ch.ID())
		if err != nil {
			return nil, err
		}
		ret = append(ret, agentID)
	}
	return ret, nil
}

func GetTotalAssets(ch chain.Chain, blockIndexOrTrieRoot string) (*isc.Assets, error) {
	ret, err := common.CallView(ch, accounts.Contract.Hname(), accounts.ViewTotalAssets.Hname(), nil, blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}

	return isc.AssetsFromDict(ret)
}

func GetAccountBalance(ch chain.Chain, agentID isc.AgentID, blockIndexOrTrieRoot string) (*isc.Assets, error) {
	ret, err := common.CallView(
		ch,
		accounts.Contract.Hname(),
		accounts.ViewBalance.Hname(), codec.MakeDict(map[string]interface{}{accounts.ParamAgentID: agentID}),
		blockIndexOrTrieRoot,
	)
	if err != nil {
		return nil, err
	}

	return isc.AssetsFromDict(ret)
}

func GetAccountNFTs(ch chain.Chain, agentID isc.AgentID, blockIndexOrTrieRoot string) ([]iotago.NFTID, error) {
	res, err := common.CallView(
		ch,
		accounts.Contract.Hname(),
		accounts.ViewAccountNFTs.Hname(), codec.MakeDict(map[string]interface{}{accounts.ParamAgentID: agentID}),
		blockIndexOrTrieRoot,
	)
	if err != nil {
		return nil, err
	}

	nftIDs := collections.NewArrayReadOnly(res, accounts.ParamNFTIDs)
	ret := make([]iotago.NFTID, nftIDs.Len())
	for i := range ret {
		copy(ret[i][:], nftIDs.GetAt(uint32(i)))
	}
	return ret, nil
}

func GetAccountFoundries(ch chain.Chain, agentID isc.AgentID, blockIndexOrTrieRoot string) ([]uint32, error) {
	foundrySNs, err := common.CallView(
		ch,
		accounts.Contract.Hname(),
		accounts.ViewAccountFoundries.Hname(), dict.Dict{accounts.ParamAgentID: codec.EncodeAgentID(agentID)},
		blockIndexOrTrieRoot,
	)
	if err != nil {
		return nil, err
	}
	ret := make([]uint32, 0, len(foundrySNs))
	for foundrySN := range foundrySNs {
		sn, err := codec.DecodeUint32([]byte(foundrySN))
		if err != nil {
			return nil, err
		}
		ret = append(ret, sn)
	}
	return ret, nil
}

func GetAccountNonce(ch chain.Chain, agentID isc.AgentID, blockIndexOrTrieRoot string) (uint64, error) {
	ret, err := common.CallView(
		ch,
		accounts.Contract.Hname(),
		accounts.ViewGetAccountNonce.Hname(), codec.MakeDict(map[string]interface{}{accounts.ParamAgentID: agentID}),
		blockIndexOrTrieRoot,
	)
	if err != nil {
		return 0, err
	}

	nonce := ret.Get(accounts.ParamAccountNonce)

	return codec.DecodeUint64(nonce)
}

func GetNFTData(ch chain.Chain, nftID iotago.NFTID, blockIndexOrTrieRoot string) (*isc.NFT, error) {
	ret, err := common.CallView(
		ch,
		accounts.Contract.Hname(),
		accounts.ViewNFTData.Hname(), codec.MakeDict(map[string]interface{}{accounts.ParamNFTID: nftID[:]}),
		blockIndexOrTrieRoot,
	)
	if err != nil {
		return nil, err
	}

	nftData, err := isc.NFTFromBytes(ret.Get(accounts.ParamNFTData))
	if err != nil {
		return nil, err
	}

	return nftData, nil
}

func GetNativeTokenIDRegistry(ch chain.Chain, blockIndexOrTrieRoot string) ([]iotago.NativeTokenID, error) {
	nativeTokenIDs, err := common.CallView(ch, accounts.Contract.Hname(), accounts.ViewGetNativeTokenIDRegistry.Hname(), nil, blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}

	ret := make([]iotago.NativeTokenID, 0, len(nativeTokenIDs))
	for nativeTokenID := range nativeTokenIDs {
		tokenID, err := isc.NativeTokenIDFromBytes([]byte(nativeTokenID))
		if err != nil {
			return nil, err
		}
		ret = append(ret, tokenID)
	}

	return ret, nil
}

func GetFoundryOutput(ch chain.Chain, serialNumber uint32, blockIndexOrTrieRoot string) (*iotago.FoundryOutput, error) {
	res, err := common.CallView(
		ch,
		accounts.Contract.Hname(),
		accounts.ViewFoundryOutput.Hname(), codec.MakeDict(map[string]interface{}{accounts.ParamFoundrySN: serialNumber}),
		blockIndexOrTrieRoot,
	)
	if err != nil {
		return nil, err
	}

	outBin := res.Get(accounts.ParamFoundryOutputBin)
	out := &iotago.FoundryOutput{}
	_, err = out.Deserialize(outBin, serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil, err
	}

	return out, nil
}
