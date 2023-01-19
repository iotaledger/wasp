package accounts

import (
	"errors"

	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractAccounts, "Chain account ledger contract")

var (
	// Views
	ViewBalance                      = coreutil.ViewFunc("balance")
	ViewBalanceBaseToken             = coreutil.ViewFunc("balanceBaseToken")
	ViewBalanceNativeToken           = coreutil.ViewFunc("balanceNativeToken")
	ViewTotalAssets                  = coreutil.ViewFunc("totalAssets")
	ViewAccounts                     = coreutil.ViewFunc("accounts")
	ViewGetAccountNonce              = coreutil.ViewFunc("getAccountNonce")
	ViewGetNativeTokenIDRegistry     = coreutil.ViewFunc("getNativeTokenIDRegistry")
	ViewFoundryOutput                = coreutil.ViewFunc("foundryOutput")
	ViewAccountNFTs                  = coreutil.ViewFunc("accountNFTs")
	ViewAccountNFTAmount             = coreutil.ViewFunc("accountNFTAmount")
	ViewAccountNFTsInCollection      = coreutil.ViewFunc("accountNFTsInCollection")
	ViewAccountNFTAmountInCollection = coreutil.ViewFunc("accountNFTAmountInCollection")
	ViewAccountFoundries             = coreutil.ViewFunc("accountFoundries")
	ViewNFTData                      = coreutil.ViewFunc("nftData")

	// Funcs
	FuncDeposit             = coreutil.Func("deposit")
	FuncTransferAllowanceTo = coreutil.Func("transferAllowanceTo")
	FuncWithdraw            = coreutil.Func("withdraw")
	FuncHarvest             = coreutil.Func("harvest")
	FuncFoundryCreateNew    = coreutil.Func("foundryCreateNew")
	FuncFoundryDestroy      = coreutil.Func("foundryDestroy")
	FuncFoundryModifySupply = coreutil.Func("foundryModifySupply")
	// TODO implement grant/claim protocol of moving ownership of the foundry
	//  Including ownership of the foundry by the common account/chain owner
)

const (
	// MinimumBaseTokensOnCommonAccount can't harvest the minimum
	MinimumBaseTokensOnCommonAccount = 3000

	ParamAgentID                      = "a"
	ParamAccountNonce                 = "n"
	ParamForceMinimumBaseTokens       = "f"
	ParamFoundrySN                    = "s"
	ParamFoundryOutputBin             = "b"
	ParamTokenScheme                  = "t"
	ParamSupplyDeltaAbs               = "d"
	ParamDestroyTokens                = "y"
	ParamStorageDepositAssumptionsBin = "u"
	ParamNFTAmount                    = "A"
	ParamNFTIDs                       = "i"
	ParamNFTID                        = "z"
	ParamCollectionID                 = "C"
	ParamNFTData                      = "e"
	ParamBalance                      = "B"
	ParamNativeTokenID                = "N"
)

var ErrStorageDepositAssumptionsWrong = errors.New("'storage deposit assumptions' parameter not specified or wrong")
