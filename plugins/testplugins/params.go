package testplugins

import "github.com/iotaledger/hive.go/node"

const testPluginsEnabled = true

var enabledTests = map[string]bool{
	"TestingSCMetaData": true,
	"TestingRoundTrip":  true,
}

func Status(pluginName string) int {
	if !testPluginsEnabled {
		return node.Disabled
	}
	enabled, ok := enabledTests[pluginName]
	if !ok {
		return node.Disabled
	}
	if enabled {
		return node.Enabled
	}
	return node.Disabled
}
