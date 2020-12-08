// ony needed to link packages with examples
package examples

import (
	"github.com/iotaledger/hive.go/node"
	_ "github.com/iotaledger/wasp/packages/vm/examples/inccounter"
)

// PluginName is the name of the banner plugin.
const PluginName = "Examples"

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

func configure(ctx *node.Plugin) {
}

func run(ctx *node.Plugin) {
}
