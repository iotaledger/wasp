package vmcontext

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accounts"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func (vmctx *VMContext) GetIncoming() coretypes.ColoredBalances {
	return vmctx.getCallContext().transfer
}

func (vmctx *VMContext) GetBalance(col balance.Color) int64 {
	return vmctx.getBalance(col)
}

func (vmctx *VMContext) GetMyBalances() coretypes.ColoredBalances {
	return vmctx.getMyBalances()
}

func (vmctx *VMContext) DoMoveBalance(target coretypes.AgentID, col balance.Color, amount int64) bool {
	return vmctx.moveBalance(target, col, amount)
}

// TransferToAddress includes output of colored tokens into the transaction
// i.e. it is a transfer of tokens from chain to layer 1 ledger
func (vmctx *VMContext) TransferToAddress(targetAddr address.Address, transfer coretypes.ColoredBalances) bool {
	privileged := vmctx.CurrentContractHname() == accounts.Interface.Hname()
	fmt.Printf("TransferToAddress: %s privileged = %v\n", targetAddr.String(), privileged)
	if !privileged {
		// if caller is accounts, it must debit from account by itself
		agentID := vmctx.MyAgentID()
		vmctx.pushCallContext(accounts.Interface.Hname(), nil, nil) // create local context for the state
		defer vmctx.popCallContext()

		if !accounts.DebitFromAccount(vmctx.State(), agentID, transfer) {
			return false
		}
	}
	return vmctx.txBuilder.TransferToAddress(targetAddr, transfer) == nil
}

// TransferCrossChain moves the whole transfer to another chain to the target account
// 1 request token should not be included into the transfer parameter but it is transferred automatically
// as a request token from the caller's account on top of specified transfer. It will be taken as a fee or accrued
// to the caller's account
// node fee is deducted from the transfer by the target
func (vmctx *VMContext) TransferCrossChain(targetAgentID coretypes.AgentID, targetChainID coretypes.ChainID, transfer coretypes.ColoredBalances) bool {
	if targetChainID == vmctx.ChainID() {
		return false
	}
	// the transfer is performed by the accountsc contract on another chain
	// it deposits received funds to the target on behalf of the caller
	par := dict.New()
	par.Set(accounts.ParamAgentID, codec.EncodeAgentID(targetAgentID))
	return vmctx.PostRequest(vmtypes.NewRequestParams{
		TargetContractID: coretypes.NewContractID(targetChainID, accounts.Interface.Hname()),
		EntryPoint:       coretypes.Hn(accounts.FuncDeposit),
		Params:           par,
		Transfer:         transfer,
	})
}
