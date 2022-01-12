package vmcontext

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/accounts/commonaccount"
	"github.com/iotaledger/wasp/packages/vm/core/root"
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

func (vmctx *VMContext) Request() iscp.Calldata {
	return vmctx.req
}

func (vmctx *VMContext) AccountID() *iscp.AgentID {
	hname := vmctx.CurrentContractHname()
	if commonaccount.IsCoreHname(hname) {
		return commonaccount.Get(vmctx.ChainID())
	}
	return iscp.NewAgentID(vmctx.task.AnchorOutput.AliasID.ToAddress(), hname)
}

func (vmctx *VMContext) AllowanceAvailable() *iscp.Assets {
	return vmctx.getCallContext().allowanceRemains
}

func (vmctx *VMContext) isOnChainAccount(agentID *iscp.AgentID) bool {
	if agentID.IsNil() {
		return false
	}
	return agentID.Address().Equal(vmctx.ChainID().AsAddress())
}

func (vmctx *VMContext) isCoreAccount(agentID *iscp.AgentID) bool {
	return vmctx.isOnChainAccount(agentID) && commonaccount.IsCoreHname(agentID.Hname())
}

// targetAccountExists check if there's an account with non-zero balance,
// or it is an existing smart contract
func (vmctx *VMContext) targetAccountExists(agentID *iscp.AgentID) bool {
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

func (vmctx *VMContext) spendAllowedBudget(toSpend *iscp.Assets) {
	if !vmctx.getCallContext().allowanceRemains.SpendFromBudget(toSpend) {
		panic(accounts.ErrNotEnoughAllowance)
	}
}

// TransferAllowedFunds transfers funds withing the budget set by the Allowance() to the existing target account on chain
func (vmctx *VMContext) TransferAllowedFunds(target *iscp.AgentID, assets ...*iscp.Assets) {
	if vmctx.isCoreAccount(target) {
		// if the target is one of core contracts, assume target is the common account
		target = commonaccount.Get(vmctx.ChainID())
	} else {
		// check if target exists
		if !vmctx.targetAccountExists(target) {
			panic(ErrTransferTargetAccountDoesNotExists)
		}
	}

	var assetsToMove *iscp.Assets
	if len(assets) == 0 {
		assetsToMove = vmctx.AllowanceAvailable().Clone()
	} else {
		assetsToMove = assets[0]
	}

	vmctx.spendAllowedBudget(assetsToMove) // panics if not enough

	caller := vmctx.Caller() // have to take it here because callCore changes that
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		accounts.MoveBetweenAccounts(s, caller, target, assetsToMove)
	})
}

func (vmctx *VMContext) StateAnchor() *iscp.StateAnchor {
	sd, err := iscp.StateDataFromBytes(vmctx.task.AnchorOutput.StateMetadata)
	if err != nil {
		panic(err)
	}
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
		StateController:      vmctx.task.AnchorOutput.StateController,
		GovernanceController: vmctx.task.AnchorOutput.GovernanceController,
		StateIndex:           vmctx.task.AnchorOutput.StateIndex,
		OutputID:             vmctx.task.AnchorOutputID.ID(),
		StateData:            sd,
		Deposit:              vmctx.task.AnchorOutput.Amount,
		NativeTokens:         vmctx.task.AnchorOutput.NativeTokens,
	}
}

// DeployContract deploys contract by its program hash with the name and description specific to the instance
func (vmctx *VMContext) DeployContract(programHash hashing.HashValue, name, description string, initParams dict.Dict) error {
	vmctx.Debugf("vmcontext.DeployContract: %s, name: %s, dscr: '%s'", programHash.String(), name, description)

	// calling root contract from another contract to install contract
	// adding parameters specific to deployment
	par := initParams.Clone()
	par.Set(root.ParamProgramHash, codec.EncodeHashValue(programHash))
	par.Set(root.ParamName, codec.EncodeString(name))
	par.Set(root.ParamDescription, codec.EncodeString(description))
	_, err := vmctx.Call(root.Contract.Hname(), root.FuncDeployContract.Hname(), par, nil)
	return err
}
