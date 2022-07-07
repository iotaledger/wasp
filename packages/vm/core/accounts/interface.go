package accounts

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"golang.org/x/xerrors"
)

var Contract = coreutil.NewContract(coreutil.CoreContractAccounts, "Chain account ledger contract")

var (
	// Views
	ViewBalance                  = coreutil.ViewFunc("balance")
	ViewBalanceBaseToken         = coreutil.ViewFunc("balanceBaseToken")
	ViewBalanceNativeToken       = coreutil.ViewFunc("balanceNativeToken")
	ViewTotalAssets              = coreutil.ViewFunc("totalAssets")
	ViewAccounts                 = coreutil.ViewFunc("accounts")
	ViewGetAccountNonce          = coreutil.ViewFunc("getAccountNonce")
	ViewGetNativeTokenIDRegistry = coreutil.ViewFunc("getNativeTokenIDRegistry")
	ViewFoundryOutput            = coreutil.ViewFunc("foundryOutput")
	ViewAccountNFTs              = coreutil.ViewFunc("accountNFTs")
	ViewNFTData                  = coreutil.ViewFunc("nftData")

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
	// prefixNFTOutput Records a map with all NFT outputs
	prefixNFTOutputRecords
	// prefixNFTData Records a map with all NFT data (issuer/metadata)
	prefixNFTData
	//
	stateVarMinimumDustDepositAssumptionsBin

	ParamAgentID                   = "a"
	ParamAccountNonce              = "n"
	ParamForceMinimumBaseTokens    = "f"
	ParamFoundrySN                 = "s"
	ParamFoundryOutputBin          = "b"
	ParamTokenScheme               = "t"
	ParamSupplyDeltaAbs            = "d"
	ParamDestroyTokens             = "y"
	ParamDustDepositAssumptionsBin = "u"
	ParamForceOpenAccount          = "c"
	ParamNFTIDs                    = "i"
	ParamNFTID                     = "z"
	ParamNFTData                   = "e"
	ParamBalance                   = "B"
	ParamNativeTokenID             = "N"
)

var ErrDustDepositAssumptionsWrong = xerrors.New("'dust deposit assumptions' parameter not specified or wrong")
