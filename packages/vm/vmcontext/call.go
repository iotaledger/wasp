package vmcontext

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"golang.org/x/xerrors"
)

var (
	ErrContractNotFound          = xerrors.New("contract not found")
	ErrEntryPointNotFound        = xerrors.New("entry point not found")
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
		return nil, ErrEntryPointNotFound
	}
	// distinguishing between two types of entry points. Passing different types of sandboxes
	if ep.IsView() {
		if epCode == iscp.EntryPointInit {
			return nil, ErrInitEntryPointCantBeAView
		}
		// passing nil as transfer: calling the view should not have effect on chain ledger
		if err := vmctx.pushCallContextAndMoveAssets(targetContract, params, nil); err != nil {
			return nil, err
		}
		defer vmctx.popCallContext()

		return ep.Call(NewSandboxView(vmctx))
	}
	if err := vmctx.pushCallContextAndMoveAssets(targetContract, params, transfer); err != nil {
		return nil, err
	}
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
		return nil, ErrEntryPointNotFound
	}
	if ep.IsView() {
		return nil, xerrors.New("non-view entry point expected")
	}
	if err := vmctx.pushCallContextAndMoveAssets(targetContract, params, transfer); err != nil {
		return nil, err
	}
	defer vmctx.popCallContext()

	// prevent calling 'init' not from root contract or not while initializing root
	if epCode == iscp.EntryPointInit && targetContract != root.Contract.Hname() {
		if !vmctx.callerIsRoot() {
			return nil, xerrors.New("attempt to callByProgramHash init not from the root contract")
		}
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
