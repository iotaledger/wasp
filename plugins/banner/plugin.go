package banner

import (
	"fmt"

	"github.com/iotaledger/hive.go/node"
)

// PluginName is the name of the banner plugin.
const PluginName = "Banner"

const (
	// Version version number
	Version = "v0.2.0"

	// Name app code name
	Name = "Wasp"
)

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

func configure(ctx *node.Plugin) {
	fmt.Printf(`
     __          __
     \ \        / /
      \ \  /\  / /_ _ ___ _ __
       \ \/  \/ / _| / __| |_ \
        \  /\  / (_| \__ \ |_) |
         \/  \/ \__,_|___/ |__/
                         | |
                         |_|
                %s
`, Version)
	fmt.Println()

	// TODO embed build time see https://stackoverflow.com/questions/53031035/generate-build-timestamp-in-go/53045029
	ctx.Node.Logger.Info("Loading plugins ...")
}

func run(_ *node.Plugin) {
}
