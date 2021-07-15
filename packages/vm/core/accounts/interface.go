package accounts

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

const (
	Name        = coreutil.CoreContractAccounts
	description = "Chain account ledger contract"
)

var Interface = &coreutil.ContractInterface{
	Name:        Name,
	Description: description,
	ProgramHash: hashing.HashStrings(Name),
}

func init() {
	Interface.WithFunctions(initialize, []coreutil.ContractFunctionInterface{
		coreutil.ViewFunc(FuncViewBalance, viewBalance),
		coreutil.ViewFunc(FuncViewTotalAssets, viewTotalAssets),
		coreutil.ViewFunc(FuncViewAccounts, viewAccounts),
		coreutil.Func(FuncDeposit, deposit),
		coreutil.Func(FuncWithdraw, withdraw),
		coreutil.Func(FuncHarvest, harvest),
	})
}

const (
	FuncViewBalance     = "balance"
	FuncViewTotalAssets = "totalAssets"
	FuncViewAccounts    = "accounts"
	FuncDeposit         = "deposit"
	FuncWithdraw        = "withdraw"
	FuncHarvest         = "harvest"

	ParamAgentID        = "a"
	ParamWithdrawColor  = "c"
	ParamWithdrawAmount = "m"
)
