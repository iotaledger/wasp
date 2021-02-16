package banner

import (
	"fmt"

	"github.com/iotaledger/hive.go/node"
)

// PluginName is the name of the banner plugin.
const PluginName = "Banner"

const (
	// AppVersion version number
	AppVersion = "v0.1.0"

	// AppName app code name
	AppName = "Wasp"
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
`, AppVersion)
	fmt.Println()

	// TODO embed build time see https://stackoverflow.com/questions/53031035/generate-build-timestamp-in-go/53045029
	ctx.Node.Logger.Info("Loading plugins ...")
}

func run(_ *node.Plugin) {
}
