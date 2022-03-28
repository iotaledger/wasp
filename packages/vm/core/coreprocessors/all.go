package coreprocessors

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/governance/governanceimpl"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/root/rootimpl"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

var All = map[hashing.HashValue]iscp.VMProcessor{
	root.Contract.ProgramHash:       rootimpl.Processor,
	errors.Contract.ProgramHash:     errors.Processor,
	accounts.Contract.ProgramHash:   accounts.Processor,
	blob.Contract.ProgramHash:       blob.Processor,
	blocklog.Contract.ProgramHash:   blocklog.Processor,
	governance.Contract.ProgramHash: governanceimpl.Processor,
	evm.Contract.ProgramHash:        evmimpl.Processor,
}

func init() {
	if len(corecontracts.All) != len(All) {
		panic("static check: mismatch between corecontracts.All and coreprocessors.All")
	}
}

func Config() *processors.Config {
	return processors.NewConfig().WithCoreContracts(All)
}
