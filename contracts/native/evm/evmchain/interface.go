// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package evmchain provides the `evmchain` contract, which allows to emulate an
// Ethereum blockchain on top of ISCP.
package evmchain

import "github.com/iotaledger/wasp/packages/iscp/coreutil"

var Contract = coreutil.NewContract("evmchain", "EVM chain smart contract")
