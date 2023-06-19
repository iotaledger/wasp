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

func GetControlAddresses(ch chain.Chain) (*blocklog.ControlAddresses, error) {
	ret, err := common.CallView(ch, blocklog.Contract.Hname(), blocklog.ViewControlAddresses.Hname(), nil)
	if err != nil {
		return nil, err
	}

	par := kvdecoder.New(ret)

	stateAddress, err := par.GetAddress(blocklog.ParamStateControllerAddress)
	if err != nil {
		return nil, err
	}

	governingAddress, err := par.GetAddress(blocklog.ParamGoverningAddress)
	if err != nil {
		return nil, err
	}

	sinceBlockIndex, err := par.GetUint32(blocklog.ParamBlockIndex)
	if err != nil {
		return nil, err
	}

	controlAddresses := &blocklog.ControlAddresses{
		StateAddress:     stateAddress,
		GoverningAddress: governingAddress,
		SinceBlockIndex:  sinceBlockIndex,
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

func GetLatestBlockInfo(ch chain.Chain) (*blocklog.BlockInfo, error) {
	ret, err := common.CallView(ch, blocklog.Contract.Hname(), blocklog.ViewGetBlockInfo.Hname(), nil)
	if err != nil {
		return nil, err
	}

	return handleBlockInfo(ret)
}

func GetBlockInfo(ch chain.Chain, blockIndex uint32) (*blocklog.BlockInfo, error) {
	ret, err := common.CallView(ch, blocklog.Contract.Hname(), blocklog.ViewGetBlockInfo.Hname(), codec.MakeDict(map[string]interface{}{
		blocklog.ParamBlockIndex: blockIndex,
	}))
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

func GetRequestIDsForLatestBlock(ch chain.Chain) ([]isc.RequestID, error) {
	ret, err := common.CallView(ch, blocklog.Contract.Hname(), blocklog.ViewGetRequestIDsForBlock.Hname(), nil)
	if err != nil {
		return nil, err
	}

	return handleRequestIDs(ret)
}

func GetRequestIDsForBlock(ch chain.Chain, blockIndex uint32) ([]isc.RequestID, error) {
	ret, err := common.CallView(ch, blocklog.Contract.Hname(), blocklog.ViewGetRequestIDsForBlock.Hname(), codec.MakeDict(map[string]interface{}{
		blocklog.ParamBlockIndex: blockIndex,
	}))
	if err != nil {
		return nil, err
	}

	return handleRequestIDs(ret)
}

func GetRequestReceipt(ch chain.Chain, requestID isc.RequestID) (*blocklog.RequestReceipt, error) {
	ret, err := common.CallView(ch, blocklog.Contract.Hname(), blocklog.ViewGetRequestReceipt.Hname(), codec.MakeDict(map[string]interface{}{
		blocklog.ParamRequestID: requestID,
	}))
	if err != nil || ret == nil {
		return nil, err
	}

	resultDecoder := kvdecoder.New(ret)
	binRec, err := resultDecoder.GetBytes(blocklog.ParamRequestRecord)
	if err != nil {
		return nil, err
	}

	requestReceipt, err := blocklog.RequestReceiptFromBytes(binRec)
	if err != nil {
		return nil, err
	}

	requestReceipt.BlockIndex, err = resultDecoder.GetUint32(blocklog.ParamBlockIndex)
	if err != nil {
		return nil, err
	}

	requestReceipt.RequestIndex, err = resultDecoder.GetUint16(blocklog.ParamRequestIndex)
	if err != nil {
		return nil, err
	}

	return requestReceipt, err
}

func GetRequestReceiptsForBlock(ch chain.Chain, blockIndex uint32) ([]*blocklog.RequestReceipt, error) {
	res, err := common.CallView(ch, blocklog.Contract.Hname(), blocklog.ViewGetRequestReceiptsForBlock.Hname(), codec.MakeDict(map[string]interface{}{
		blocklog.ParamBlockIndex: blockIndex,
	}))
	if err != nil {
		return nil, err
	}

	blockIndex, err = codec.DecodeUint32(res.Get(blocklog.ParamBlockIndex))
	if err != nil {
		return nil, err
	}

	receipts := collections.NewArrayReadOnly(res, blocklog.ParamRequestRecord)
	ret := make([]*blocklog.RequestReceipt, receipts.Len())
	for i := range ret {
		ret[i], err = blocklog.RequestReceiptFromBytes(receipts.GetAt(uint32(i)))
		if err != nil {
			return nil, err
		}
		ret[i].WithBlockData(blockIndex, uint16(i))
	}
	return ret, nil
}

func IsRequestProcessed(ch chain.Chain, requestID isc.RequestID) (bool, error) {
	ret, err := common.CallView(ch, blocklog.Contract.Hname(), blocklog.ViewIsRequestProcessed.Hname(), codec.MakeDict(map[string]interface{}{
		blocklog.ParamRequestID: requestID,
	}))
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

func GetEventsForRequest(ch chain.Chain, requestID isc.RequestID) ([]*isc.Event, error) {
	ret, err := common.CallView(ch, blocklog.Contract.Hname(), blocklog.ViewGetEventsForRequest.Hname(), codec.MakeDict(map[string]interface{}{
		blocklog.ParamRequestID: requestID,
	}))
	if err != nil {
		return nil, err
	}

	return blocklog.EventsFromViewResult(ret)
}

func GetEventsForBlock(ch chain.Chain, blockIndex uint32) ([]*isc.Event, error) {
	ret, err := common.CallView(ch, blocklog.Contract.Hname(), blocklog.ViewGetEventsForBlock.Hname(), codec.MakeDict(map[string]interface{}{
		blocklog.ParamBlockIndex: blockIndex,
	}))
	if err != nil {
		return nil, err
	}

	return blocklog.EventsFromViewResult(ret)
}

func GetEventsForContract(ch chain.Chain, contractHname isc.Hname) ([]*isc.Event, error) {
	ret, err := common.CallView(ch, blocklog.Contract.Hname(), blocklog.ViewGetEventsForContract.Hname(), codec.MakeDict(map[string]interface{}{
		blocklog.ParamContractHname: contractHname,
	}))
	if err != nil {
		return nil, err
	}

	return blocklog.EventsFromViewResult(ret)
}
