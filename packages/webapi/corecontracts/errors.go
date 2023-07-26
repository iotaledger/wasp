package corecontracts

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/webapi/common"
)

func ErrorMessageFormat(ch chain.Chain, contractID isc.Hname, errorID uint16, blockIndexOrTrieRoot string) (string, error) {
	errorCode := isc.NewVMErrorCode(contractID, errorID)

	ret, err := common.CallView(
		ch,
		errors.Contract.Hname(),
		errors.ViewGetErrorMessageFormat.Hname(),
		codec.MakeDict(map[string]interface{}{errors.ParamErrorCode: errorCode.Bytes()}),
		blockIndexOrTrieRoot,
	)
	if err != nil {
		return "", err
	}

	resultDecoder := kvdecoder.New(ret)
	messageFormat, err := resultDecoder.GetString(errors.ParamErrorMessageFormat)
	if err != nil {
		return "", err
	}

	return messageFormat, nil
}
