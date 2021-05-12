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
		coreutil.Func(FuncSendTransaction, applyTransaction),
		coreutil.ViewFunc(FuncGetBalance, getBalance),
		coreutil.ViewFunc(FuncCallView, callView),
		coreutil.ViewFunc(FuncGetNonce, getNonce),
		coreutil.ViewFunc(FuncGetReceipt, getReceipt),
		coreutil.ViewFunc(FuncGetCode, getCode),
		coreutil.ViewFunc(FuncGetBlockNumber, getBlockNumber),
		coreutil.ViewFunc(FuncGetBlockByNumber, getBlockByNumber),
	})
	native.AddProcessor(Interface)
}

const (
	FuncGetBalance       = "getBalance"
	FuncSendTransaction  = "sendTransaction"
	FuncCallView         = "callView"
	FuncGetNonce         = "getNonce"
	FuncGetReceipt       = "getReceipt"
	FuncGetCode          = "getCode"
	FuncGetBlockNumber   = "getBlockNumber"
	FuncGetBlockByNumber = "getBlockByNumber"
)

const (
	FieldGenesisAlloc    = "g"
	FieldAddress         = "a"
	FieldTransactionHash = "h"
	FieldTransactionData = "t"
	FieldBalance         = "b"
	FieldCallArguments   = "c"
	FieldResult          = "r"
	FieldBlockNumber     = "bn"
)
