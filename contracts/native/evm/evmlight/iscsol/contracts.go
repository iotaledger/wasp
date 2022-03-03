// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscsol

import (
	_ "embed"
)

// If you change ISC.sol, you must recompile.  You will need
// the `solc` binary installed in your system. Then, simply run `go generate`
// in this directory.

//go:generate sh -c "solc --abi --overwrite ISC.sol -o ."
var (
	//go:embed ISC.abi
	ISCContractABI string
)
