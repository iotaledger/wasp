package accounts

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractAccounts)

var (
	// Funcs
	FuncDeposit                = coreutil.Func("deposit")
	FuncFoundryCreateNew       = coreutil.Func("foundryCreateNew")
	FuncFoundryDestroy         = coreutil.Func("foundryDestroy")
	FuncFoundryModifySupply    = coreutil.Func("foundryModifySupply")
	FuncMintNFT                = coreutil.Func("mintNFT")
	FuncTransferAccountToChain = coreutil.Func("transferAccountToChain")
	FuncTransferAllowanceTo    = coreutil.Func("transferAllowanceTo")
	FuncWithdraw               = coreutil.Func("withdraw")
	// TODO implement grant/claim protocol of moving ownership of the foundry
	//  Including ownership of the foundry by the common account/chain owner

	// Views
	ViewAccountFoundries             = coreutil.ViewFunc("accountFoundries")
	ViewAccountNFTAmount             = coreutil.ViewFunc("accountNFTAmount")
	ViewAccountNFTAmountInCollection = coreutil.ViewFunc("accountNFTAmountInCollection")
	ViewAccountNFTs                  = coreutil.ViewFunc("accountNFTs")
	ViewAccountNFTsInCollection      = coreutil.ViewFunc("accountNFTsInCollection")
	ViewNFTIDbyMintID                = coreutil.ViewFunc("NFTIDbyMintID")
	ViewAccounts                     = coreutil.ViewFunc("accounts")
	ViewBalance                      = coreutil.ViewFunc("balance")
	ViewBalanceBaseToken             = coreutil.ViewFunc("balanceBaseToken")
	ViewBalanceNativeToken           = coreutil.ViewFunc("balanceNativeToken")
	ViewFoundryOutput                = coreutil.ViewFunc("foundryOutput")
	ViewGetAccountNonce              = coreutil.ViewFunc("getAccountNonce")
	ViewGetNativeTokenIDRegistry     = coreutil.ViewFunc("getNativeTokenIDRegistry")
	ViewNFTData                      = coreutil.ViewFunc("nftData")
	ViewTotalAssets                  = coreutil.ViewFunc("totalAssets")
)

// request parameters
const (
	ParamAccountNonce           = "n"
	ParamAgentID                = "a"
	ParamBalance                = "B"
	ParamCollectionID           = "C"
	ParamDestroyTokens          = "y"
	ParamForceMinimumBaseTokens = "f"
	ParamFoundryOutputBin       = "b"
	ParamFoundrySN              = "s"
	ParamGasReserve             = "g"
	ParamNFTAmount              = "A"
	ParamNFTData                = "e"
	ParamNFTID                  = "z"
	ParamNFTIDs                 = "i"
	ParamNFTImmutableData       = "I"
	ParamNFTWithdrawOnMint      = "w"
	ParamMintID                 = "D"
	ParamNativeTokenID          = "N"
	ParamSupplyDeltaAbs         = "d"
	ParamTokenScheme            = "t"
)
