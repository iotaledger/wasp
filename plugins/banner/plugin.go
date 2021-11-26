package banner

import (
	"fmt"

	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/wasp"
)

// PluginName is the name of the banner plugin.
const PluginName = "Banner"

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
                %s (commit: %s)
`, wasp.Version, wasp.VersionHash)
	fmt.Println()

	// TODO embed build time see https://stackoverflow.com/questions/53031035/generate-build-timestamp-in-go/53045029
	ctx.Node.Logger.Info("Loading plugins ...")
}

func run(_ *node.Plugin) {
}
