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
		coreutil.ViewFunc(FuncCallView, callView),
		coreutil.ViewFunc(FuncGetNonce, getNonce),
		coreutil.ViewFunc(FuncGetReceipt, getReceipt),
		coreutil.ViewFunc(FuncGetCode, getCode),

		// EVMchain SC management
		coreutil.Func(FuncSetOwner, setOwner),
		coreutil.Func(FuncSetGasPerIota, setGasPerIota),
		coreutil.Func(FuncWithdrawGasFees, withdrawGasFees),
		coreutil.ViewFunc(FuncGetOwner, getOwner),
		coreutil.ViewFunc(FuncGetGasPerIota, getGasPerIota),
	})
	native.AddProcessor(Interface)
}

const (
	// Ethereum blockchain
	FuncGetBalance      = "getBalance"
	FuncSendTransaction = "sendTransaction"
	FuncCallView        = "callView"
	FuncGetNonce        = "getNonce"
	FuncGetReceipt      = "getReceipt"
	FuncGetCode         = "getCode"

	// EVMchain SC management
	FuncSetOwner        = "setOwner"
	FuncGetOwner        = "getOwner"
	FuncSetGasPerIota   = "setGasPerIota"
	FuncGetGasPerIota   = "getGasPerIota"
	FuncWithdrawGasFees = "withdrawGasFees"
)

const (
	FieldGenesisAlloc            = "g"
	FieldAddress                 = "a"
	FieldAgentId                 = "i"
	FieldTransactionHash         = "h"
	FieldTransactionData         = "t"
	FieldTransactionDataBlobHash = "th"
	FieldBalance                 = "b"
	FieldCallArguments           = "c"
	FieldResult                  = "r"
	FieldEvmOwner                = "evmOwner"
	FieldGasPerIota              = "gasPerIota"
	FieldGasFee                  = "gasFee"
	FieldGasFeesCollected        = "gasFeeCollected"
)

const DefaultGasPerIota int64 = 1000
