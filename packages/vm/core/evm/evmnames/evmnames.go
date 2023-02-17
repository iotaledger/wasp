// package evmnames provides the names of EVM core contract functions and fields.
// It is separated from the evm interface to avoid import loops (the names are used
// by isc.evmOffLedgerRequest).
package evmnames

const (
	Contract = "evm"

	// EVM state
	FuncSendTransaction                     = "sendTransaction"
	FuncEstimateGas                         = "estimateGas"
	FuncGetBalance                          = "getBalance"
	FuncCallContract                        = "callContract"
	FuncGetNonce                            = "getNonce"
	FuncGetReceipt                          = "getReceipt"
	FuncGetCode                             = "getCode"
	FuncGetBlockNumber                      = "getBlockNumber"
	FuncGetBlockByNumber                    = "getBlockByNumber"
	FuncGetBlockByHash                      = "getBlockByHash"
	FuncGetTransactionByHash                = "getTransactionByHash"
	FuncGetTransactionByBlockHashAndIndex   = "getTransactionByBlockHashAndIndex"
	FuncGetTransactionByBlockNumberAndIndex = "getTransactionByBlockNumberAndIndex"
	FuncGetTransactionCountByBlockHash      = "getTransactionCountByBlockHash"
	FuncGetTransactionCountByBlockNumber    = "getTransactionCountByBlockNumber"
	FuncGetStorage                          = "getStorage"
	FuncGetLogs                             = "getLogs"
	FuncGetChainID                          = "getChainID"

	FuncRegisterERC20NativeToken              = "registerERC20NativeToken"
	FuncRegisterERC20NativeTokenOnRemoteChain = "registerERC20NativeTokenOnRemoteChain"
	FuncRegisterERC20ExternalNativeToken      = "registerERC20ExternalNativeToken"
	FuncGetERC20ExternalNativeTokenAddress    = "getERC20ExternalNativeTokenAddress"
	FuncRegisterERC721NFTCollection           = "registerERC721NFTCollection"

	// block context
	FuncOpenBlockContext  = "openBlockContext"
	FuncCloseBlockContext = "closeBlockContext"

	FieldTransaction      = "t"
	FieldCallMsg          = "c"
	FieldChainID          = "chid"
	FieldGenesisAlloc     = "g"
	FieldAddress          = "a"
	FieldKey              = "k"
	FieldAgentID          = "i"
	FieldTransactionIndex = "ti"
	FieldTransactionHash  = "h"
	FieldResult           = "r"
	FieldBlockNumber      = "bn"
	FieldBlockHash        = "bh"
	FieldGasRatio         = "w"
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
)
