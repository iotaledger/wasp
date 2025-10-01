package chain

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/config"
	"github.com/spf13/cobra"
)

func withChainFlag(cmd *cobra.Command, chainName *string) {
	cmd.Flags().StringVar(chainName, "chain", "", "target chain name")
}

func defaultChainFallback(chainName string) (string, error) {
	if chainName != "" {
		return chainName, nil
	}
	return getDefaultChain()
}

func getDefaultChain() (string, error) {
	chainSettings := map[string]interface{}{}
	chainsKey := config.Config.Cut("chains")
	if chainsKey != nil {
		chainSettings = chainsKey.All()
	}
	switch len(chainSettings) {
	case 0:
		return "", fmt.Errorf("no chains configured, you can add a new chain with `wasp-cli chain add <name> <chain id>`")
	case 1:
		for nodeName := range chainSettings {
			return nodeName, nil
		}
	default:
		return "", fmt.Errorf("more than 1 chain in the configuration, you can specify the target chain with `--chain=<name>`")
	}
	return "", nil
}
