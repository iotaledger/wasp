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
	FuncSendTo          = coreutil.Func("sendTo")
	FuncWithdraw        = coreutil.Func("withdraw")
	FuncHarvest         = coreutil.Func("harvest")
	FuncGetAccountNonce = coreutil.ViewFunc("getAccountNonce")
)

const (
	// prefix for a name of a particular account
	prefixAccount = string(byte(iota) + 'A')
	// map with all accounts listed
	prefixAllAccounts
	// map of account with all on-chain totals listed
	prefixTotalAssetsAccount
	// prefix for the map of nonces
	prefixMaxAssumedNonceKey

	ParamAgentID         = "a"
	ParamWithdrawAssetID = "c"
	ParamWithdrawAmount  = "m"
	ParamAccountNonce    = "n"
)
