// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package evmlight provides the `evmlight` contract, which allows to run EVM code
package evmlight

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

var Contract = coreutil.NewContract("evmlight", "evmlight smart contract")

var (
	// Ethereum blockchain
	FuncGetBalance                       = coreutil.ViewFunc("getBalance")
	FuncSendTransaction                  = coreutil.Func("sendTransaction")
	FuncCallContract                     = coreutil.ViewFunc("callContract")
	FuncEstimateGas                      = coreutil.ViewFunc("estimateGas")
	FuncGetNonce                         = coreutil.ViewFunc("getNonce")
	FuncGetReceipt                       = coreutil.ViewFunc("getReceipt")
	FuncGetCode                          = coreutil.ViewFunc("getCode")
	FuncGetBlockNumber                   = coreutil.ViewFunc("getBlockNumber")
	FuncGetBlockByNumber                 = coreutil.ViewFunc("getBlockByNumber")
	FuncGetBlockByHash                   = coreutil.ViewFunc("getBlockByHash")
	FuncGetTransactionByHash             = coreutil.ViewFunc("getTransactionByHash")
	FuncGetTransactionByBlockHash        = coreutil.ViewFunc("getTransactionByBlockHash")
	FuncGetTransactionByBlockNumber      = coreutil.ViewFunc("getTransactionByBlockNumber")
	FuncGetTransactionCountByBlockHash   = coreutil.ViewFunc("getTransactionCountByHash")
	FuncGetTransactionCountByBlockNumber = coreutil.ViewFunc("getTransactionCountByBlockNumber")
	FuncGetStorage                       = coreutil.ViewFunc("getStorage")
	FuncGetLogs                          = coreutil.ViewFunc("getLogs")

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
	FieldTransactionHash         = "h"
	FieldTransactionData         = "t"
	FieldTransactionDataBlobHash = "th"
	FieldCallArguments           = "c"
	FieldResult                  = "r"
	FieldBlockNumber             = "bn"
	FieldBlockHash               = "bh"
	FieldCallMsg                 = "c"
	FieldNextEvmOwner            = "n"
	FieldGasPerIota              = "w"
	FieldGasFee                  = "f"
	FieldGasUsed                 = "gu"
	FieldFilterQuery             = "fq"
)

const DefaultGasPerIota uint64 = 1000
