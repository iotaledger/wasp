package vmcontext

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
)

func (vmctx *VMContext) PushCallContext(contract coretypes.Hname, params codec.ImmutableCodec, transfer coretypes.ColoredBalances) error {
	if transfer != nil {
		if len(vmctx.callStack) == 0 {
			vmctx.doDeposit(contract, transfer)
		} else {
			if !vmctx.doTransfer(vmctx.CurrentContractHname(), contract, transfer) {
				return fmt.Errorf("PushCallContext: transfer failed")
			}
		}
	}
	vmctx.pushCallContextIntern(contract, params, transfer)
	return nil
}

func (vmctx *VMContext) pushCallContextIntern(contract coretypes.Hname, params codec.ImmutableCodec, transfer coretypes.ColoredBalances) {
	vmctx.Log().Debugf("+++++++++++ PUSH %d, stack depth = %d", contract, len(vmctx.callStack))
	var caller coretypes.AgentID
	isRequestContext := len(vmctx.callStack) == 0
	if isRequestContext {
		// request context
		caller = vmctx.reqRef.SenderAgentID()
	} else {
		caller = coretypes.NewAgentIDFromContractID(vmctx.CurrentContractID())
	}
	vmctx.callStack = append(vmctx.callStack, &callContext{
		isRequestContext: isRequestContext,
		caller:           caller,
		contract:         contract,
		params:           params,
		transfer:         transfer,
	})
}

func (vmctx *VMContext) PopCallContext() {
	vmctx.Log().Debugf("+++++++++++ POP @ depth %d", len(vmctx.callStack))
	vmctx.callStack = vmctx.callStack[:len(vmctx.callStack)-1]
}

func (vmctx *VMContext) CurrentContractID() coretypes.ContractID {
	return coretypes.NewContractID(vmctx.ChainID(), vmctx.CurrentContractHname())
}

func (vmctx *VMContext) getCallContext() *callContext {
	if len(vmctx.callStack) == 0 {
		panic("getCallContext: stack is empty")
	}
	return vmctx.callStack[len(vmctx.callStack)-1]
}

// doDeposit deposits transfer from request to internal account of the called contracts
func (vmctx *VMContext) doDeposit(toContract coretypes.Hname, transfer coretypes.ColoredBalances) {
	vmctx.pushCallContextIntern(accountsc.Hname, nil, nil) // create local context for the state
	defer vmctx.PopCallContext()

	targetAgentID := coretypes.NewAgentIDFromContractID(coretypes.NewContractID(vmctx.ChainID(), toContract))
	account := codec.NewMustCodec(vmctx).GetMap(kv.Key(targetAgentID[:]))
	transfer.Iterate(func(col balance.Color, bal int64) bool {
		var currentBalance int64
		v := account.GetAt(col[:])
		if v != nil {
			currentBalance = int64(util.Uint64From8Bytes(v))
		}
		account.SetAt(col[:], util.Uint64To8Bytes(uint64(currentBalance+bal)))
		return true
	})
}

// doTransfer transfers tokens from caller's account to target contract account
func (vmctx *VMContext) doTransfer(fromContract, toContract coretypes.Hname, transfer coretypes.ColoredBalances) bool {
	vmctx.pushCallContextIntern(accountsc.Hname, nil, nil) // create local context for the state
	defer vmctx.PopCallContext()

	sourceAgentID := coretypes.NewAgentIDFromContractID(coretypes.NewContractID(vmctx.ChainID(), fromContract))
	targetAgentID := coretypes.NewAgentIDFromContractID(coretypes.NewContractID(vmctx.ChainID(), toContract))

	state := codec.NewMustCodec(vmctx)
	sourceAccount := state.GetMap(kv.Key(sourceAgentID[:]))
	targetAccount := state.GetMap(kv.Key(targetAgentID[:]))

	// first check balances
	balancesOk := true
	transfer.Iterate(func(col balance.Color, bal int64) bool {
		var sourceBalance int64
		v := sourceAccount.GetAt(col[:])
		if v != nil {
			sourceBalance = int64(util.Uint64From8Bytes(v))
		}
		if sourceBalance < bal {
			balancesOk = false
			return false
		}
		return true
	})
	if !balancesOk {
		return false
	}
	transfer.Iterate(func(col balance.Color, bal int64) bool {
		var sourceBalance, targetBalance int64
		v := sourceAccount.GetAt(col[:])
		if v != nil {
			sourceBalance = int64(util.Uint64From8Bytes(v))
		}
		v = targetAccount.GetAt(col[:])
		if v != nil {
			targetBalance = int64(util.Uint64From8Bytes(v))
		}
		targetAccount.SetAt(col[:], util.Uint64To8Bytes(uint64(targetBalance+bal)))
		if sourceBalance == bal {
			sourceAccount.DelAt(col[:])
		} else {
			targetAccount.SetAt(col[:], util.Uint64To8Bytes(uint64(sourceBalance-bal)))
		}
		return true
	})
	return true
}
