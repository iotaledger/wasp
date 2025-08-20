// Package coreprocessors provides a registry of core VM processors for IOTA Smart Contracts.
// It manages mappings between contract hnames and their processor implementations for both
// production core contracts and test-only contracts. The package allows creating processor
// configurations for different environments and ensures integrity between core contract
// definitions and their processor implementations.
package coreprocessors

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/v2/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/errors"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance/governanceimpl"
	"github.com/iotaledger/wasp/v2/packages/vm/core/root"
	"github.com/iotaledger/wasp/v2/packages/vm/core/root/rootimpl"
	"github.com/iotaledger/wasp/v2/packages/vm/core/testcore/contracts/inccounter"
	"github.com/iotaledger/wasp/v2/packages/vm/core/testcore/contracts/manyevents"
	"github.com/iotaledger/wasp/v2/packages/vm/core/testcore/contracts/testerrors"
	"github.com/iotaledger/wasp/v2/packages/vm/core/testcore/sbtests/sbtestsc"
	"github.com/iotaledger/wasp/v2/packages/vm/processors"
)

var All = map[isc.Hname]isc.VMProcessor{
	root.Contract.Hname():       rootimpl.Processor,
	errors.Contract.Hname():     errors.Processor,
	accounts.Contract.Hname():   accounts.Processor,
	blocklog.Contract.Hname():   blocklog.Processor,
	governance.Contract.Hname(): governanceimpl.Processor,
	evm.Contract.Hname():        evmimpl.Processor,
}

var Test = map[isc.Hname]isc.VMProcessor{
	inccounter.Contract.Hname(): inccounter.Processor,
	manyevents.Contract.Hname(): manyevents.Processor,
	testerrors.Contract.Hname(): testerrors.Processor,
	sbtestsc.Contract.Hname():   sbtestsc.Processor,
}

func init() {
	if len(corecontracts.All) != len(All) {
		panic("static check: mismatch between corecontracts.All and coreprocessors.All")
	}
}

func NewConfig() *processors.Config {
	return processors.NewConfig().WithCoreContracts(All)
}

func NewConfigWithTestContracts() *processors.Config {
	combined := make(map[isc.Hname]isc.VMProcessor, len(All)+len(Test))
	for k, v := range All {
		combined[k] = v
	}
	for k, v := range Test {
		combined[k] = v
	}

	return processors.NewConfig().WithCoreContracts(combined)
}
