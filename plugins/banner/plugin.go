package banner

import (
	"fmt"

	"github.com/iotaledger/hive.go/node"
)

// PluginName is the name of the banner plugin.
const PluginName = "Banner"

const (
	// AppVersion version number
	AppVersion = "v0.0.0"

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

	ctx.Node.Logger.Infof("Wasp version %s ...", AppVersion)
	ctx.Node.Logger.Info("Loading plugins ...")
}

func run(ctx *node.Plugin) {

}
