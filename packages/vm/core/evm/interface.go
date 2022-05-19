// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/util"
)

var Contract = coreutil.NewContract("evm", "EVM contract")

var (
	// EVM state
	FuncGetBalance                          = coreutil.ViewFunc("getBalance")
	FuncSendTransaction                     = coreutil.Func("sendTransaction")
	FuncCallContract                        = coreutil.ViewFunc("callContract")
	FuncGetNonce                            = coreutil.ViewFunc("getNonce")
	FuncGetReceipt                          = coreutil.ViewFunc("getReceipt")
	FuncGetCode                             = coreutil.ViewFunc("getCode")
	FuncGetBlockNumber                      = coreutil.ViewFunc("getBlockNumber")
	FuncGetBlockByNumber                    = coreutil.ViewFunc("getBlockByNumber")
	FuncGetBlockByHash                      = coreutil.ViewFunc("getBlockByHash")
	FuncGetTransactionByHash                = coreutil.ViewFunc("getTransactionByHash")
	FuncGetTransactionByBlockHashAndIndex   = coreutil.ViewFunc("getTransactionByBlockHashAndIndex")
	FuncGetTransactionByBlockNumberAndIndex = coreutil.ViewFunc("getTransactionByBlockNumberAndIndex")
	FuncGetTransactionCountByBlockHash      = coreutil.ViewFunc("getTransactionCountByBlockHash")
	FuncGetTransactionCountByBlockNumber    = coreutil.ViewFunc("getTransactionCountByBlockNumber")
	FuncGetStorage                          = coreutil.ViewFunc("getStorage")
	FuncGetLogs                             = coreutil.ViewFunc("getLogs")

	// evm SC management
	FuncSetGasRatio  = coreutil.Func("setGasRatio")
	FuncGetGasRatio  = coreutil.ViewFunc("getGasRatio")
	FuncSetBlockTime = coreutil.Func("setBlockTime")
	FuncMintBlock    = coreutil.Func("mintBlock")
)

const (
	FieldChainID                 = "chid"
	FieldGenesisAlloc            = "g"
	FieldAddress                 = "a"
	FieldKey                     = "k"
	FieldAgentID                 = "i"
	FieldTransaction             = "tx"
	FieldTransactionIndex        = "ti"
	FieldTransactionHash         = "h"
	FieldTransactionData         = "t"
	FieldTransactionDataBlobHash = "th"
	FieldCallArguments           = "c"
	FieldResult                  = "r"
	FieldBlockNumber             = "bn"
	FieldBlockHash               = "bh"
	FieldCallMsg                 = "c"
	FieldGasRatio                = "w"
	FieldBlockGasLimit           = "gl"
	FieldFilterQuery             = "fq"
	FieldBlockTime               = "bt" // uint32, avg block time in seconds
	FieldBlockKeepAmount         = "bk" // int32
)

const (
	// TODO shouldn't this be different between chain, to prevent replay attacks? (maybe derived from ISC ChainID)
	DefaultChainID = 1074 // IOTA -- get it?

	BlockGasLimitDefault = uint64(15000000)

	BlockKeepAll           = -1
	BlockKeepAmountDefault = BlockKeepAll
)

var (
	// Gas is charged in iotas, not ETH
	GasPrice = big.NewInt(0)

	// <ISC gas> = <EVM Gas> * <A> / <B>
	DefaultGasRatio = util.Ratio32{A: 1, B: 1}
)

func ISCGasBudgetToEVM(iscGasBudget uint64, gasRatio util.Ratio32) uint64 {
	// EVM gas budget = floor(ISC gas budget * B / A)
	return gasRatio.YFloor64(iscGasBudget)
}

func EVMGasToISC(evmGas uint64, gasRatio util.Ratio32) uint64 {
	// ISC gas burned = ceil(EVM gas * A / B)
	return gasRatio.XCeil64(evmGas)
}
