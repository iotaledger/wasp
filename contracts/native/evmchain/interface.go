// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package evmchain provides the `evmchain` contract, which allows to emulate an
// Ethereum blockchain on top of ISCP.
package evmchain

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

var Contract = coreutil.NewContract("evmchain", "EVM chain smart contract")

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
	FuncGetBlockTransactionCountByHash      = coreutil.ViewFunc("getBlockTransactionCountByHash")
	FuncGetBlockTransactionCountByNumber    = coreutil.ViewFunc("getBlockTransactionCountByNumber")
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
	FieldEvmOwner                = "o"
	FieldNextEvmOwner            = "n"
	FieldGasPerIota              = "w"
	FieldGasFee                  = "f"
	FieldGasUsed                 = "gu"
	FieldFilterQuery             = "fq"
)

const DefaultGasPerIota uint64 = 1000
