// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmflavors

import (
	"github.com/iotaledger/wasp/contracts/native/evm/evmchain"
	"github.com/iotaledger/wasp/contracts/native/evm/evmlight"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

var All = map[string]*coreutil.ContractInfo{
	evmchain.Contract.Name: evmchain.Contract,
	evmlight.Contract.Name: evmlight.Contract,
}

var Processors = map[string]*coreutil.ContractProcessor{
	evmlight.Contract.Name: evmlight.Processor,
	evmchain.Contract.Name: evmchain.Processor,
}
