package corecontracts

import (
	"errors"

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

func GetAccounts(ch chain.Chain) ([]isc.AgentID, error) {
	ret, err := common.CallView(ch, accounts.Contract.Hname(), accounts.ViewAccounts.Hname(), nil)
	if err != nil {
		return nil, err
	}

	accountIds := make([]isc.AgentID, 0)

	for k := range ret {
		agentID, err := codec.DecodeAgentID([]byte(k))
		if err != nil {
			return nil, err
		}

		accountIds = append(accountIds, agentID)
	}

	return accountIds, nil
}

func GetTotalAssets(ch chain.Chain) (*isc.Assets, error) {
	ret, err := common.CallView(ch, accounts.Contract.Hname(), accounts.ViewTotalAssets.Hname(), nil)
	if err != nil {
		return nil, err
	}

	return isc.AssetsFromDict(ret)
}

func GetAccountBalance(ch chain.Chain, agentID isc.AgentID) (*isc.Assets, error) {
	ret, err := common.CallView(ch, accounts.Contract.Hname(), accounts.ViewBalance.Hname(), codec.MakeDict(map[string]interface{}{
		accounts.ParamAgentID: agentID,
	}))
	if err != nil {
		return nil, err
	}

	return isc.AssetsFromDict(ret)
}

func GetAccountNFTs(ch chain.Chain, agentID isc.AgentID) ([]iotago.NFTID, error) {
	res, err := common.CallView(ch, accounts.Contract.Hname(), accounts.ViewAccountNFTs.Hname(), codec.MakeDict(map[string]interface{}{
		accounts.ParamAgentID: agentID,
	}))
	if err != nil {
		return nil, err
	}

	nftIDs := collections.NewArrayReadOnly(res, accounts.ParamNFTIDs)
	ret := make([]iotago.NFTID, 0, nftIDs.Len())
	for i := range ret {
		nftID := iotago.NFTID{}
		copy(nftID[:], nftIDs.GetAt(uint32(i)))
		ret = append(ret, nftID)
	}
	return ret, nil
}

func GetAccountFoundries(ch chain.Chain, agentID isc.AgentID) ([]uint32, error) {
	ret, err := common.CallView(ch, accounts.Contract.Hname(), accounts.ViewAccountFoundries.Hname(), dict.Dict{
		accounts.ParamAgentID: codec.EncodeAgentID(agentID),
	})
	if err != nil {
		return nil, err
	}
	sns := make([]uint32, 0, len(ret))
	for k := range ret {
		sn, err := codec.DecodeUint32([]byte(k))
		if err != nil {
			return nil, err
		}
		sns = append(sns, sn)
	}
	return sns, nil
}

func GetAccountNonce(ch chain.Chain, agentID isc.AgentID) (uint64, error) {
	ret, err := common.CallView(ch, accounts.Contract.Hname(), accounts.ViewGetAccountNonce.Hname(), codec.MakeDict(map[string]interface{}{
		accounts.ParamAgentID: agentID,
	}))
	if err != nil {
		return 0, err
	}

	nonce := ret.Get(accounts.ParamAccountNonce)

	return codec.DecodeUint64(nonce)
}

func GetNFTData(ch chain.Chain, nftID iotago.NFTID) (*isc.NFT, error) {
	ret, err := common.CallView(ch, accounts.Contract.Hname(), accounts.ViewNFTData.Hname(), codec.MakeDict(map[string]interface{}{
		accounts.ParamNFTID: nftID[:],
	}))
	if err != nil {
		return nil, err
	}

	nftData, err := isc.NFTFromBytes(ret.Get(accounts.ParamNFTData))
	if err != nil {
		return nil, err
	}

	return nftData, nil
}

func parseNativeTokenIDFromBytes(data []byte) (iotago.NativeTokenID, error) {
	if len(data) != iotago.NativeTokenIDLength {
		return iotago.NativeTokenID{}, errors.New("len(data) != iotago.NativeTokenIDLength")
	}

	ret := iotago.NativeTokenID{}
	copy(ret[:], data)

	return ret, nil
}

func GetNativeTokenIDRegistry(ch chain.Chain) ([]iotago.NativeTokenID, error) {
	ret, err := common.CallView(ch, accounts.Contract.Hname(), accounts.ViewGetNativeTokenIDRegistry.Hname(), nil)
	if err != nil {
		return nil, err
	}

	nativeTokenIDs := make([]iotago.NativeTokenID, len(ret))
	for k := range ret {
		parsedTokenID, err := parseNativeTokenIDFromBytes([]byte(k))
		if err != nil {
			return nil, err
		}

		nativeTokenIDs = append(nativeTokenIDs, parsedTokenID)
	}

	return nativeTokenIDs, nil
}

func GetFoundryOutput(ch chain.Chain, serialNumber uint32) (*iotago.FoundryOutput, error) {
	res, err := common.CallView(ch, accounts.Contract.Hname(), accounts.ViewFoundryOutput.Hname(), codec.MakeDict(map[string]interface{}{
		accounts.ParamFoundrySN: serialNumber,
	}))
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
