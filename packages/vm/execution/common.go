package execution

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

var ErrTooManyNFTsInAllowance = coreerrors.Register("expected at most 1 NFT in allowance").Create()

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
		panic(vm.ErrTargetEntryPointNotFound)
	}
	return ep
}
