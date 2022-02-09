// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/util"
)

var (
	// Ethereum blockchain
	FuncGetBalance                          = coreutil.ViewFunc("getBalance")
	FuncSendTransaction                     = coreutil.Func("sendTransaction")
	FuncCallContract                        = coreutil.ViewFunc("callContract")
	FuncEstimateGas                         = coreutil.ViewFunc("estimateGas")
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

	// EVMchain SC management
	FuncSetNextOwner   = coreutil.Func("setNextOwner")
	FuncClaimOwnership = coreutil.Func("claimOwnership")
	FuncGetOwner       = coreutil.ViewFunc("getOwner")
	FuncSetGasRatio    = coreutil.Func("setGasRatio")
	FuncGetGasRatio    = coreutil.ViewFunc("getGasRatio")
	FuncSetBlockTime   = coreutil.Func("setBlockTime") // only implemented by evmlight
	FuncMintBlock      = coreutil.Func("mintBlock")    // only implemented by evmlight
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
	FieldNextEVMOwner            = "n"
	FieldGasRatio                = "w"
	FieldBlockGasLimit           = "gl"
	FieldFilterQuery             = "fq"

	// evmlight only:

	FieldBlockTime       = "bt" // uint32, avg block time in seconds
	FieldBlockKeepAmount = "bk" // int32
)

const (
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
