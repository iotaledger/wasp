// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package micropay

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/contract"
	"github.com/iotaledger/wasp/packages/vm/examples"
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
		contract.Func(FuncOpenChannel, openChannel),
		contract.ViewFunc(FuncGetChannelInfo, getChannelInfo),
	})
	examples.AddProcessor(Interface)
}

const (
	MinimumWarrantIotas = 500

	FuncOpenChannel    = "openChannel"
	FuncGetChannelInfo = "getChannelInfo"

	ParamServiceAddress = "sa"
)
