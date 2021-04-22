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
		coreutil.ViewFunc(FuncGetBalance, getBalance),
	})
	native.AddProcessor(Interface)
}

const FuncGetBalance = "getBalance"

const (
	FieldAddress = "a"
	FieldBalance = "b"
)
