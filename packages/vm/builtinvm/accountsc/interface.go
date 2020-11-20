package accountsc

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/vm/builtinvm/util"
	"github.com/iotaledger/wasp/packages/vm/contract"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const (
	Name        = "accounts"
	Version     = "0.1"
	Description = "Chain account ledger contract"
)

var (
	Interface = contract.ContractInterface{
		Name:        Name,
		Version:     Version,
		Description: Description,
		VMType:      "builtin",
		Functions: contract.Funcs(initialize, []contract.ContractFunctionInterface{
			contract.ViewFunc(FuncBalance, getBalance),
			contract.ViewFunc(FuncAccounts, getAccounts),
			contract.Func(FuncDeposit, deposit),
			contract.Func(FuncMove, move),
			contract.Func(FuncWithdraw, withdraw),
		}),
	}

	ProgramHash = util.BuiltinProgramHash(Name, Version)
	Hname       = util.BuiltinHname(Name, Version)
)

const (
	FuncBalance  = "balance"
	FuncDeposit  = "deposit"
	FuncMove     = "move"
	FuncWithdraw = "withdraw"
	FuncAccounts = "accounts"

	VarStateInitialized = "i"
	VarStateAllAccounts = "a"

	ParamAgentID = "a"
)

var ErrParamsAgentIDNotFound = fmt.Errorf("wrong parameters: agent ID not specified")

func GetProcessor() vmtypes.Processor {
	return &Interface
}
