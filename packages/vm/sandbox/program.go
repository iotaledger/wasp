package sandbox

import (
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func (s *sandbox) InstallProgram(vmtype string, programBinary []byte, description string) (uint16, error) {
	if s.GetContractIndex() == 0 {
		return s.vmctx.InstallContract(vmtype, programBinary, description)
	}

	var newContractIndex uint16
	if s.GetContractIndex() != 0 {
		// calling root contract from another contract
		par := codec.NewCodec(dict.NewDict())
		par.SetString("vmtype", vmtype)
		par.Set("programBinary", programBinary)
		resp, err := s.CallContract(0, "newContract", par)
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
	// call from the root
	//proc, err := processors.NewProcessorFromBinary(vmtype, programBinary)
	//if err != nil{
	//	return 0, err
	//}
	//registry := s.AccessState().GetArray(root.VarContractRegistry)
	//nextIndex := registry.Len()
	//deploymenHash := processors.deploymentHash(nextIndex, programBinary, vmtype)

	return newContractIndex, nil
}

func (s *sandbox) CallContract(contractIndex uint16, funName string, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	// TODO
	// find processor and entry point
	// push index
	// call entry point
	// pop index
	return nil, nil
}
