// package evmnames provides the names of EVM core contract functions and fields.
// It is separated from the evm interface to avoid import loops (the names are used
// by isc.evmOffLedgerRequest).
package evmnames

const (
	Contract = "evm"

	// EVM state
	FuncSendTransaction = "sendTransaction"
	FuncCallContract    = "callContract"
	FuncGetChainID      = "getChainID"

	FuncRegisterERC20NativeToken              = "registerERC20NativeToken"
	FuncRegisterERC20NativeTokenOnRemoteChain = "registerERC20NativeTokenOnRemoteChain"
	FuncRegisterERC20ExternalNativeToken      = "registerERC20ExternalNativeToken"
	FuncGetERC20ExternalNativeTokenAddress    = "getERC20ExternalNativeTokenAddress"
	FuncGetERC721CollectionAddress            = "getERC721CollectionAddress"
	FuncRegisterERC721NFTCollection           = "registerERC721NFTCollection"

	FuncNewL1Deposit = "newL1Deposit"

	FieldTransaction      = "t"
	FieldCallMsg          = "c"
	FieldChainID          = "chid"
	FieldAddress          = "a"
	FieldAssets           = "s"
	FieldKey              = "k"
	FieldAgentID          = "i"
	FieldTransactionIndex = "ti"
	FieldTransactionHash  = "h"
	FieldResult           = "r"
	FieldBlockNumber      = "bn"
	FieldBlockHash        = "bh"
	FieldFilterQuery      = "fq"
	FieldBlockKeepAmount  = "bk"

	FieldNativeTokenID      = "N"
	FieldFoundrySN          = "fs"
	FieldTokenName          = "n"
	FieldTokenTickerSymbol  = "t"
	FieldTokenDecimals      = "d"
	FieldNFTCollectionID    = "C"
	FieldFoundryTokenScheme = "T"
	FieldTargetAddress      = "A"

	FieldAgentIDDepositOriginator = "l"
	FieldAgentIDWithdrawalTarget  = "t"
	FieldFromAddress              = "f"
	FieldToAddress                = "t"
)
