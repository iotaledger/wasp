package vmcontext

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/execution"
	"github.com/iotaledger/wasp/packages/vm/sandbox"
	"golang.org/x/xerrors"
)

// Call implements sandbox logic of the call between contracts on-chain
func (vmctx *VMContext) Call(targetContract, epCode isc.Hname, params dict.Dict, allowance *isc.Allowance) dict.Dict {
	vmctx.Debugf("Call: targetContract: %s entry point: %s", targetContract, epCode)
	return vmctx.callProgram(targetContract, epCode, params, allowance)
}

func (vmctx *VMContext) callProgram(targetContract, epCode isc.Hname, params dict.Dict, allowance *isc.Allowance) dict.Dict {
	contractRecord := vmctx.getOrCreateContractRecord(targetContract)
	ep := execution.GetEntryPointByProgHash(vmctx, targetContract, epCode, contractRecord.ProgramHash)

	vmctx.pushCallContext(targetContract, params, allowance)
	defer vmctx.popCallContext()

	// distinguishing between two types of entry points. Passing different types of sandboxes
	if ep.IsView() {
		return ep.Call(sandbox.NewSandboxView(vmctx))
	}
	// prevent calling 'init' not from root contract or not while initializing root
	if epCode == isc.EntryPointInit && targetContract != root.Contract.Hname() {
		if !vmctx.callerIsRoot() {
			panic(xerrors.Errorf("%v: target=(%s, %s)",
				vm.ErrRepeatingInitCall, vmctx.req.CallTarget().Contract, epCode))
		}
	}
	return ep.Call(NewSandbox(vmctx))
}

func (vmctx *VMContext) callerIsRoot() bool {
	caller, ok := vmctx.Caller().(*isc.ContractAgentID)
	if !ok {
		return false
	}
	if !caller.ChainID().Equals(vmctx.ChainID()) {
		return false
	}
	return caller.Hname() == root.Contract.Hname()
}

const traceStack = false

func (vmctx *VMContext) pushCallContext(contract isc.Hname, params dict.Dict, allowance *isc.Allowance) {
	ctx := &callContext{
		caller:   vmctx.getToBeCaller(),
		contract: contract,
		params: isc.Params{
			Dict:      params,
			KVDecoder: kvdecoder.New(params, vmctx.task.Log),
		},
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

func (vmctx *VMContext) getToBeCaller() isc.AgentID {
	if len(vmctx.callStack) > 0 {
		return vmctx.MyAgentID()
	}
	if vmctx.req == nil {
		// e.g. saving the anchor ID
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
