package vmimpl

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

func (reqctx *requestContext) ChainID() isc.ChainID {
	return reqctx.vm.ChainID()
}

func (vmctx *vmContext) ChainID() isc.ChainID {
	var chainID isc.ChainID
	if vmctx.task.AnchorOutput.StateIndex == 0 {
		// origin
		chainID = isc.ChainIDFromAliasID(iotago.AliasIDFromOutputID(vmctx.task.AnchorOutputID))
	} else {
		chainID = isc.ChainIDFromAliasID(vmctx.task.AnchorOutput.AliasID)
	}
	return chainID
}

func (reqctx *requestContext) ChainInfo() *isc.ChainInfo {
	return reqctx.vm.ChainInfo()
}

func (vmctx *vmContext) ChainInfo() *isc.ChainInfo {
	return vmctx.chainInfo
}

func (reqctx *requestContext) ChainOwnerID() isc.AgentID {
	return reqctx.vm.ChainOwnerID()
}

func (vmctx *vmContext) ChainOwnerID() isc.AgentID {
	return vmctx.chainInfo.ChainOwnerID
}

func (reqctx *requestContext) CurrentContractAgentID() isc.AgentID {
	return isc.NewContractAgentID(reqctx.vm.ChainID(), reqctx.CurrentContractHname())
}

func (reqctx *requestContext) CurrentContractHname() isc.Hname {
	return reqctx.getCallContext().contract
}

func (reqctx *requestContext) Params() *isc.Params {
	return &reqctx.getCallContext().params
}

func (reqctx *requestContext) Caller() isc.AgentID {
	return reqctx.getCallContext().caller
}

func (reqctx *requestContext) Timestamp() time.Time {
	return reqctx.vm.task.TimeAssumption
}

func (reqctx *requestContext) CurrentContractAccountID() isc.AgentID {
	hname := reqctx.CurrentContractHname()
	if corecontracts.IsCoreHname(hname) {
		return accounts.CommonAccount()
	}
	return isc.NewContractAgentID(reqctx.vm.ChainID(), hname)
}

func (reqctx *requestContext) allowanceAvailable() *isc.Assets {
	allowance := reqctx.getCallContext().allowanceAvailable
	if allowance == nil {
		return isc.NewEmptyAssets()
	}
	return allowance.Clone()
}

func (vmctx *vmContext) isCoreAccount(agentID isc.AgentID) bool {
	contract, ok := agentID.(*isc.ContractAgentID)
	if !ok {
		return false
	}
	return contract.ChainID().Equals(vmctx.ChainID()) && corecontracts.IsCoreHname(contract.Hname())
}

func (reqctx *requestContext) spendAllowedBudget(toSpend *isc.Assets) {
	if !reqctx.getCallContext().allowanceAvailable.Spend(toSpend) {
		panic(accounts.ErrNotEnoughAllowance)
	}
}

// TransferAllowedFunds transfers funds within the budget set by the Allowance() to the existing target account on chain
func (reqctx *requestContext) transferAllowedFunds(target isc.AgentID, transfer ...*isc.Assets) *isc.Assets {
	if reqctx.vm.isCoreAccount(target) {
		// if the target is one of core contracts, assume target is the common account
		target = accounts.CommonAccount()
	}

	var toMove *isc.Assets
	if len(transfer) == 0 {
		toMove = reqctx.allowanceAvailable()
	} else {
		toMove = transfer[0]
	}

	reqctx.spendAllowedBudget(toMove) // panics if not enough

	caller := reqctx.Caller() // have to take it here because callCore changes that

	// if the caller is a core contract, funds should be taken from the common account
	if reqctx.vm.isCoreAccount(caller) {
		caller = accounts.CommonAccount()
	}
	reqctx.callCore(accounts.Contract, func(s kv.KVStore) {
		if err := accounts.MoveBetweenAccounts(s, caller, target, toMove, reqctx.ChainID()); err != nil {
			panic(vm.ErrNotEnoughFundsForAllowance)
		}
	})
	return reqctx.allowanceAvailable()
}

func (vmctx *vmContext) stateAnchor() *isc.StateAnchor {
	var nilAliasID iotago.AliasID
	blockset := vmctx.task.AnchorOutput.FeatureSet()
	senderBlock := blockset.SenderFeature()
	var sender iotago.Address
	if senderBlock != nil {
		sender = senderBlock.Address
	}
	return &isc.StateAnchor{
		ChainID:              vmctx.ChainID(),
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

// DeployContract deploys contract by its program hash with the name specific to the instance
func (reqctx *requestContext) deployContract(programHash hashing.HashValue, name string, initParams dict.Dict) {
	reqctx.Debugf("vmcontext.DeployContract: %s, name: %s", programHash.String(), name)

	// calling root contract from another contract to install contract
	// adding parameters specific to deployment
	par := initParams.Clone()
	par.Set(root.ParamProgramHash, codec.EncodeHashValue(programHash))
	par.Set(root.ParamName, codec.EncodeString(name))
	reqctx.Call(root.Contract.Hname(), root.FuncDeployContract.Hname(), par, nil)
}

func (reqctx *requestContext) registerError(messageFormat string) *isc.VMErrorTemplate {
	reqctx.Debugf("vmcontext.RegisterError: messageFormat: '%s'", messageFormat)

	params := dict.New()
	params.Set(errors.ParamErrorMessageFormat, codec.EncodeString(messageFormat))

	result := reqctx.Call(errors.Contract.Hname(), errors.FuncRegisterError.Hname(), params, nil)
	errorCode := codec.MustDecodeVMErrorCode(result.Get(errors.ParamErrorCode))

	reqctx.Debugf("vmcontext.RegisterError: errorCode: '%s'", errorCode)

	return isc.NewVMErrorTemplate(errorCode, messageFormat)
}
