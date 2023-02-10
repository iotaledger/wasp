package waspcmd

import (
	"regexp"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

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
	var setAsDefault bool

	cmd := &cobra.Command{
		Use:   "add <name> <api url>",
		Short: "adds a wasp node",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			nodeName := args[0]
			if !slugRegex.Match([]byte(nodeName)) {
				log.Fatalf("invalid node name: %s, must be in slug format, only lowercase and hypens, example: foo-bar", nodeName)
			}
			config.AddWaspNode(nodeName, args[1])
			if setAsDefault {
				config.SetDefaultWaspNode(nodeName)
			}
		},
	}

	cmd.Flags().BoolVar(&setAsDefault, "default", false, "sets this as the default node")
	return cmd
}

func WithWaspNodesFlag(cmd *cobra.Command, nodes *[]string) {
	cmd.Flags().StringSliceVar(nodes, "nodes", nil, "wasp nodes to execute the command in (ex: bob,alice,foo,bar) (default: the default wasp node)")
}

func DefaultNodesFallback(nodes []string) []string {
	if len(nodes) == 0 {
		return []string{config.MustGetDefaultWaspNode()}
	}
	return nodes
}

func WithWaspNodeFlag(cmd *cobra.Command, node *string) {
	cmd.Flags().StringVar(node, "node", "", "wasp node to execute the command in (ex: wasp-0) (default: the default wasp node)")
}

func DefaultWaspNodeFallback(node string) string {
	if node == "" {
		return config.MustGetDefaultWaspNode()
	}
	return node
}
