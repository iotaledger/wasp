package internal

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type Errors struct {
	vmService interfaces.VMService
}

func NewErrors(vmService interfaces.VMService) *Errors {
	return &Errors{
		vmService: vmService,
	}
}

func (e *Errors) GetMessageFormat(chainID *isc.ChainID, contractID isc.Hname, errorID uint16) (string, error) {
	errorCode := isc.NewVMErrorCode(contractID, errorID)

	ret, err := e.vmService.CallViewByChainID(chainID, errors.Contract.Hname(), errors.ViewGetErrorMessageFormat.Hname(), codec.MakeDict(map[string]interface{}{
		errors.ParamErrorCode: errorCode.Bytes(),
	}))

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
