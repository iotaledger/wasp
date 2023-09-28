package vmimpl

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/execution"
	"github.com/iotaledger/wasp/packages/vm/sandbox"
)

// Call implements sandbox logic of the call between contracts on-chain
func (reqctx *requestContext) Call(targetContract, epCode isc.Hname, params dict.Dict, allowance *isc.Assets) dict.Dict {
	reqctx.Debugf("Call: targetContract: %s entry point: %s", targetContract, epCode)
	return reqctx.callProgram(targetContract, epCode, params, allowance, reqctx.CurrentContractAgentID())
}

func (reqctx *requestContext) withoutGasBurn(f func()) {
	prev := reqctx.gas.burnEnabled
	reqctx.GasBurnEnable(false)
	f()
	reqctx.GasBurnEnable(prev)
}

func (reqctx *requestContext) callProgram(targetContract, epCode isc.Hname, params dict.Dict, allowance *isc.Assets, caller isc.AgentID) dict.Dict {
	// don't charge gas for finding the contract (otherwise EVM requests may not produce EVM receipt)
	var ep isc.VMProcessorEntryPoint
	reqctx.withoutGasBurn(func() {
		contractRecord := reqctx.GetContractRecord(targetContract)
		ep = execution.GetEntryPointByProgHash(reqctx, targetContract, epCode, contractRecord.ProgramHash)
	})

	reqctx.pushCallContext(targetContract, params, allowance, caller)
	defer reqctx.popCallContext()

	// distinguishing between two types of entry points. Passing different types of sandboxes
	if ep.IsView() {
		return ep.Call(sandbox.NewSandboxView(reqctx))
	}
	// prevent calling 'init' not from root contract
	if epCode == isc.EntryPointInit && !caller.Equals(isc.NewContractAgentID(reqctx.vm.ChainID(), root.Contract.Hname())) {
		panic(fmt.Errorf("%v: target=(%s, %s)", vm.ErrRepeatingInitCall, targetContract, epCode))
	}
	return ep.Call(NewSandbox(reqctx))
}

const traceStack = false

func (reqctx *requestContext) pushCallContext(contract isc.Hname, params dict.Dict, allowance *isc.Assets, caller isc.AgentID) {
	ctx := &callContext{
		caller:   caller,
		contract: contract,
		params: isc.Params{
			Dict:      params,
			KVDecoder: kvdecoder.New(params, reqctx.vm.task.Log),
		},
		allowanceAvailable: allowance.Clone(), // we have to clone it because it will be mutated by TransferAllowedFunds
	}
	if traceStack {
		reqctx.Debugf("+++++++++++ PUSH %d, stack depth = %d caller = %s", contract, len(reqctx.callStack), ctx.caller)
	}
	reqctx.callStack = append(reqctx.callStack, ctx)
}

func (reqctx *requestContext) popCallContext() {
	if traceStack {
		reqctx.Debugf("+++++++++++ POP @ depth %d", len(reqctx.callStack))
	}
	reqctx.callStack[len(reqctx.callStack)-1] = nil // for GC
	reqctx.callStack = reqctx.callStack[:len(reqctx.callStack)-1]
}

func (reqctx *requestContext) getCallContext() *callContext {
	if len(reqctx.callStack) == 0 {
		panic("getCallContext: stack is empty")
	}
	return reqctx.callStack[len(reqctx.callStack)-1]
}

func withContractState(chainState kv.KVStore, c *coreutil.ContractInfo, f func(s kv.KVStore)) {
	f(subrealm.New(chainState, kv.Key(c.Hname().Bytes())))
}

func (reqctx *requestContext) callCore(c *coreutil.ContractInfo, f func(s kv.KVStore)) {
	var caller isc.AgentID
	if len(reqctx.callStack) > 0 {
		caller = reqctx.CurrentContractAgentID()
	} else {
		caller = reqctx.req.SenderAccount()
	}
	reqctx.pushCallContext(c.Hname(), nil, nil, caller)
	defer reqctx.popCallContext()

	f(reqctx.contractStateWithGasBurn())
}
