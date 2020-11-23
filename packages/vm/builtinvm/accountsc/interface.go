package accountsc

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
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
			contract.Func(FuncMoveOnChain, moveOnChain),
			contract.Func(FuncWithdraw, withdraw),
		}),
	}

	ProgramHash          = util.BuiltinProgramHash(Name, Version)
	Hname                = util.BuiltinHname(Name, Version)
	TotalAssetsAccountID = coretypes.NewAgentIDFromContractID(coretypes.NewContractID(coretypes.ChainID{}, Hname))
)

const (
	FuncBalance     = "balance"
	FuncDeposit     = "deposit"
	FuncMoveOnChain = "moveOnChain"
	FuncWithdraw    = "withdraw"
	FuncAccounts    = "accounts"

	VarStateInitialized = "i"
	VarStateAllAccounts = "a"

	ParamAgentID = "a"
	ParamColor   = "c"
	ParamAmount  = "t"
)

var (
	ErrParamWrongOrNotFound = fmt.Errorf("wrong parameters: agent ID is wrong or not found")
)

func GetProcessor() vmtypes.Processor {
	return &Interface
}

func ChainOwnerAgentID(chainID coretypes.ChainID) coretypes.AgentID {
	return coretypes.NewAgentIDFromContractID(coretypes.NewContractID(chainID, Hname))
}
