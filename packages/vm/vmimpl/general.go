package vmimpl

import (
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
)

func (reqctx *requestContext) ChainID() isc.ChainID {
	return reqctx.vm.ChainID()
}

func (vmctx *vmContext) ChainID() isc.ChainID {
	return isc.ChainID(*vmctx.task.Anchor.GetObjectID())
}

func (reqctx *requestContext) ChainInfo() *isc.ChainInfo {
	return reqctx.vm.ChainInfo()
}

func (vmctx *vmContext) ChainInfo() *isc.ChainInfo {
	return vmctx.chainInfo
}

func (reqctx *requestContext) ChainAdmin() isc.AgentID {
	return reqctx.vm.ChainAdmin()
}

func (vmctx *vmContext) ChainAdmin() isc.AgentID {
	return vmctx.chainInfo.ChainAdmin
}

func (reqctx *requestContext) CurrentContractAgentID() isc.AgentID {
	return isc.NewContractAgentID(reqctx.CurrentContractHname())
}

func (reqctx *requestContext) CurrentContractHname() isc.Hname {
	return reqctx.getCallContext().contract
}

func (reqctx *requestContext) Params() isc.CallArguments {
	return reqctx.getCallContext().params
}

func (reqctx *requestContext) Caller() isc.AgentID {
	return reqctx.getCallContext().caller
}

func (reqctx *requestContext) Timestamp() time.Time {
	return reqctx.vm.task.Timestamp
}

func (reqctx *requestContext) CurrentContractAccountID() isc.AgentID {
	hname := reqctx.CurrentContractHname()
	if corecontracts.IsCoreHname(hname) {
		return accounts.CommonAccount()
	}
	return isc.NewContractAgentID(hname)
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
	return corecontracts.IsCoreHname(contract.Hname())
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
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		if err := s.MoveBetweenAccounts(caller, target, toMove); err != nil {
			panic(vm.ErrNotEnoughFundsForAllowance)
		}
	})
	return reqctx.allowanceAvailable()
}

func (reqctx *requestContext) registerError(messageFormat string) *isc.VMErrorTemplate {
	reqctx.Debugf("vmcontext.RegisterError: messageFormat: '%s'", messageFormat)
	args := reqctx.Call(errors.FuncRegisterError.Message(messageFormat), isc.NewEmptyAssets())

	errorCode := lo.Must(errors.FuncRegisterError.DecodeOutput(args))
	reqctx.Debugf("vmcontext.RegisterError: errorCode: '%s'", errorCode)
	return isc.NewVMErrorTemplate(errorCode, messageFormat)
}
