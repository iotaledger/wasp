package cli

import (
	"fmt"
	"os"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/wasp"
	flag "github.com/spf13/pflag"
	"go.uber.org/dig"
)

// PluginName is the name of the CLI plugin.
const PluginName = "CLI"

var printVersion bool

func Init() *node.Plugin {
	flag.BoolVarP(&printVersion, "version", "v", false, "Prints the Wasp version")

	Plugin := node.NewPlugin(PluginName, nil, node.Enabled)
	Plugin.Events.Init.Attach(events.NewClosure(onInit))
	return Plugin
}

func onAddPlugin(name string, status int) {
	AddPluginStatus(node.GetPluginIdentifier(name), status)
}

func onInit(*node.Plugin, *dig.Container) {
	for name, plugin := range node.GetPlugins() {
		onAddPlugin(name, plugin.Status)
	}
	node.Events.AddPlugin.Attach(events.NewClosure(onAddPlugin))

	flag.Usage = printUsage

	if printVersion {
		fmt.Println(wasp.Name + " " + wasp.Version)
		os.Exit(0)
	}
}
