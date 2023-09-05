package corecontracts

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/webapi/common"
)

func GetControlAddresses(ch chain.Chain) (*isc.ControlAddresses, error) {
	aliasOutputID, err := ch.LatestAliasOutput(chain.ConfirmedState)
	if err != nil {
		return nil, err
	}
	aliasOutput := aliasOutputID.GetAliasOutput()

	controlAddresses := &isc.ControlAddresses{
		StateAddress:     aliasOutputID.GetStateAddress(),
		GoverningAddress: aliasOutput.GovernorAddress(),
		SinceBlockIndex:  aliasOutput.StateIndex,
	}

	return controlAddresses, nil
}

func handleBlockInfo(info dict.Dict) (*blocklog.BlockInfo, error) {
	resultDecoder := kvdecoder.New(info)

	blockInfoBin, err := resultDecoder.GetBytes(blocklog.ParamBlockInfo)
	if err != nil {
		return nil, err
	}

	blockInfo, err := blocklog.BlockInfoFromBytes(blockInfoBin)
	if err != nil {
		return nil, err
	}

	return blockInfo, nil
}

func GetLatestBlockInfo(ch chain.Chain, blockIndexOrTrieRoot string) (*blocklog.BlockInfo, error) {
	ret, err := common.CallView(ch, blocklog.Contract.Hname(), blocklog.ViewGetBlockInfo.Hname(), nil, blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}

	return handleBlockInfo(ret)
}

func GetBlockInfo(ch chain.Chain, blockIndex uint32, blockIndexOrTrieRoot string) (*blocklog.BlockInfo, error) {
	ret, err := common.CallView(ch, blocklog.Contract.Hname(), blocklog.ViewGetBlockInfo.Hname(), codec.MakeDict(map[string]interface{}{
		blocklog.ParamBlockIndex: blockIndex,
	}), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}

	return handleBlockInfo(ret)
}

func handleRequestIDs(requestIDsDict dict.Dict) (ret []isc.RequestID, err error) {
	requestIDs := collections.NewArrayReadOnly(requestIDsDict, blocklog.ParamRequestID)
	ret = make([]isc.RequestID, requestIDs.Len())
	for i := range ret {
		ret[i], err = isc.RequestIDFromBytes(requestIDs.GetAt(uint32(i)))
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func GetRequestIDsForLatestBlock(ch chain.Chain, blockIndexOrTrieRoot string) ([]isc.RequestID, error) {
	ret, err := common.CallView(ch, blocklog.Contract.Hname(), blocklog.ViewGetRequestIDsForBlock.Hname(), nil, blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}

	return handleRequestIDs(ret)
}

func GetRequestIDsForBlock(ch chain.Chain, blockIndex uint32, blockIndexOrTrieRoot string) ([]isc.RequestID, error) {
	ret, err := common.CallView(
		ch,
		blocklog.Contract.Hname(),
		blocklog.ViewGetRequestIDsForBlock.Hname(),
		codec.MakeDict(map[string]interface{}{blocklog.ParamBlockIndex: blockIndex}),
		blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}

	return handleRequestIDs(ret)
}

func GetRequestReceipt(ch chain.Chain, requestID isc.RequestID, blockIndexOrTrieRoot string) (*blocklog.RequestReceipt, error) {
	ret, err := common.CallView(
		ch,
		blocklog.Contract.Hname(),
		blocklog.ViewGetRequestReceipt.Hname(),
		codec.MakeDict(map[string]interface{}{blocklog.ParamRequestID: requestID}),
		blockIndexOrTrieRoot,
	)
	if err != nil || ret == nil {
		return nil, err
	}

	resultDecoder := kvdecoder.New(ret)

	binRec, err := resultDecoder.GetBytes(blocklog.ParamRequestRecord)
	if err != nil {
		return nil, err
	}
	blockIndex, err := resultDecoder.GetUint32(blocklog.ParamBlockIndex)
	if err != nil {
		return nil, err
	}
	requestIndex, err := resultDecoder.GetUint16(blocklog.ParamRequestIndex)
	if err != nil {
		return nil, err
	}
	return blocklog.RequestReceiptFromBytes(binRec, blockIndex, requestIndex)
}

func GetRequestReceiptsForBlock(ch chain.Chain, blockIndex uint32, blockIndexOrTrieRoot string) ([]*blocklog.RequestReceipt, error) {
	res, err := common.CallView(
		ch,
		blocklog.Contract.Hname(),
		blocklog.ViewGetRequestReceiptsForBlock.Hname(),
		codec.MakeDict(map[string]interface{}{blocklog.ParamBlockIndex: blockIndex}),
		blockIndexOrTrieRoot,
	)
	if err != nil {
		return nil, err
	}
	return blocklog.ReceiptsFromViewCallResult(res)
}

func IsRequestProcessed(ch chain.Chain, requestID isc.RequestID, blockIndexOrTrieRoot string) (bool, error) {
	ret, err := common.CallView(
		ch,
		blocklog.Contract.Hname(),
		blocklog.ViewIsRequestProcessed.Hname(),
		codec.MakeDict(map[string]interface{}{blocklog.ParamRequestID: requestID}),
		blockIndexOrTrieRoot,
	)
	if err != nil {
		return false, err
	}

	resultDecoder := kvdecoder.New(ret)
	isProcessed, err := resultDecoder.GetBool(blocklog.ParamRequestProcessed)
	if err != nil {
		return false, err
	}

	return isProcessed, nil
}

func GetEventsForRequest(ch chain.Chain, requestID isc.RequestID, blockIndexOrTrieRoot string) ([]*isc.Event, error) {
	ret, err := common.CallView(
		ch,
		blocklog.Contract.Hname(),
		blocklog.ViewGetEventsForRequest.Hname(),
		codec.MakeDict(map[string]interface{}{blocklog.ParamRequestID: requestID}),
		blockIndexOrTrieRoot,
	)
	if err != nil {
		return nil, err
	}

	return blocklog.EventsFromViewResult(ret)
}

func GetEventsForBlock(ch chain.Chain, blockIndex uint32, blockIndexOrTrieRoot string) ([]*isc.Event, error) {
	ret, err := common.CallView(
		ch,
		blocklog.Contract.Hname(),
		blocklog.ViewGetEventsForBlock.Hname(),
		codec.MakeDict(map[string]interface{}{blocklog.ParamBlockIndex: blockIndex}),
		blockIndexOrTrieRoot,
	)
	if err != nil {
		return nil, err
	}

	return blocklog.EventsFromViewResult(ret)
}

func GetEventsForContract(ch chain.Chain, contractHname isc.Hname, blockIndexOrTrieRoot string) ([]*isc.Event, error) {
	ret, err := common.CallView(
		ch,
		blocklog.Contract.Hname(),
		blocklog.ViewGetEventsForContract.Hname(),
		codec.MakeDict(map[string]interface{}{blocklog.ParamContractHname: contractHname}),
		blockIndexOrTrieRoot,
	)
	if err != nil {
		return nil, err
	}

	return blocklog.EventsFromViewResult(ret)
}
