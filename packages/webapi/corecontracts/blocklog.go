package corecontracts

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/chain"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/v2/packages/webapi/common"
)

func GetControlAddresses(ch chain.Chain) (*isc.ControlAddresses, error) {
	state, err := ch.LatestState(chain.ConfirmedState)
	if err != nil {
		return nil, err
	}

	committeeAddr := ch.GetCommitteeInfo().Address

	admin, err := GetChainAdmin(ch, fmt.Sprintf("%d", state.BlockIndex()))
	if err != nil {
		return nil, err
	}

	return &isc.ControlAddresses{
		AnchorOwner:     committeeAddr,
		ChainAdmin:      admin,
		SinceBlockIndex: state.BlockIndex(),
	}, nil
}

func GetLatestBlockInfo(ch chain.Chain, blockIndexOrTrieRoot string) (uint32, *blocklog.BlockInfo, error) {
	ret, err := common.CallView(ch, blocklog.ViewGetBlockInfo.Message(nil), blockIndexOrTrieRoot)
	if err != nil {
		return 0, nil, err
	}
	return blocklog.ViewGetBlockInfo.DecodeOutput(ret)
}

func GetBlockInfo(ch chain.Chain, blockIndex uint32, blockIndexOrTrieRoot string) (uint32, *blocklog.BlockInfo, error) {
	ret, err := common.CallView(ch, blocklog.ViewGetBlockInfo.Message(&blockIndex), blockIndexOrTrieRoot)
	if err != nil {
		return 0, nil, err
	}
	return blocklog.ViewGetBlockInfo.DecodeOutput(ret)
}

func GetRequestIDsForLatestBlock(ch chain.Chain, blockIndexOrTrieRoot string) (uint32, []isc.RequestID, error) {
	ret, err := common.CallView(ch, blocklog.ViewGetRequestIDsForBlock.Message(nil), blockIndexOrTrieRoot)
	if err != nil {
		return 0, nil, err
	}
	return blocklog.ViewGetRequestIDsForBlock.DecodeOutput(ret)
}

func GetRequestIDsForBlock(ch chain.Chain, blockIndex uint32, blockIndexOrTrieRoot string) (uint32, []isc.RequestID, error) {
	ret, err := common.CallView(ch, blocklog.ViewGetRequestIDsForBlock.Message(&blockIndex), blockIndexOrTrieRoot)
	if err != nil {
		return 0, nil, err
	}
	return blocklog.ViewGetRequestIDsForBlock.DecodeOutput(ret)
}

func GetRequestReceipt(ch chain.Chain, requestID isc.RequestID, blockIndexOrTrieRoot string) (*blocklog.RequestReceipt, error) {
	ret, err := common.CallView(ch, blocklog.ViewGetRequestReceipt.Message(requestID), blockIndexOrTrieRoot)
	if err != nil || ret == nil {
		return nil, err
	}
	return blocklog.ViewGetRequestReceipt.DecodeOutput(ret)
}

func GetRequestReceiptsForBlock(ch chain.Chain, blockIndex uint32, blockIndexOrTrieRoot string) (*blocklog.RequestReceiptsResponse, error) {
	ret, err := common.CallView(ch, blocklog.ViewGetRequestReceiptsForBlock.Message(&blockIndex), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return blocklog.ViewGetRequestReceiptsForBlock.DecodeOutput(ret)
}

func IsRequestProcessed(ch chain.Chain, requestID isc.RequestID, blockIndexOrTrieRoot string) (bool, error) {
	ret, err := common.CallView(ch, blocklog.ViewIsRequestProcessed.Message(requestID), blockIndexOrTrieRoot)
	if err != nil {
		return false, err
	}
	return blocklog.ViewIsRequestProcessed.DecodeOutput(ret)
}

func GetEventsForRequest(ch chain.Chain, requestID isc.RequestID, blockIndexOrTrieRoot string) ([]*isc.Event, error) {
	ret, err := common.CallView(ch, blocklog.ViewGetEventsForRequest.Message(requestID), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return blocklog.ViewGetEventsForRequest.DecodeOutput(ret)
}

func GetEventsForBlock(ch chain.Chain, blockIndex uint32, blockIndexOrTrieRoot string) (uint32, []*isc.Event, error) {
	ret, err := common.CallView(ch, blocklog.ViewGetEventsForBlock.Message(&blockIndex), blockIndexOrTrieRoot)
	if err != nil {
		return 0, nil, err
	}
	return blocklog.ViewGetEventsForBlock.DecodeOutput(ret)
}
