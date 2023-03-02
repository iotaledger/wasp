package processors

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:    "Processors",
			Provide: provide,
		},
	}
}

var CoreComponent *app.CoreComponent

func provide(c *dig.Container) error {
	type processorsConfigResult struct {
		dig.Out

		ProcessorsConfig *processors.Config
	}

	if err := c.Provide(func() processorsConfigResult {
		CoreComponent.LogInfo("Registering native contracts...")

		nativeContracts := []*coreutil.ContractProcessor{
			inccounter.Processor,
		}

		for _, c := range nativeContracts {
			CoreComponent.LogDebugf(
				"Registering native contract: name: '%s', program hash: %s, description: '%s'\n",
				c.Contract.Name, c.Contract.ProgramHash.String(), c.Contract.Description,
			)
		}

		return processorsConfigResult{
			ProcessorsConfig: coreprocessors.NewConfigWithCoreContracts().WithNativeContracts(nativeContracts...),
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}
