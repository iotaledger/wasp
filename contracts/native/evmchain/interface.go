// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package evmchain provides the `evmchain` contract, which allows to emulate an
// Ethereum blockchain on top of ISCP.
package evmchain

import (
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
)

var Interface = coreutil.NewContractInterface("evmchain", "EVM chain smart contract")

func init() {
	Interface.WithFunctions(initialize, []coreutil.ContractFunctionInterface{
		// Ethereum blockchain
		coreutil.Func(FuncSendTransaction, applyTransaction),
		coreutil.ViewFunc(FuncGetBalance, getBalance),
		coreutil.ViewFunc(FuncCallContract, callContract),
		coreutil.ViewFunc(FuncEstimateGas, estimateGas),
		coreutil.ViewFunc(FuncGetNonce, getNonce),
		coreutil.ViewFunc(FuncGetReceipt, getReceipt),
		coreutil.ViewFunc(FuncGetCode, getCode),
		coreutil.ViewFunc(FuncGetBlockNumber, getBlockNumber),
		coreutil.ViewFunc(FuncGetBlockByNumber, getBlockByNumber),
		coreutil.ViewFunc(FuncGetBlockByHash, getBlockByHash),
		coreutil.ViewFunc(FuncGetTransactionByHash, getTransactionByHash),
		coreutil.ViewFunc(FuncGetTransactionByBlockHashAndIndex, getTransactionByBlockHashAndIndex),
		coreutil.ViewFunc(FuncGetTransactionByBlockNumberAndIndex, getTransactionByBlockNumberAndIndex),
		coreutil.ViewFunc(FuncGetBlockTransactionCountByHash, getBlockTransactionCountByHash),
		coreutil.ViewFunc(FuncGetBlockTransactionCountByNumber, getBlockTransactionCountByNumber),
		coreutil.ViewFunc(FuncGetStorage, getStorage),
		coreutil.ViewFunc(FuncGetLogs, getLogs),

		// EVMchain SC management
		coreutil.Func(FuncSetNextOwner, setNextOwner),
		coreutil.Func(FuncClaimOwnership, claimOwnership),
		coreutil.Func(FuncSetGasPerIota, setGasPerIota),
		coreutil.Func(FuncWithdrawGasFees, withdrawGasFees),
		coreutil.ViewFunc(FuncGetOwner, getOwner),
		coreutil.ViewFunc(FuncGetGasPerIota, getGasPerIota),
	})
}

const (
	// Ethereum blockchain
	FuncGetBalance                          = "getBalance"
	FuncSendTransaction                     = "sendTransaction"
	FuncCallContract                        = "callContract"
	FuncEstimateGas                         = "estimateGas"
	FuncGetNonce                            = "getNonce"
	FuncGetReceipt                          = "getReceipt"
	FuncGetCode                             = "getCode"
	FuncGetBlockNumber                      = "getBlockNumber"
	FuncGetBlockByNumber                    = "getBlockByNumber"
	FuncGetBlockByHash                      = "getBlockByHash"
	FuncGetTransactionByHash                = "getTransactionByHash"
	FuncGetTransactionByBlockHashAndIndex   = "getTransactionByBlockHashAndIndex"
	FuncGetTransactionByBlockNumberAndIndex = "getTransactionByBlockNumberAndIndex"
	FuncGetBlockTransactionCountByHash      = "getBlockTransactionCountByHash"
	FuncGetBlockTransactionCountByNumber    = "getBlockTransactionCountByNumber"
	FuncGetStorage                          = "getStorage"
	FuncGetLogs                             = "getLogs"

	// EVMchain SC management
	FuncSetNextOwner    = "setNextOwner"
	FuncClaimOwnership  = "claimOwnership"
	FuncGetOwner        = "getOwner"
	FuncSetGasPerIota   = "setGasPerIota"
	FuncGetGasPerIota   = "getGasPerIota"
	FuncWithdrawGasFees = "withdrawGasFees"
)

const (
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
