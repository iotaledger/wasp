package coreprocessors

import (
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/governance/governanceimpl"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/root/rootimpl"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/contracts/inccounter"
	"github.com/iotaledger/wasp/packages/vm/processors"
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
	return processors.NewConfig().WithCoreContracts(lo.MergeMaps(All, Test))
}
