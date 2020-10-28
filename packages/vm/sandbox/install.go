package sandbox

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (vctx *sandbox) InstallProgram(vmtype string, programBinary []byte) (coretypes.Uint16, error) {
	if vctx.mustCurrentContractIndex() != 0 {
		return 0, fmt.Errorf("InstallProgram: can be called only from the root contract")
	}

	newContractIndex := coretypes.Uint16(0)
	// must be installed into the map of contract processors in the current chain
	// as a mutation to the state
	// returns the index of new contracts
	// TODO not correct, just idea
	//proc, err := processors.NewProcessorFromBinaryCode(vmtype, programBinary)
	//if err != nil{
	//
	//}

	return newContractIndex, nil
}
