// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/iscp/coreutil"
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
	FuncSetNextOwner    = coreutil.Func("setNextOwner")
	FuncClaimOwnership  = coreutil.Func("claimOwnership")
	FuncGetOwner        = coreutil.ViewFunc("getOwner")
	FuncSetGasPerIota   = coreutil.Func("setGasPerIota")
	FuncGetGasPerIota   = coreutil.ViewFunc("getGasPerIota")
	FuncWithdrawGasFees = coreutil.Func("withdrawGasFees")
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
	FieldGasPerIota              = "w"
	FieldGasFee                  = "f"
	FieldGasUsed                 = "gu"
	FieldFilterQuery             = "fq"
)

const (
	DefaultChainID = 1074 // IOTA -- get it?

	DefaultGasPerIota uint64 = 1000
	GasLimitDefault          = uint64(15000000)
)

var GasPrice = big.NewInt(0)
