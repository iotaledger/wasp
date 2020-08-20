package config

import (
	"fmt"
	"os"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/hive.go/parameter"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// PluginName is the name of the config plugin.
const PluginName = "Config"

var (
	Plugin *node.Plugin

	// Node is viper
	Node *viper.Viper
)

func Init() *node.Plugin {
	Plugin = node.NewPlugin(PluginName, node.Enabled)
	// set the default logger config
	Node = viper.New()

	Plugin.Events.Init.Attach(events.NewClosure(func(*node.Plugin) {
		if skipConfigAvailable, err := fetch(false); err != nil {
			if !skipConfigAvailable {
				// we wanted a config file but it was not present
				// global logger instance is not initialized at this stage...
				fmt.Println(err.Error())
				fmt.Println("no config file present, terminating Wasp. please use the provided config.default.json to create a config.json.")
				// daemon is not running yet, so we just exit
				os.Exit(1)
			}
			panic(err)
		}
	}))

	return Plugin
}

// fetch fetches config values from a dir defined via CLI flag --config-dir (or the current working dir if not set).
//
// It automatically reads in a single config file starting with "config" (can be changed via the --config CLI flag)
// and ending with: .json, .toml, .yaml or .yml (in this sequence).
func fetch(printConfig bool, ignoreSettingsAtPrint ...[]string) (bool, error) {
	// flags
	configName := flag.StringP("config", "c", "config", "Filename of the config file without the file extension")
	configDirPath := flag.StringP("config-dir", "d", ".", "Path to the directory containing the config file")
	skipConfigAvailable := flag.Bool("skip-config", false, "Skip config file availability check")

	flag.Parse()

	err := parameter.LoadConfigFile(Node, *configDirPath, *configName, true, *skipConfigAvailable)
	if err != nil {
		return *skipConfigAvailable, err
	}

	if printConfig {
		parameter.PrintConfig(Node, ignoreSettingsAtPrint...)
	}

	for _, pluginName := range Node.GetStringSlice(node.CFG_DISABLE_PLUGINS) {
		node.DisabledPlugins[node.GetPluginIdentifier(pluginName)] = true
	}
	for _, pluginName := range Node.GetStringSlice(node.CFG_ENABLE_PLUGINS) {
		node.EnabledPlugins[node.GetPluginIdentifier(pluginName)] = true
	}

	return *skipConfigAvailable, nil
}
