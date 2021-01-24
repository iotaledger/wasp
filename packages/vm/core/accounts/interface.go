package accounts

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/contract"
)

const (
	Name        = "accounts"
	description = "Chain account ledger contract"
)

var (
	Interface = &contract.ContractInterface{
		Name:        Name,
		Description: description,
		ProgramHash: hashing.HashStrings(Name),
	}
)

func init() {
	Interface.WithFunctions(initialize, []contract.ContractFunctionInterface{
		contract.ViewFunc(FuncBalance, getBalance),
		contract.ViewFunc(FuncTotalAssets, getTotalAssets),
		contract.ViewFunc(FuncAccounts, getAccounts),
		contract.Func(FuncDeposit, deposit),
		contract.Func(FuncWithdrawToAddress, withdrawToAddress),
		contract.Func(FuncWithdrawToChain, withdrawToChain),
		contract.Func(FuncAllow, allow),
	})
}

const (
	FuncBalance     = "balance"
	FuncTotalAssets = "totalAssets"
	FuncDeposit     = "deposit"

	FuncAllow             = "allow"
	FuncWithdrawToAddress = "withdrawToAddress"
	FuncWithdrawToChain   = "withdrawToChain"
	FuncAccounts          = "accounts"

	ParamAgentID = "a"
	ParamAmount  = "t"
)

var (
	ErrParamWrongOrNotFound = fmt.Errorf("wrong parameters: agent ID is wrong or not found")
)
