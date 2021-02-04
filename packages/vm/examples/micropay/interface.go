// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package micropay

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/contract"
	"github.com/iotaledger/wasp/packages/vm/examples"
	"time"
)

const (
	Name        = "micropay"
	description = "Micro payment PoC smart contract"
)

var (
	Interface = &contract.ContractInterface{
		Name:        Name,
		Description: description,
		ProgramHash: hashing.HashStrings(Name),
	}
)

func init() {
	Interface.WithFunctions(initialize, []contract.ContractFunctionInterface{
		contract.Func(FuncPublicKey, publicKey),
		contract.Func(FuncAddWarrant, addWarrant),
		contract.Func(FuncRevokeWarrant, revokeWarrant),
		contract.Func(FuncCloseWarrant, closeWarrant),
		contract.ViewFunc(FuncGetChannelInfo, getWarrantInfo),
	})
	examples.AddProcessor(Interface)
}

const (
	MinimumWarrantIotas = 500

	FuncPublicKey      = "publicKey"
	FuncAddWarrant     = "addWarrant"
	FuncRevokeWarrant  = "revokeWarrant"
	FuncCloseWarrant   = "closeWarrant"
	FuncGetChannelInfo = "getWarrantInfo"

	ParamPublicKey      = "pk"
	ParamPayerAddress   = "pa"
	ParamServiceAddress = "sa"

	ParamWarrant = "wa"
	ParamRevoked = "re"

	StateVarPublicKeys = "k"

	WarrantRevokePeriod = 1 * time.Hour
)
