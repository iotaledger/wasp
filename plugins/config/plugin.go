package config

import (
	"fmt"
	"os"

	"github.com/iotaledger/hive.go/configuration"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/node"
	flag "github.com/spf13/pflag"
)

// PluginName is the name of the config plugin.
const PluginName = "Config"

const (
	// CfgDisablePlugins contains the name of the parameter that allows to manually disable node plugins.
	CfgDisablePlugins = "node.disablePlugins"

	// CfgEnablePlugins contains the name of the parameter that allows to manually enable node plugins.
	CfgEnablePlugins = "node.enablePlugins"
)

var (
	Plugin *node.Plugin

	// Node is viper
	Node *configuration.Configuration
)

func Init() *node.Plugin {
	Plugin = node.NewPlugin(PluginName, node.Enabled)
	// set the default logger config
	Node = configuration.New()

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
	configFilePath := flag.StringP("config", "c", "config.json", "File path of the config file")
	skipConfigAvailable := flag.Bool("skip-config", false, "Skip config file availability check")

	flag.Parse()

	err := Node.LoadFile(*configFilePath)
	if err != nil {
		return *skipConfigAvailable, err
	}

	if err := Node.LoadFlagSet(flag.CommandLine); err != nil {
		return *skipConfigAvailable, err
	}

	// read in ENV variables
	// load the env vars after default values from flags were set (otherwise the env vars are not added because the keys don't exist)
	if err := Node.LoadEnvironmentVars(""); err != nil {
		return *skipConfigAvailable, err
	}

	// load the flags again to overwrite env vars that were also set via command line
	if err := Node.LoadFlagSet(flag.CommandLine); err != nil {
		return *skipConfigAvailable, err
	}

	if printConfig {
		Node.Print(ignoreSettingsAtPrint...)
	}

	for _, pluginName := range Node.Strings(CfgDisablePlugins) {
		node.DisabledPlugins[node.GetPluginIdentifier(pluginName)] = true
	}
	for _, pluginName := range Node.Strings(CfgEnablePlugins) {
		node.EnabledPlugins[node.GetPluginIdentifier(pluginName)] = true
	}

	return *skipConfigAvailable, nil
}
