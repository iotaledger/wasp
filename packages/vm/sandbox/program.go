package sandbox

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func (s *sandbox) InstallProgram(vmtype string, programBinary []byte, description string) (uint16, error) {
	if s.GetContractIndex() == 0 {
		return s.vmctx.InstallContract(vmtype, programBinary, description)
	}
	// calling root contract from another contract
	par := codec.NewCodec(dict.NewDict())
	par.SetString("vmtype", vmtype)
	par.Set("programBinary", programBinary)
	resp, err := s.CallContract(0, "deployContract", par)
	if err != nil {
		return 0, err
	}
	idx, ok, err := resp.GetInt64("index")
	if err != nil || !ok {
		s.Panic("internal error")
		return 0, nil
	}
	return (uint16)(idx), nil
}

func (s *sandbox) CallContract(contractIndex uint16, funName string, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	epCode := coretypes.NewEntryPointCodeFromFunctionName(funName)
	// TODO budget
	return s.vmctx.CallContract(contractIndex, epCode, params, nil)
}
