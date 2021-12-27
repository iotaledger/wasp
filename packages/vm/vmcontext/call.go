package vmcontext

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"golang.org/x/xerrors"
)

var (
	ErrContractNotFound          = xerrors.New("contract not found")
	ErrTargetEntryPointNotFound  = xerrors.New("entry point not found")
	ErrInitEntryPointCantBeAView = xerrors.New("'init' entry point can't be a view")
	ErrCallInitNotFromRoot       = xerrors.New("attempt to call `init` not from the root contract")
)

func (vmctx *VMContext) Call(targetContract, epCode iscp.Hname, params dict.Dict, transfer *iscp.Assets) (dict.Dict, error) {
	vmctx.Debugf("Call. TargetContract: %s entry point: %s", targetContract, epCode)
	rec := vmctx.findContractByHname(targetContract)
	if rec == nil {
		return nil, ErrContractNotFound
	}
	return vmctx.callByProgramHash(targetContract, epCode, params, transfer, rec.ProgramHash)
}

func (vmctx *VMContext) callByProgramHash(targetContract, epCode iscp.Hname, params dict.Dict, transfer *iscp.Assets, progHash hashing.HashValue) (dict.Dict, error) {
	proc, err := vmctx.task.Processors.GetOrCreateProcessorByProgramHash(progHash, vmctx.getBinary)
	if err != nil {
		return nil, err
	}
	ep, ok := proc.GetEntryPoint(epCode)
	if !ok {
		return nil, ErrTargetEntryPointNotFound
	}
	// distinguishing between two types of entry points. Passing different types of sandboxes
	if ep.IsView() {
		if epCode == iscp.EntryPointInit {
			return nil, ErrInitEntryPointCantBeAView
		}
		// passing nil as transfer: calling the view should not have effect on chain ledger
		vmctx.pushCallContextAndMoveAssets(targetContract, params, nil)
		defer vmctx.popCallContext()

		return ep.Call(NewSandboxView(vmctx))
	}
	vmctx.pushCallContextAndMoveAssets(targetContract, params, transfer)
	defer vmctx.popCallContext()

	// prevent calling 'init' not from root contract or not while initializing root
	if epCode == iscp.EntryPointInit && targetContract != root.Contract.Hname() {
		if !vmctx.callerIsRoot() {
			return nil, ErrCallInitNotFromRoot
		}
	}
	return ep.Call(NewSandbox(vmctx))
}

func (vmctx *VMContext) callNonViewByProgramHash(targetContract, epCode iscp.Hname, params dict.Dict, transfer *iscp.Assets, progHash hashing.HashValue) (dict.Dict, error) {
	proc, err := vmctx.task.Processors.GetOrCreateProcessorByProgramHash(progHash, vmctx.getBinary)
	if err != nil {
		return nil, err
	}
	ep, ok := proc.GetEntryPoint(epCode)
	if !ok {
		vmctx.GasBurn(gas.NotFoundTarget)
		panic(xerrors.Errorf("%w: target contract: '%s', entry point: %s",
			ErrTargetEntryPointNotFound, vmctx.req.CallTarget().Contract.String(), epCode.String()))
	}
	if ep.IsView() {
		panic(ErrNonViewExpected)
	}
	vmctx.pushCallContextAndMoveAssets(targetContract, params, transfer)
	defer vmctx.popCallContext()

	// prevent calling 'init' not from root contract or not while initializing root
	if epCode == iscp.EntryPointInit && targetContract != root.Contract.Hname() && !vmctx.callerIsRoot() {
		panic(ErrRepeatingInitCall)
	}
	return ep.Call(NewSandbox(vmctx))
}

func (vmctx *VMContext) callerIsRoot() bool {
	caller := vmctx.Caller()
	if !caller.Address().Equal(vmctx.ChainID().AsAddress()) {
		return false
	}
	return caller.Hname() == root.Contract.Hname()
}

func (vmctx *VMContext) Params() dict.Dict {
	return vmctx.getCallContext().params
}
