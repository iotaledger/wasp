package internal

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type BlockLog struct {
	vmService interfaces.VMService
}

func (b *BlockLog) GetControlAddresses(chainID *isc.ChainID) (*blocklog.ControlAddresses, error) {
	ret, err := b.vmService.CallViewByChainID(chainID, blocklog.Contract.Hname(), blocklog.ViewControlAddresses.Hname(), nil)
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

func (b *BlockLog) GetBlockInfo(chainID *isc.ChainID, blockIndex uint32) (*blocklog.BlockInfo, error) {
	ret, err := b.vmService.CallViewByChainID(chainID, blocklog.Contract.Hname(), blocklog.ViewGetBlockInfo.Hname(), codec.MakeDict(map[string]interface{}{
		blocklog.ParamBlockIndex: blockIndex,
	}))
	if err != nil {
		return nil, err
	}

	resultDecoder := kvdecoder.New(ret)

	blockInfoBin, err := resultDecoder.GetBytes(blocklog.ParamBlockInfo)
	if err != nil {
		return nil, err
	}

	blockIndexRet, err := resultDecoder.GetUint32(blocklog.ParamBlockIndex)
	if err != nil {
		return nil, err
	}

	blockInfo, err := blocklog.BlockInfoFromBytes(blockIndexRet, blockInfoBin)
	return blockInfo, nil
}
