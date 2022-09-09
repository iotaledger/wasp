// package evmnames provides the names of EVM core contract functions and fields.
// It is separated from the evm interface to avoid import loops (the names are used
// by package isc (EVMOffLedgerRequestData)).
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

	// evm SC management
	FuncSetGasRatio = "setGasRatio"
	FuncGetGasRatio = "getGasRatio"

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
	FieldBlockGasLimit    = "gl"
	FieldFilterQuery      = "fq"
	FieldBlockKeepAmount  = "bk" // int32
)
