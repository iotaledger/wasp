package corecontracts

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/webapi/common"
)

func GetControlAddresses(ch chain.Chain) (*isc.ControlAddresses, error) {
	aliasOutputID, err := ch.LatestAliasOutput(chain.ConfirmedState)
	if err != nil {
		return nil, err
	}
	aliasOutput := aliasOutputID.GetAliasOutput()
	return &isc.ControlAddresses{
		StateAddress:     cryptolib.NewAddressFromIotago(aliasOutputID.GetStateAddress()),
		GoverningAddress: cryptolib.NewAddressFromIotago(aliasOutput.GovernorAddress()),
		SinceBlockIndex:  aliasOutput.StateIndex,
	}, nil
}

func GetLatestBlockInfo(ch chain.Chain, blockIndexOrTrieRoot string) (*blocklog.BlockInfo, error) {
	ret, err := common.CallView(ch, blocklog.ViewGetBlockInfo.Message(nil), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return blocklog.ViewGetBlockInfo.Output2.Decode(ret)
}

func GetBlockInfo(ch chain.Chain, blockIndex uint32, blockIndexOrTrieRoot string) (*blocklog.BlockInfo, error) {
	ret, err := common.CallView(ch, blocklog.ViewGetBlockInfo.Message(&blockIndex), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return blocklog.ViewGetBlockInfo.Output2.Decode(ret)
}

func GetRequestIDsForLatestBlock(ch chain.Chain, blockIndexOrTrieRoot string) ([]isc.RequestID, error) {
	ret, err := common.CallView(ch, blocklog.ViewGetRequestIDsForBlock.Message(nil), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return blocklog.ViewGetRequestIDsForBlock.Output2.Decode(ret)
}

func GetRequestIDsForBlock(ch chain.Chain, blockIndex uint32, blockIndexOrTrieRoot string) ([]isc.RequestID, error) {
	ret, err := common.CallView(ch, blocklog.ViewGetRequestIDsForBlock.Message(&blockIndex), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return blocklog.ViewGetRequestIDsForBlock.Output2.Decode(ret)
}

func GetRequestReceipt(ch chain.Chain, requestID isc.RequestID, blockIndexOrTrieRoot string) (*blocklog.RequestReceipt, error) {
	ret, err := common.CallView(ch, blocklog.ViewGetRequestReceipt.Message(requestID), blockIndexOrTrieRoot)
	if err != nil || ret == nil {
		return nil, err
	}
	return blocklog.ViewGetRequestReceipt.Output.Decode(ret)
}

func GetRequestReceiptsForBlock(ch chain.Chain, blockIndex uint32, blockIndexOrTrieRoot string) ([]*blocklog.RequestReceipt, error) {
	ret, err := common.CallView(ch, blocklog.ViewGetRequestReceiptsForBlock.Message(&blockIndex), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return blocklog.ViewGetRequestReceiptsForBlock.Output2.Decode(ret)
}

func IsRequestProcessed(ch chain.Chain, requestID isc.RequestID, blockIndexOrTrieRoot string) (bool, error) {
	ret, err := common.CallView(ch, blocklog.ViewIsRequestProcessed.Message(requestID), blockIndexOrTrieRoot)
	if err != nil {
		return false, err
	}
	return blocklog.ViewIsRequestProcessed.Output.Decode(ret)
}

func GetEventsForRequest(ch chain.Chain, requestID isc.RequestID, blockIndexOrTrieRoot string) ([]*isc.Event, error) {
	ret, err := common.CallView(ch, blocklog.ViewGetEventsForRequest.Message(requestID), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return blocklog.ViewGetEventsForRequest.Output.Decode(ret)
}

func GetEventsForBlock(ch chain.Chain, blockIndex uint32, blockIndexOrTrieRoot string) ([]*isc.Event, error) {
	ret, err := common.CallView(ch, blocklog.ViewGetEventsForBlock.Message(&blockIndex), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return blocklog.ViewGetEventsForBlock.Output2.Decode(ret)
}
