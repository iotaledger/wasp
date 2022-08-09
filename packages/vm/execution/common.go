package execution

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm"
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

func GetEntryPointByProgHash(ctx WaspContext, targetContract, epCode isc.Hname, progHash hashing.HashValue) isc.VMProcessorEntryPoint {
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
		panic(vm.ErrTargetEntryPointNotFound)
	}
	return ep
}
