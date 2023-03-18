package chain

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func withChainFlag(cmd *cobra.Command, chainName *string) {
	cmd.Flags().StringVar(chainName, "chain", "", "target chain name")
}

func defaultChainFallback(chainName string) string {
	if chainName != "" {
		return chainName
	}
	return getDefaultChain()
}

func getDefaultChain() string {
	chainSettings := map[string]interface{}{}
	chainsKey := viper.Sub("chains")
	if chainsKey != nil {
		chainSettings = chainsKey.AllSettings()
	}
	switch len(chainSettings) {
	case 0:
		log.Fatalf("no chains configured, you can add a new chain with `wasp-cli add <name> <chain id>`")
	case 1:
		for nodeName := range chainSettings {
			return nodeName
		}
	default:
		log.Fatalf("more than 1 chain in the configuration, you can specify the target chain with `--chain=<name>`")
	}
	panic("unreachable")
}
