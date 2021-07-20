package accounts

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractAccounts, "Chain account ledger contract")

var (
	FuncViewBalance     = coreutil.ViewFunc("balance")
	FuncViewTotalAssets = coreutil.ViewFunc("totalAssets")
	FuncViewAccounts    = coreutil.ViewFunc("accounts")
	FuncDeposit         = coreutil.Func("deposit")
	FuncWithdraw        = coreutil.Func("withdraw")
	FuncHarvest         = coreutil.Func("harvest")
)

const (
	ParamAgentID        = "a"
	ParamWithdrawColor  = "c"
	ParamWithdrawAmount = "m"
)
