package processors

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

const pluginName = "Processors"

var (
	log    *logger.Logger
	Config *processors.Config

	nativeContracts = []*coreutil.ContractProcessor{
		inccounter.Processor,
	}
)

func Init() *node.Plugin {
	return node.NewPlugin(pluginName, nil, node.Enabled, configure, run)
}

func configure(ctx *node.Plugin) {
	log = logger.NewLogger(pluginName)

	log.Info("Registering native contracts...")
	for _, c := range nativeContracts {
		log.Debugf(
			"Registering native contract: name: '%s', program hash: %s, description: '%s'\n",
			c.Contract.Name, c.Contract.ProgramHash.String(), c.Contract.Description,
		)
	}
	Config = coreprocessors.Config().WithNativeContracts(nativeContracts...)
}

func run(_ *node.Plugin) {
}
