package banner

import (
	"fmt"

	"github.com/iotaledger/hive.go/node"
)

// PluginName is the name of the banner plugin.
const PluginName = "Banner"

// Plugin is the plugin instance of the banner plugin.
var Plugin = node.NewPlugin(PluginName, node.Enabled, configure, run)

const (
	// AppVersion version number
	AppVersion = "v0.0.0"

	// AppName app code name
	AppName = "Wasp"
)

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

	ctx.Node.Logger.Infof("Wasp version %s ...", AppVersion)
	ctx.Node.Logger.Info("Loading plugins ...")
}

func run(ctx *node.Plugin) {

}
