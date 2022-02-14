package vmcontext

import (
	"runtime/debug"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"golang.org/x/xerrors"
)

func CreateVMContextForViewCall(ch chain.ChainCore) *VMContext {
	anchorOutput, anchorOutputID := ch.GetAnchorOutput()

	task := &vm.VMTask{
		AnchorOutput:       anchorOutput,
		AnchorOutputID:     *anchorOutputID,
		Processors:         ch.Processors(),
		SolidStateBaseline: ch.GlobalStateSync().GetSolidIndexBaseline(),
		VirtualStateAccess: ch.VirtualStateAccess(),
		Log:                ch.Log(),
	}

	virtualState := state.WrapMustOptimisticVirtualStateAccess(task.VirtualStateAccess, task.SolidStateBaseline)

	return &VMContext{
		blockContextCloseSeq: make([]iscp.Hname, 0),
		callStack:            make([]*callContext, 0),
		task:                 task,
		currentStateUpdate:   state.NewStateUpdate(),
		virtualState:         virtualState,
	}
}

// callViewExternal is onlu for internal usage by CallViewExternal
func (vmctx *VMContext) callViewExternal(targetContract, epCode iscp.Hname, params dict.Dict) dict.Dict {
	contractRecord := vmctx.getContractRecord(targetContract)
	ep := vmctx.getEntryPointByProgHash(targetContract, epCode, contractRecord.ProgramHash)

	if !ep.IsView() {
		panic("target entrypoint is not a view")
	}

	vmctx.pushCallContext(targetContract, params, nil)
	defer vmctx.popCallContext()

	vmctx.gasSetBudget(gas.ViewCallGasBudget)
	vmctx.gasBurnEnable(true)

	return ep.Call(NewSandboxView(vmctx))
}

// CallViewExternal is used to call a view outside of SC execution (webapi/dashboard/etc)
// implements own panic catcher
func (vmctx *VMContext) CallViewExternal(targetContract, epCode iscp.Hname, params dict.Dict) (ret dict.Dict, err error) {
	// TODO look into refactor to use the new panic catcher utility
	func() {
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			ret = nil
			switch err1 := r.(type) {
			case *kv.DBError:
				vmctx.task.Log.Panicf("DB error: %v", err1)
			case error:
				err = err1
			default:
				err = xerrors.Errorf("viewcontext: panic in VM: %v", err1)
			}
			vmctx.task.Log.Debugf("CallView: %v", err)
			vmctx.task.Log.Debugf(string(debug.Stack()))
		}()
		ret = vmctx.callViewExternal(targetContract, epCode, params)
	}()
	return ret, err
}
