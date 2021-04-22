// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

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
	})
	native.AddProcessor(Interface)
}

const (
	FuncGetBalance      = "getBalance"
	FuncSendTransaction = "sendTransaction"
	FuncCallView        = "callView"
)

const (
	FieldAddress         = "a"
	FieldTransactionData = "t"
	FieldBalance         = "b"
	FieldCallArguments   = "c"
	FieldResult          = "r"
)
