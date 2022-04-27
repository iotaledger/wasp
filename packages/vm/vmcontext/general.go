package vmcontext

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func (vmctx *VMContext) ChainID() *iscp.ChainID {
	var ret iscp.ChainID
	if vmctx.task.AnchorOutput.StateIndex == 0 {
		// origin
		ret = iscp.ChainIDFromAliasID(iotago.AliasIDFromOutputID(vmctx.task.AnchorOutputID))
	} else {
		ret = iscp.ChainIDFromAliasID(vmctx.task.AnchorOutput.AliasID)
	}
	return &ret
}

func (vmctx *VMContext) ChainOwnerID() *iscp.AgentID {
	return vmctx.chainOwnerID
}

func (vmctx *VMContext) ContractAgentID() *iscp.AgentID {
	return iscp.NewAgentID(vmctx.ChainID().AsAddress(), vmctx.CurrentContractHname())
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

func (vmctx *VMContext) Params() *iscp.Params {
	return &vmctx.getCallContext().params
}

func (vmctx *VMContext) MyAgentID() *iscp.AgentID {
	return iscp.NewAgentID(vmctx.ChainID().AsAddress(), vmctx.CurrentContractHname())
}

func (vmctx *VMContext) Caller() *iscp.AgentID {
	return vmctx.getCallContext().caller
}

func (vmctx *VMContext) Timestamp() int64 {
	return vmctx.virtualState.Timestamp().Unix()
}

func (vmctx *VMContext) Entropy() hashing.HashValue {
	return vmctx.entropy
}

func (vmctx *VMContext) Request() iscp.Calldata {
	return vmctx.req
}

func (vmctx *VMContext) AccountID() *iscp.AgentID {
	hname := vmctx.CurrentContractHname()
	if corecontracts.IsCoreHname(hname) {
		return vmctx.ChainID().CommonAccount()
	}
	return iscp.NewAgentID(vmctx.task.AnchorOutput.AliasID.ToAddress(), hname)
}

func (vmctx *VMContext) AllowanceAvailable() *iscp.Allowance {
	allowance := vmctx.getCallContext().allowanceAvailable
	if allowance == nil {
		return iscp.NewEmptyAllowance()
	}
	return allowance.Clone()
}

func (vmctx *VMContext) isOnChainAccount(agentID *iscp.AgentID) bool {
	if agentID.IsNil() {
		return false
	}
	return agentID.Address().Equal(vmctx.ChainID().AsAddress())
}

func (vmctx *VMContext) isCoreAccount(agentID *iscp.AgentID) bool {
	return vmctx.isOnChainAccount(agentID) && corecontracts.IsCoreHname(agentID.Hname())
}

// targetAccountExists check if there's an account with non-zero balance,
// or it is an existing smart contract
func (vmctx *VMContext) targetAccountExists(agentID *iscp.AgentID) bool {
	if agentID.Equals(vmctx.ChainID().CommonAccount()) {
		return true
	}
	accountExists := false
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		accountExists = accounts.AccountExists(s, agentID)
	})
	if accountExists {
		return true
	}
	// it may be a smart contract with 0 balance
	if !vmctx.isOnChainAccount(agentID) {
		return false
	}
	vmctx.callCore(root.Contract, func(s kv.KVStore) {
		accountExists = root.ContractExists(s, agentID.Hname())
	})
	return accountExists
}

func (vmctx *VMContext) spendAllowedBudget(toSpend *iscp.Allowance) {
	if !vmctx.getCallContext().allowanceAvailable.SpendFromBudget(toSpend) {
		panic(accounts.ErrNotEnoughAllowance)
	}
}

// TransferAllowedFunds transfers funds within the budget set by the Allowance() to the existing target account on chain
func (vmctx *VMContext) TransferAllowedFunds(target *iscp.AgentID, forceOpenAccount bool, transfer ...*iscp.Allowance) *iscp.Allowance {
	if vmctx.isCoreAccount(target) {
		// if the target is one of core contracts, assume target is the common account
		target = vmctx.ChainID().CommonAccount()
	} else {
		// check if target exists, if it is not forced
		// forceOpenAccount == true it is not checked and the transfer will occur even if the target does not exist
		if !forceOpenAccount && !vmctx.targetAccountExists(target) {
			panic(vm.ErrTransferTargetAccountDoesNotExists)
		}
	}

	var toMove *iscp.Allowance
	if len(transfer) == 0 {
		toMove = vmctx.AllowanceAvailable()
	} else {
		toMove = transfer[0]
	}

	vmctx.spendAllowedBudget(toMove) // panics if not enough

	caller := vmctx.Caller() // have to take it here because callCore changes that
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		accounts.MoveBetweenAccounts(s, caller, target, toMove.Assets, toMove.NFTs)
	})
	return vmctx.AllowanceAvailable()
}

func (vmctx *VMContext) StateAnchor() *iscp.StateAnchor {
	var nilAliasID iotago.AliasID
	blockset, err := vmctx.task.AnchorOutput.FeatureBlocks().Set()
	if err != nil {
		panic(err)
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
		StateController:      vmctx.task.AnchorOutput.StateController(),
		GovernanceController: vmctx.task.AnchorOutput.GovernorAddress(),
		StateIndex:           vmctx.task.AnchorOutput.StateIndex,
		OutputID:             vmctx.task.AnchorOutputID,
		StateData:            vmctx.task.AnchorOutput.StateMetadata,
		Deposit:              vmctx.task.AnchorOutput.Amount,
		NativeTokens:         vmctx.task.AnchorOutput.NativeTokens,
	}
}

// DeployContract deploys contract by its program hash with the name and description specific to the instance
func (vmctx *VMContext) DeployContract(programHash hashing.HashValue, name, description string, initParams dict.Dict) {
	vmctx.Debugf("vmcontext.DeployContract: %s, name: %s, dscr: '%s'", programHash.String(), name, description)

	// calling root contract from another contract to install contract
	// adding parameters specific to deployment
	par := initParams.Clone()
	par.Set(root.ParamProgramHash, codec.EncodeHashValue(programHash))
	par.Set(root.ParamName, codec.EncodeString(name))
	par.Set(root.ParamDescription, codec.EncodeString(description))
	vmctx.Call(root.Contract.Hname(), root.FuncDeployContract.Hname(), par, nil)
}

func (vmctx *VMContext) RegisterError(messageFormat string) *iscp.VMErrorTemplate {
	vmctx.Debugf("vmcontext.RegisterError: messageFormat: '%s'", messageFormat)

	params := dict.New()
	params.Set(errors.ParamErrorMessageFormat, codec.EncodeString(messageFormat))

	result := vmctx.Call(errors.Contract.Hname(), errors.FuncRegisterError.Hname(), params, nil)
	errorCode := codec.MustDecodeVMErrorCode(result.MustGet(errors.ParamErrorCode))

	vmctx.Debugf("vmcontext.RegisterError: errorCode: '%s'", errorCode)

	return iscp.NewVMErrorTemplate(errorCode, messageFormat)
}
