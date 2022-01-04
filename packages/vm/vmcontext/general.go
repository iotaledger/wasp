package vmcontext

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/accounts/commonaccount"
	"golang.org/x/xerrors"
)

func (vmctx *VMContext) ChainID() *iscp.ChainID {
	var ret iscp.ChainID
	if vmctx.task.AnchorOutput.StateIndex == 0 {
		// origin
		ret = iscp.ChainIDFromAliasID(iotago.AliasIDFromOutputID(vmctx.task.AnchorOutputID.ID()))
	} else {
		ret = iscp.ChainIDFromAliasID(vmctx.task.AnchorOutput.AliasID)
	}
	return &ret
}

func (vmctx *VMContext) ChainOwnerID() *iscp.AgentID {
	return vmctx.chainOwnerID
}

func (vmctx *VMContext) ContractCreator() *iscp.AgentID {
	rec := vmctx.findContractByHname(vmctx.CurrentContractHname())
	if rec == nil {
		panic("can't find current contract")
	}
	return rec.Creator
}

func (vmctx *VMContext) CurrentContractHname() iscp.Hname {
	return vmctx.getCallContext().contract
}

func (vmctx *VMContext) Params() dict.Dict {
	return vmctx.getCallContext().params
}

func (vmctx *VMContext) MyAgentID() *iscp.AgentID {
	return iscp.NewAgentID(vmctx.ChainID().AsAddress(), vmctx.CurrentContractHname())
}

func (vmctx *VMContext) Caller() *iscp.AgentID {
	return vmctx.getCallContext().caller
}

func (vmctx *VMContext) Timestamp() int64 {
	return vmctx.virtualState.Timestamp().UnixNano()
}

func (vmctx *VMContext) Entropy() hashing.HashValue {
	return vmctx.entropy
}

func (vmctx *VMContext) Request() iscp.Request {
	return vmctx.req
}

func (vmctx *VMContext) AccountID() *iscp.AgentID {
	hname := vmctx.CurrentContractHname()
	if commonaccount.IsCoreHname(hname) {
		return commonaccount.Get(vmctx.ChainID())
	}
	return iscp.NewAgentID(vmctx.task.AnchorOutput.AliasID.ToAddress(), hname)
}

func (vmctx *VMContext) Allowance() *iscp.Assets {
	return vmctx.getCallContext().allowance
}

func (vmctx *VMContext) TakeAllowance(assets ...*iscp.Assets) {
	// getCaller() is the caller of the contract with AccountID()
	// AccountID() is the current contract id
	if len(assets) == 0 {
		// take all Allowance()
		vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
			accounts.MoveBetweenAccounts(s, vmctx.getCaller(), vmctx.AccountID(), vmctx.Allowance())
		})
		return
	}
	// check if assets fits the budget set by Allowance()
	allowed := vmctx.Allowance()
	if allowed.Iotas < assets[0].Iotas {
		panic(accounts.ErrNotEnoughFunds)
	}
	allowedTokens, err := allowed.Tokens.Set()
	if err != nil {
		panic(err)
	}
	big0 := big.NewInt(0)
	for _, requiredTokens := range assets[0].Tokens {
		if requiredTokens.Amount.Cmp(big0) <= 0 {
			panic(ErrWrongParamsInSandboxCall)
		}
		allowedAmount, ok := allowedTokens[requiredTokens.ID]
		if !ok || allowedAmount.Amount.Cmp(requiredTokens.Amount) < 0 {
			panic(accounts.ErrNotEnoughFunds)
		}
	}
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		accounts.MoveBetweenAccounts(s, vmctx.getCaller(), vmctx.AccountID(), assets[0])
	})
}

func (vmctx *VMContext) StateAnchor() *iscp.StateAnchor {
	sd, err := iscp.StateDataFromBytes(vmctx.task.AnchorOutput.StateMetadata)
	if err != nil {
		panic(xerrors.Errorf("StateAnchor: %w", err))
	}
	var nilAliasID iotago.AliasID
	blockset, err := vmctx.task.AnchorOutput.FeatureBlocks().Set()
	if err != nil {
		panic(xerrors.Errorf("StateAnchor: %w", err))
	}
	senderBlock := blockset.SenderFeatureBlock()
	var sender iotago.Address
	if senderBlock != nil {
		sender = senderBlock.Address
	}
	return &iscp.StateAnchor{
		ChainID:              *vmctx.ChainID(),
		Sender:               sender,
		IsOrigin:             vmctx.task.AnchorOutput.AliasID == nilAliasID,
		StateController:      vmctx.task.AnchorOutput.StateController,
		GovernanceController: vmctx.task.AnchorOutput.GovernanceController,
		StateIndex:           vmctx.task.AnchorOutput.StateIndex,
		OutputID:             vmctx.task.AnchorOutputID.ID(),
		StateData:            sd,
		Deposit:              vmctx.task.AnchorOutput.Amount,
		NativeTokens:         vmctx.task.AnchorOutput.NativeTokens,
	}
}
