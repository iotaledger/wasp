package accounts

import (
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
)

const (
	Name        = "accounts"
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
		coreutil.Func(FuncWithdrawToAddress, withdrawToAddress),
		coreutil.Func(FuncWithdrawToChain, withdrawToChain),
	})
}

const (
	FuncBalance           = "balance"
	FuncTotalAssets       = "totalAssets"
	FuncDeposit           = "deposit"
	FuncWithdrawToAddress = "withdrawToAddress"
	FuncWithdrawToChain   = "withdrawToChain"
	FuncAccounts          = "accounts"

	ParamAgentID = "a"
)
