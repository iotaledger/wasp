package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/plugins/config"
	flag "github.com/spf13/pflag"
)

var enabledPlugins []string
var disabledPlugins []string

// AddPluginStatus adds the status (enabled=1, disabled=0) of a given plugin.
func AddPluginStatus(name string, status int) {
	switch status {
	case node.Enabled:
		enabledPlugins = append(enabledPlugins, name)
	case node.Disabled:
		disabledPlugins = append(disabledPlugins, name)
	}
}

func getList(a []string) string {
	sort.Strings(a)
	return strings.Join(a, " ")
}

func printUsage() {
	_, _ = fmt.Fprintf(
		os.Stderr,
		"\n"+
			"Waspr\n\n"+
			"  A IOTA Smartc Contracts node.\n\n"+
			"Usage:\n\n"+
			"  %s [OPTIONS]\n\n"+
			"Options:\n",
		filepath.Base(os.Args[0]),
	)
	flag.PrintDefaults()

	_, _ = fmt.Fprintf(os.Stderr, "\nThe following plugins are enabled by default and can be disabled with -%s:\n  %s\n", config.CfgDisablePlugins, getList(enabledPlugins))
	_, _ = fmt.Fprintf(os.Stderr, "The following plugins are disabled by default and can be enabled with -%s:\n  %s\n", config.CfgEnablePlugins, getList(disabledPlugins))
	_, _ = fmt.Fprintf(os.Stderr, "The enabled/disabled plugins can be overridden by altering %s/%s inside config.json\n\n", config.CfgEnablePlugins, config.CfgDisablePlugins)
}
