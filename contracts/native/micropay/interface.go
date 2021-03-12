// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package micropay

import (
	"github.com/iotaledger/wasp/contracts/native"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"time"
)

const (
	Name        = "micropay"
	description = "Micro payment PoC smart contract"
)

var (
	Interface = &coreutil.ContractInterface{
		Name:        Name,
		Description: description,
		ProgramHash: hashing.HashStrings(Name),
	}
)

func init() {
	Interface.WithFunctions(initialize, []coreutil.ContractFunctionInterface{
		coreutil.Func(FuncPublicKey, publicKey),
		coreutil.Func(FuncAddWarrant, addWarrant),
		coreutil.Func(FuncRevokeWarrant, revokeWarrant),
		coreutil.Func(FuncCloseWarrant, closeWarrant),
		coreutil.Func(FuncSettle, settle),
		coreutil.ViewFunc(FuncGetChannelInfo, getWarrantInfo),
	})
	native.AddProcessor(Interface)
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
