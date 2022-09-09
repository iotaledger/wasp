package vmcontext

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func (vmctx *VMContext) ChainID() *isc.ChainID {
	var ret isc.ChainID
	if vmctx.task.AnchorOutput.StateIndex == 0 {
		// origin
		ret = isc.ChainIDFromAliasID(iotago.AliasIDFromOutputID(vmctx.task.AnchorOutputID))
	} else {
		ret = isc.ChainIDFromAliasID(vmctx.task.AnchorOutput.AliasID)
	}
	return &ret
}

func (vmctx *VMContext) ChainOwnerID() isc.AgentID {
	return vmctx.chainOwnerID
}

func (vmctx *VMContext) AgentID() isc.AgentID {
	return isc.NewContractAgentID(vmctx.ChainID(), vmctx.CurrentContractHname())
}

func (vmctx *VMContext) CurrentContractHname() isc.Hname {
	return vmctx.getCallContext().contract
}

func (vmctx *VMContext) Params() *isc.Params {
	return &vmctx.getCallContext().params
}

func (vmctx *VMContext) MyAgentID() isc.AgentID {
	return isc.NewContractAgentID(vmctx.ChainID(), vmctx.CurrentContractHname())
}

func (vmctx *VMContext) Caller() isc.AgentID {
	return vmctx.getCallContext().caller
}

func (vmctx *VMContext) Timestamp() time.Time {
	return vmctx.virtualState.Timestamp()
}

func (vmctx *VMContext) Entropy() hashing.HashValue {
	return vmctx.entropy
}

func (vmctx *VMContext) Request() isc.Calldata {
	return vmctx.req
}

func (vmctx *VMContext) AccountID() isc.AgentID {
	hname := vmctx.CurrentContractHname()
	if corecontracts.IsCoreHname(hname) {
		return vmctx.ChainID().CommonAccount()
	}
	return isc.NewContractAgentID(vmctx.ChainID(), hname)
}

func (vmctx *VMContext) AllowanceAvailable() *isc.Allowance {
	allowance := vmctx.getCallContext().allowanceAvailable
	if allowance == nil {
		return isc.NewEmptyAllowance()
	}
	return allowance.Clone()
}

func (vmctx *VMContext) isOnChainAccount(agentID isc.AgentID) bool {
	return vmctx.ChainID().IsSameChain(agentID)
}

func (vmctx *VMContext) IsCoreAccount(agentID isc.AgentID) bool {
	contract, ok := agentID.(*isc.ContractAgentID)
	if !ok {
		return false
	}
	return contract.ChainID().Equals(vmctx.ChainID()) && corecontracts.IsCoreHname(contract.Hname())
}

// targetAccountExists check if there's an account with non-zero balance,
// or it is an existing smart contract
func (vmctx *VMContext) targetAccountExists(agentID isc.AgentID) bool {
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
	hname, _ := isc.HnameFromAgentID(agentID)
	vmctx.callCore(root.Contract, func(s kv.KVStore) {
		accountExists = root.ContractExists(s, hname)
	})
	return accountExists
}

func (vmctx *VMContext) spendAllowedBudget(toSpend *isc.Allowance) {
	if !vmctx.getCallContext().allowanceAvailable.SpendFromBudget(toSpend) {
		panic(accounts.ErrNotEnoughAllowance)
	}
}

// TransferAllowedFunds transfers funds within the budget set by the Allowance() to the existing target account on chain
func (vmctx *VMContext) TransferAllowedFunds(target isc.AgentID, forceOpenAccount bool, transfer ...*isc.Allowance) *isc.Allowance {
	if vmctx.IsCoreAccount(target) {
		// if the target is one of core contracts, assume target is the common account
		target = vmctx.ChainID().CommonAccount()
	} else if !forceOpenAccount && !vmctx.targetAccountExists(target) {
		// check if target exists, if it is not forced
		// forceOpenAccount == true it is not checked and the transfer will occur even if the target does not exist
		panic(vm.ErrTransferTargetAccountDoesNotExists)
	}

	var toMove *isc.Allowance
	if len(transfer) == 0 {
		toMove = vmctx.AllowanceAvailable()
	} else {
		toMove = transfer[0]
	}

	vmctx.spendAllowedBudget(toMove) // panics if not enough

	caller := vmctx.Caller() // have to take it here because callCore changes that

	// if the caller is a core contract, funds should be taken from the common account
	if vmctx.IsCoreAccount(caller) {
		caller = vmctx.ChainID().CommonAccount()
	}
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		if !accounts.MoveBetweenAccounts(s, caller, target, toMove.Assets, toMove.NFTs) {
			panic(vm.ErrNotEnoughFundsForAllowance)
		}
	})
	return vmctx.AllowanceAvailable()
}

func (vmctx *VMContext) StateAnchor() *isc.StateAnchor {
	var nilAliasID iotago.AliasID
	blockset := vmctx.task.AnchorOutput.FeatureSet()
	senderBlock := blockset.SenderFeature()
	var sender iotago.Address
	if senderBlock != nil {
		sender = senderBlock.Address
	}
	return &isc.StateAnchor{
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

func (vmctx *VMContext) RegisterError(messageFormat string) *isc.VMErrorTemplate {
	vmctx.Debugf("vmcontext.RegisterError: messageFormat: '%s'", messageFormat)

	params := dict.New()
	params.Set(errors.ParamErrorMessageFormat, codec.EncodeString(messageFormat))

	result := vmctx.Call(errors.Contract.Hname(), errors.FuncRegisterError.Hname(), params, nil)
	errorCode := codec.MustDecodeVMErrorCode(result.MustGet(errors.ParamErrorCode))

	vmctx.Debugf("vmcontext.RegisterError: errorCode: '%s'", errorCode)

	return isc.NewVMErrorTemplate(errorCode, messageFormat)
}
