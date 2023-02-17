package corecontracts

import (
	"errors"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
)

type Accounts struct {
	vmService interfaces.VMService
}

func NewAccounts(vmService interfaces.VMService) *Accounts {
	return &Accounts{
		vmService: vmService,
	}
}

func (a *Accounts) GetAccounts(chainID isc.ChainID) ([]isc.AgentID, error) {
	ret, err := a.vmService.CallViewByChainID(chainID, accounts.Contract.Hname(), accounts.ViewAccounts.Hname(), nil)
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

func (a *Accounts) GetTotalAssets(chainID isc.ChainID) (*isc.Assets, error) {
	ret, err := a.vmService.CallViewByChainID(chainID, accounts.Contract.Hname(), accounts.ViewTotalAssets.Hname(), nil)
	if err != nil {
		return nil, err
	}

	return isc.AssetsFromDict(ret)
}

func (a *Accounts) GetAccountBalance(chainID isc.ChainID, agentID isc.AgentID) (*isc.Assets, error) {
	ret, err := a.vmService.CallViewByChainID(chainID, accounts.Contract.Hname(), accounts.ViewBalance.Hname(), codec.MakeDict(map[string]interface{}{
		accounts.ParamAgentID: agentID,
	}))
	if err != nil {
		return nil, err
	}

	return isc.AssetsFromDict(ret)
}

func (a *Accounts) GetAccountNFTs(chainID isc.ChainID, agentID isc.AgentID) ([]iotago.NFTID, error) {
	ret, err := a.vmService.CallViewByChainID(chainID, accounts.Contract.Hname(), accounts.ViewAccountNFTs.Hname(), codec.MakeDict(map[string]interface{}{
		accounts.ParamAgentID: agentID,
	}))
	if err != nil {
		return nil, err
	}

	nftIDsCollection := collections.NewArray16ReadOnly(ret, accounts.ParamNFTIDs)
	nftLen, err := nftIDsCollection.Len()
	nftIDs := make([]iotago.NFTID, 0)
	if err != nil {
		return nil, err
	}

	for i := uint16(0); i < nftLen; i++ {
		nftID := iotago.NFTID{}
		nftIDBytes, err := nftIDsCollection.GetAt(i)
		if err != nil {
			return nil, err
		}

		copy(nftID[:], nftIDBytes)
		nftIDs = append(nftIDs, nftID)
	}

	return nftIDs, nil
}

func (a *Accounts) GetAccountFoundries(chainID isc.ChainID, agentID isc.AgentID) ([]uint32, error) {
	ret, err := a.vmService.CallViewByChainID(chainID, accounts.Contract.Hname(), accounts.ViewAccountFoundries.Hname(), dict.Dict{
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

func (a *Accounts) GetAccountNonce(chainID isc.ChainID, agentID isc.AgentID) (uint64, error) {
	ret, err := a.vmService.CallViewByChainID(chainID, accounts.Contract.Hname(), accounts.ViewGetAccountNonce.Hname(), codec.MakeDict(map[string]interface{}{
		accounts.ParamAgentID: agentID,
	}))
	if err != nil {
		return 0, err
	}

	nonce, err := ret.Get(accounts.ParamAccountNonce)
	if err != nil {
		return 0, err
	}

	return codec.DecodeUint64(nonce)
}

func (a *Accounts) GetNFTData(chainID isc.ChainID, nftID iotago.NFTID) (*isc.NFT, error) {
	ret, err := a.vmService.CallViewByChainID(chainID, accounts.Contract.Hname(), accounts.ViewNFTData.Hname(), codec.MakeDict(map[string]interface{}{
		accounts.ParamNFTID: nftID[:],
	}))
	if err != nil {
		return nil, err
	}

	nftData, err := isc.NFTFromBytes(ret.MustGet(accounts.ParamNFTData))
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

func (a *Accounts) GetNativeTokenIDRegistry(chainID isc.ChainID) ([]iotago.NativeTokenID, error) {
	ret, err := a.vmService.CallViewByChainID(chainID, accounts.Contract.Hname(), accounts.ViewGetNativeTokenIDRegistry.Hname(), nil)
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

func (a *Accounts) GetFoundryOutput(chainID isc.ChainID, serialNumber uint32) (*iotago.FoundryOutput, error) {
	res, err := a.vmService.CallViewByChainID(chainID, accounts.Contract.Hname(), accounts.ViewFoundryOutput.Hname(), codec.MakeDict(map[string]interface{}{
		accounts.ParamFoundrySN: serialNumber,
	}))
	if err != nil {
		return nil, err
	}

	outBin, err := res.Get(accounts.ParamFoundryOutputBin)
	if err != nil {
		return nil, err
	}

	out := &iotago.FoundryOutput{}
	_, err = out.Deserialize(outBin, serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil, err
	}

	return out, nil
}
