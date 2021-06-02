package accounts

import (
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
)

const (
	Name        = coreutil.CoreContractAccounts
	description = "Chain account ledger contract"
)

var (
	Interface = &coreutil.ContractInterface{
		Name:        Name,
		Description: description,
		ProgramHash: hashing.HashStrings(Name),
	}
)

func init() {
	Interface.WithFunctions(initialize, []coreutil.ContractFunctionInterface{
		coreutil.ViewFunc(FuncBalance, getBalance),
		coreutil.ViewFunc(FuncTotalAssets, getTotalAssets),
		coreutil.ViewFunc(FuncAccounts, getAccounts),
		coreutil.Func(FuncDeposit, deposit),
		coreutil.Func(FuncWithdraw, withdraw),
	})
}

const (
	FuncBalance     = "balance"
	FuncTotalAssets = "totalAssets"
	FuncDeposit     = "deposit"
	FuncWithdraw    = "withdraw"
	FuncAccounts    = "accounts"

	ParamAgentID = "a"
)
