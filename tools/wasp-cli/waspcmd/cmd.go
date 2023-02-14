package waspcmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initWaspNodesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "wasp <command>",
		Short: "Interact with a chain",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			log.Check(cmd.Help())
		},
	}
}

func Init(rootCmd *cobra.Command) {
	waspNodesCmd := initWaspNodesCmd()
	rootCmd.AddCommand(waspNodesCmd)

	waspNodesCmd.AddCommand(initAddWaspNodeCmd())
}

func initAddWaspNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name> <api url>",
		Short: "adds a wasp node",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			nodeName := args[0]
			if !util.IsSlug(nodeName) {
				log.Fatalf("invalid node name: %s, must be in slug format, only lowercase and hypens, example: foo-bar", nodeName)
			}
			config.AddWaspNode(nodeName, args[1])
		},
	}

	return cmd
}

func WithPeersFlag(cmd *cobra.Command, peers *[]string) {
	cmd.Flags().StringSliceVar(peers, "peers", nil, "peers to be included the command in (ex: bob,alice,foo,bar) (default: no peers)")
}

func WithWaspNodeFlag(cmd *cobra.Command, node *string) {
	cmd.Flags().StringVar(node, "node", "", "wasp node to execute the command in (ex: wasp-0) (default: the default wasp node)")
}

func DefaultWaspNodeFallback(node string) string {
	if node != "" {
		return node
	}
	return getDefaultWaspNode()
}

func getDefaultWaspNode() string {
	waspSettings := viper.Sub("wasp").AllSettings()
	switch len(waspSettings) {
	case 0:
		log.Fatalf("no wasp node configured, you can add a node with `wasp-cli wasp add <name> <api url>`")
	case 1:
		for nodeName := range waspSettings {
			return nodeName
		}
	default:
		log.Fatalf("more than 1 wasp node in the configuration, you can specify the target node with `--node=<name>`")
	}
	panic("unreachable")
}
