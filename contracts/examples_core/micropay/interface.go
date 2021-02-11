// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package micropay

import (
	"github.com/iotaledger/wasp/contracts"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/contract"
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
		contract.Func(FuncSettle, settle),
		contract.ViewFunc(FuncGetChannelInfo, getWarrantInfo),
	})
	contracts.AddProcessor(Interface)
}

const (
	MinimumWarrantIotas = 500

	FuncPublicKey      = "publicKey"
	FuncAddWarrant     = "addWarrant"
	FuncRevokeWarrant  = "revokeWarrant"
	FuncCloseWarrant   = "closeWarrant"
	FuncSettle         = "settle"
	FuncGetChannelInfo = "getWarrantInfo"

	ParamPublicKey      = "pk"
	ParamPayerAddress   = "pa"
	ParamServiceAddress = "sa"
	ParamPayments       = "m"

	ParamWarrant = "wa"
	ParamRevoked = "re"
	ParamLastOrd = "lo"

	StateVarPublicKeys = "k"
	StateVarLastOrdNum = "o"

	WarrantRevokePeriod = 1 * time.Hour
)
