// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package micropay

import (
	"time"

	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

var Contract = coreutil.NewContract("micropay", "Micro payment PoC smart contract")

const MinimumWarrantBaseTokens = 500

var (
	FuncPublicKey      = coreutil.Func("publicKey")
	FuncAddWarrant     = coreutil.Func("addWarrant")
	FuncRevokeWarrant  = coreutil.Func("revokeWarrant")
	FuncCloseWarrant   = coreutil.Func("closeWarrant")
	FuncSettle         = coreutil.Func("settle")
	FuncGetChannelInfo = coreutil.ViewFunc("getWarrantInfo")
)

const (
	ParamPublicKey      = "pk"
	ParamPayerAddress   = "pa"
	ParamServiceAddress = "sa"
	ParamPayments       = "m"

	ParamWarrant = "wa"
	ParamRevoked = "re"
	ParamLastOrd = "lo"

	StateVarPublicKeys = "k"

	WarrantRevokePeriod = 1 * time.Hour
)
