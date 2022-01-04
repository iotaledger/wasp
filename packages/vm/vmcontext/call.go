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
	ErrContractNotFound         = xerrors.New("contract not found")
	ErrTargetEntryPointNotFound = xerrors.New("entry point not found")
	ErrEntryPointCantBeAView    = xerrors.New("'init' entry point can't be a view")
	ErrCallInitNotFromRoot      = xerrors.New("attempt to call `init` not from the root contract")
)

// Call implements logic of the call between contracts on-chain
func (vmctx *VMContext) Call(targetContract, epCode iscp.Hname, params dict.Dict, allowance *iscp.Assets) (dict.Dict, error) {
	vmctx.Debugf("Call. TargetContract: %s entry point: %s", targetContract, epCode)
	if rec := vmctx.findContractByHname(targetContract); rec != nil {
		return vmctx.callByProgramHash(targetContract, epCode, params, allowance, rec.ProgramHash, false)
	}
	vmctx.GasBurn(gas.NotFoundTarget)
	panic(xerrors.Errorf("%w: target contract: '%s'", ErrContractNotFound, targetContract))
}

func (vmctx *VMContext) callByProgramHash(targetContract, epCode iscp.Hname, params dict.Dict, allowance *iscp.Assets, progHash hashing.HashValue, mustNonView bool) (dict.Dict, error) {
	proc, err := vmctx.task.Processors.GetOrCreateProcessorByProgramHash(progHash, vmctx.getBinary)
	if err != nil {
		return nil, err
	}
	ep, ok := proc.GetEntryPoint(epCode)
	if !ok {
		if !ok {
			vmctx.GasBurn(gas.NotFoundTarget)
			panic(xerrors.Errorf("%w: target contract: '%s', entry point: %s",
				ErrTargetEntryPointNotFound, targetContract, epCode))
		}
	}

	vmctx.pushCallContext(targetContract, params, allowance)
	defer vmctx.popCallContext()

	// distinguishing between two types of entry points. Passing different types of sandboxes
	if ep.IsView() {
		if mustNonView || epCode == iscp.EntryPointInit {
			panic(xerrors.Errorf("%w: target contract: '%s', entry point: %s",
				ErrEntryPointCantBeAView, vmctx.req.CallTarget().Contract, epCode))
		}
		return ep.Call(NewSandboxView(vmctx))
	}
	// no view
	// prevent calling 'init' not from root contract or not while initializing root
	if epCode == iscp.EntryPointInit && targetContract != root.Contract.Hname() {
		if !vmctx.callerIsRoot() {
			panic(xerrors.Errorf("%w: target contract: '%s', entry point: %s",
				ErrRepeatingInitCall, vmctx.req.CallTarget().Contract, epCode))
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
