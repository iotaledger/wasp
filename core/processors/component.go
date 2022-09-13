package processors

import (
	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:      "Processors",
			Configure: configure,
		},
	}
}

var (
	CoreComponent *app.CoreComponent

	Config          *processors.Config
	nativeContracts = []*coreutil.ContractProcessor{
		inccounter.Processor,
	}
)

func configure() error {
	CoreComponent.LogInfo("Registering native contracts...")
	for _, c := range nativeContracts {
		CoreComponent.LogDebugf(
			"Registering native contract: name: '%s', program hash: %s, description: '%s'\n",
			c.Contract.Name, c.Contract.ProgramHash.String(), c.Contract.Description,
		)
	}
	Config = coreprocessors.Config().WithNativeContracts(nativeContracts...)

	return nil
}
