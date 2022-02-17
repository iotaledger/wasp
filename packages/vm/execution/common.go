package execution

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// this file holds functions common to both context implementation (viewcontext and vmcontext)

func GetProgramBinary(ctx WaspContext, programHash hashing.HashValue) (vmtype string, binary []byte, err error) {
	vmtype, ok := ctx.Processors().Config.GetNativeProcessorType(programHash)
	if ok {
		return vmtype, nil, nil
	}
	return ctx.LocateProgram(programHash)
}

func GetEntryPointByProgHash(ctx WaspContext, targetContract, epCode iscp.Hname, progHash hashing.HashValue) iscp.VMProcessorEntryPoint {
	getBinary := func(programHash hashing.HashValue) (vmtype string, binary []byte, err error) {
		return GetProgramBinary(ctx, programHash)
	}

	proc, err := ctx.Processors().GetOrCreateProcessorByProgramHash(progHash, getBinary)
	if err != nil {
		panic(err)
	}
	ep, ok := proc.GetEntryPoint(epCode)
	if !ok {
		ctx.GasBurn(gas.BurnCodeCallTargetNotFound)
		// TODO refactor with the new errors that will be imported from vm package (currently importing vmcontext.ErrTargetEntryPointNotFound causes a loop)
		panic("entry point not found TODO REFACTOR")
		// panic(xerrors.Errorf("%v: target=(%s, %s)",
		// 	ErrTargetEntryPointNotFound, targetContract, epCode))
	}
	return ep
}
