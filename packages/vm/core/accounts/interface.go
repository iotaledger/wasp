package accounts

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"golang.org/x/xerrors"
)

var Contract = coreutil.NewContract(coreutil.CoreContractAccounts, "Chain account ledger contract")

var (
	FuncViewBalance              = coreutil.ViewFunc("balance")
	FuncViewTotalAssets          = coreutil.ViewFunc("totalAssets")
	FuncViewAccounts             = coreutil.ViewFunc("accounts")
	FuncDeposit                  = coreutil.Func("deposit")
	FuncTransferAllowanceTo      = coreutil.Func("transferAllowanceTo")
	FuncWithdraw                 = coreutil.Func("withdraw")
	FuncHarvest                  = coreutil.Func("harvest")
	FuncGetAccountNonce          = coreutil.ViewFunc("getAccountNonce")
	FuncGetNativeTokenIDRegistry = coreutil.ViewFunc("getNativeTokenIDs")
	FuncFoundryCreateNew         = coreutil.Func("foundryCreateNew")
	FuncFoundryDestroy           = coreutil.Func("foundryDestroy")
	FuncFoundryOutput            = coreutil.ViewFunc("foundryOutput")
	FuncFoundryModifySupply      = coreutil.Func("foundryModifySupply")
	// TODO implement grant/claim protocol of moving ownership of the foundry
	//  Including ownership of the foundry by the common account/chain owner
)

const (
	// MinimumIotasOnCommonAccount can't harvest the minimum
	MinimumIotasOnCommonAccount = 3000

	// prefix for a name of a particular account
	prefixAccount = string(byte(iota) + 'A')
	// map with all accounts listed
	prefixAllAccounts
	// map of account with all on-chain totals listed
	prefixTotalL2AssetsAccount
	// prefix for the map of nonces
	prefixMaxAssumedNonceKey
	// prefix for all foundries owned by the account
	prefixAccountFoundries
	// prefixNativeTokenOutputMap a map of accounts -> foundries
	prefixNativeTokenOutputMap
	// prefixFoundryOutputRecords a map with all foundry outputs
	prefixFoundryOutputRecords
	//
	stateVarMinimumDustDepositAssumptionsBin

	ParamAgentID                   = "a"
	ParamAccountNonce              = "n"
	ParamForceMinimumIotas         = "f"
	ParamFoundrySN                 = "s"
	ParamFoundryOutputBin          = "b"
	ParamTokenScheme               = "t"
	ParamTokenTag                  = "g"
	ParamMaxSupply                 = "p"
	ParamSupplyDeltaAbs            = "d"
	ParamDestroyTokens             = "y"
	ParamDustDepositAssumptionsBin = "u"
)

var ErrDustDepositAssumptionsWrong = xerrors.New("'dust deposit assumptions' parameter not specified or wrong")

// DecodeBalances TODO move to iscp package
func DecodeBalances(balances dict.Dict) (*iscp.Assets, error) {
	return iscp.AssetsFromDict(balances)
}
