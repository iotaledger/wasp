package vmcontext

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"golang.org/x/xerrors"
)

// Call implements sandbox logic of the call between contracts on-chain
func (vmctx *VMContext) Call(targetContract, epCode iscp.Hname, params dict.Dict, allowance *iscp.Assets) (dict.Dict, error) {
	vmctx.Debugf("Call. TargetContract: %s entry point: %s", targetContract, epCode)
	if rec := vmctx.findContractByHname(targetContract); rec != nil {
		return vmctx.callByProgramHash(targetContract, epCode, params, allowance, rec.ProgramHash)
	}
	vmctx.GasBurn(gas.NotFoundTarget, gas.Storage)
	panic(xerrors.Errorf("%v: contract='%s'", ErrContractNotFound, targetContract))
}

func (vmctx *VMContext) callByProgramHash(targetContract, epCode iscp.Hname, params dict.Dict, allowance *iscp.Assets, progHash hashing.HashValue) (dict.Dict, error) {
	proc, err := vmctx.task.Processors.GetOrCreateProcessorByProgramHash(progHash, vmctx.getBinary)
	if err != nil {
		return nil, err
	}
	ep, ok := proc.GetEntryPoint(epCode)
	if !ok {
		vmctx.GasBurn(gas.NotFoundTarget, gas.CallTargetNotFound)
		panic(xerrors.Errorf("%v: target=(%s, %s)",
			ErrTargetEntryPointNotFound, targetContract, epCode))
	}

	vmctx.pushCallContext(targetContract, params, allowance)
	defer vmctx.popCallContext()

	// distinguishing between two types of entry points. Passing different types of sandboxes
	if ep.IsView() {
		if epCode == iscp.EntryPointInit {
			panic(xerrors.Errorf("'%v': target=(%s, %s)",
				ErrEntryPointCantBeAView, vmctx.req.CallTarget().Contract, epCode))
		}
		return ep.Call(NewSandboxView(vmctx))
	}
	// no view
	// prevent calling 'init' not from root contract or not while initializing root
	if epCode == iscp.EntryPointInit && targetContract != root.Contract.Hname() {
		if !vmctx.callerIsRoot() {
			panic(xerrors.Errorf("%v: target=(%s, %s)",
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

const traceStack = false

func (vmctx *VMContext) pushCallContext(contract iscp.Hname, params dict.Dict, allowance *iscp.Assets) {
	ctx := &callContext{
		caller:             vmctx.getToBeCaller(),
		contract:           contract,
		params:             params.Clone(),
		allowanceAvailable: allowance.Clone(), // we have to clone it because it will be mutated by TransferAllowedFunds
	}
	if traceStack {
		vmctx.Debugf("+++++++++++ PUSH %d, stack depth = %d caller = %s", contract, len(vmctx.callStack), ctx.caller)
	}
	vmctx.callStack = append(vmctx.callStack, ctx)
}

func (vmctx *VMContext) popCallContext() {
	if traceStack {
		vmctx.Debugf("+++++++++++ POP @ depth %d", len(vmctx.callStack))
	}
	vmctx.callStack[len(vmctx.callStack)-1] = nil // for GC
	vmctx.callStack = vmctx.callStack[:len(vmctx.callStack)-1]
}

func (vmctx *VMContext) getToBeCaller() *iscp.AgentID {
	if len(vmctx.callStack) > 0 {
		return vmctx.MyAgentID()
	}
	if vmctx.req == nil {
		// core call (e.g. saving the anchor ID)
		return vmctx.chainOwnerID
	}
	// request context
	return vmctx.req.SenderAccount()
}

func (vmctx *VMContext) getCallContext() *callContext {
	if len(vmctx.callStack) == 0 {
		panic("getCallContext: stack is empty")
	}
	return vmctx.callStack[len(vmctx.callStack)-1]
}

func (vmctx *VMContext) callCore(c *coreutil.ContractInfo, f func(s kv.KVStore)) {
	vmctx.pushCallContext(c.Hname(), nil, nil)
	defer vmctx.popCallContext()

	f(vmctx.State())
}
