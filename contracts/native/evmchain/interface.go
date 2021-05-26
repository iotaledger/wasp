// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package evmchain provides the `evmchain` contract, which allows to emulate an
// Ethereum chain on top of ISCP and run EVM contracts.
package evmchain

import (
	"github.com/iotaledger/wasp/contracts/native"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
)

const (
	Name        = "evmchain"
	description = "EVM chain smart contract"
)

var Interface = &coreutil.ContractInterface{
	Name:        Name,
	Description: description,
	ProgramHash: hashing.HashStrings(Name),
}

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
		coreutil.ViewFunc(FuncGetBlockTransactionCountByHash, getBlockTransactionCountByHash),
		coreutil.ViewFunc(FuncGetUncleCountByBlockHash, getUncleCountByBlockHash),
		coreutil.ViewFunc(FuncGetBlockTransactionCountByNumber, getBlockTransactionCountByNumber),
		coreutil.ViewFunc(FuncGetUncleCountByBlockNumber, getUncleCountByBlockNumber),
		coreutil.ViewFunc(FuncGetStorage, getStorage),

		// EVMchain SC management
		coreutil.Func(FuncSetNextOwner, setNextOwner),
		coreutil.Func(FuncClaimOwnership, claimOwnership),
		coreutil.Func(FuncSetGasPerIota, setGasPerIota),
		coreutil.Func(FuncWithdrawGasFees, withdrawGasFees),
		coreutil.ViewFunc(FuncGetOwner, getOwner),
		coreutil.ViewFunc(FuncGetGasPerIota, getGasPerIota),
	})
	native.AddProcessor(Interface)
}

const (
	// Ethereum blockchain
	FuncGetBalance                       = "getBalance"
	FuncSendTransaction                  = "sendTransaction"
	FuncCallContract                     = "callContract"
	FuncEstimateGas                      = "estimateGas"
	FuncGetNonce                         = "getNonce"
	FuncGetReceipt                       = "getReceipt"
	FuncGetCode                          = "getCode"
	FuncGetBlockNumber                   = "getBlockNumber"
	FuncGetBlockByNumber                 = "getBlockByNumber"
	FuncGetBlockByHash                   = "getBlockByHash"
	FuncGetBlockTransactionCountByHash   = "getBlockTransactionCountByHash"
	FuncGetUncleCountByBlockHash         = "getUncleCountByBlockHash"
	FuncGetBlockTransactionCountByNumber = "getBlockTransactionCountByNumber"
	FuncGetUncleCountByBlockNumber       = "getUncleCountByBlockNumber"
	FuncGetStorage                       = "getStorage"

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
	FieldAgentId                 = "i"
	FieldTransactionHash         = "h"
	FieldTransactionData         = "t"
	FieldTransactionDataBlobHash = "th"
	FieldBalance                 = "b"
	FieldCallArguments           = "c"
	FieldResult                  = "r"
	FieldBlockNumber             = "bn"
	FieldBlockHash               = "bh"
	FieldCallMsg                 = "c"
	FieldEvmOwner                = "o"
	FieldNextEvmOwner            = "n"
	FieldGasPerIota              = "w"
	FieldGasFee                  = "f"
)

const DefaultGasPerIota uint64 = 1000
