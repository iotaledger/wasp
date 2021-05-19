// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package evmchainmanagement provides the `evmchainmanagement` contract, which
// is a PoC for a smart contract to manage the ownership of the evmchain contract
package evmchainmanagement

import (
	"github.com/iotaledger/wasp/contracts/native"
	"github.com/iotaledger/wasp/contracts/native/evmchain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

const (
	Name        = "evmchainmanagement"
	description = "EVM chain management"
)

var Interface = &coreutil.ContractInterface{
	Name:        Name,
	Description: description,
	ProgramHash: hashing.HashStrings(Name),
}

func init() {
	Interface.WithFunctions(initialize, []coreutil.ContractFunctionInterface{
		coreutil.Func(FuncRequestGasFees, requestGasFees),
	})
	native.AddProcessor(Interface)
}

const (
	FuncRequestGasFees = "requestGasFee"
)

func initialize(ctx coretypes.Sandbox) (dict.Dict, error) {
	return nil, nil
}

// TODO this SC could adjust gasPerIota based on some conditions

func requestGasFees(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	_, err := ctx.Call(evmchain.Interface.Hname(), coretypes.Hn(evmchain.FuncWithdrawGasFees), nil, nil)
	a.RequireNoError(err)
	return nil, nil
}
