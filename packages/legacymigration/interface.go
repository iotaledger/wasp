// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
package legacymigration

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

const contractName = "legacymigration"

var Contract = coreutil.NewContract(contractName)

// Views
var (
	ViewMigratableBalance = coreutil.ViewFunc("getMigratableBalance")
	ViewTotalBalance      = coreutil.ViewFunc("getTotalBalance")
)

// Funcs
var FuncMigrate = coreutil.Func("migrate")

const (
	ParamAddress = "a"
	ParamBundle  = "b"
)
