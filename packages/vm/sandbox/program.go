package sandbox

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// implementation of call stack of indices for inter-contract calls
// mustCurrent can panic
func (vctx *sandbox) mustCurrentContractIndex() coretypes.Uint16 {
	return vctx.contractCallStack[len(vctx.contractCallStack)]
}

func (vctx *sandbox) pushContractIndex(cindex coretypes.Uint16) {
	vctx.contractCallStack = append(vctx.contractCallStack, cindex)
}

// mustPopContractIndex may panic
func (vctx *sandbox) mustPopContractIndex() {
	vctx.contractCallStack = vctx.contractCallStack[:len(vctx.contractCallStack)-1]
}

func (vctx *sandbox) InstallProgram(vmtype string, programBinary []byte) (coretypes.Uint16, error) {
	newContractIndex := coretypes.Uint16(0)
	if vctx.mustCurrentContractIndex() != 0 {
		// calling root contract
		par := codec.NewCodec(dict.NewDict())
		par.SetString("vmtype", vmtype)
		par.Set("programBinary", programBinary)
		resp, err := vctx.CallContract(0, "newContract", par)
		if err != nil {
			return 0, err
		}
		idx, ok, err := resp.GetInt64("index")
		if err != nil || !ok {
			vctx.Panic("internal error")
			return 0, nil
		}
		return (coretypes.Uint16)(idx), nil
	}
	// TODO not finished
	// call from the root

	return newContractIndex, nil
}

func (vctx *sandbox) CallContract(contractIndex uint16, funName string, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	// TODO
	// find processor and entry point
	// push index
	// call entry point
	// pop index
	return nil, nil
}
