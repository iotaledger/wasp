// Package waspcmd provides core command functionality for interacting with
// Wasp nodes through the command-line interface.
package waspcmd

import (
	"errors"

	"github.com/iotaledger/wasp/v2/clients/apiextensions"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

func initWaspNodesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "wasp <command>",
		Short: "Configure wasp nodes",
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
	waspNodesCmd.AddCommand(initCheckVersionsCmd())
}

func initAddWaspNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name> <api url>",
		Short: "adds a wasp node",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			nodeName := args[0]
			nodeURL := args[1]

			if !util.IsSlug(nodeName) {
				log.Fatalf("invalid node name: %s, must be in slug format, only lowercase and hyphens, example: foo-bar", nodeName)
			}

			_, err := apiextensions.ValidateAbsoluteURL(nodeURL)
			log.Check(err)

			config.AddWaspNode(nodeName, nodeURL)
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

// defaultWaspNodeFallbackCore contains the shared logic for both fallback functions
func defaultWaspNodeFallbackCore(node string) (string, bool) {
	if node != "" {
		return node, true
	}
	return "", false
}

// DefaultWaspNodeFallback returns the node name or falls back to the default node
// This function calls log.Fatal and is deprecated in favor of DefaultWaspNodeFallbackE
func DefaultWaspNodeFallback(node string) string {
	if result, hasNode := defaultWaspNodeFallbackCore(node); hasNode {
		return result
	}
	return getDefaultWaspNode()
}

// DefaultWaspNodeFallbackE returns the node name or falls back to the default node
// Returns an error instead of calling log.Fatal for proper error handling
func DefaultWaspNodeFallbackE(node string) (string, error) {
	if result, hasNode := defaultWaspNodeFallbackCore(node); hasNode {
		return result, nil
	}
	return getDefaultWaspNodeE()
}

// getDefaultWaspNodeCore contains the shared logic for getting the default wasp node
func getDefaultWaspNodeCore() (string, int, error) {
	waspSettings := map[string]interface{}{}
	waspKey := config.Config.Cut("wasp")
	if waspKey != nil {
		waspSettings = waspKey.All()
	}

	nodeCount := len(waspSettings)
	switch nodeCount {
	case 0:
		return "", nodeCount, errors.New("no wasp node configured, you can add a node with `wasp-cli wasp add <name> <api url>`")
	case 1:
		for nodeName := range waspSettings {
			return nodeName, nodeCount, nil
		}
	default:
		return "", nodeCount, errors.New("more than 1 wasp node in the configuration, you can specify the target node with `--node=<name>`")
	}
	return "", nodeCount, errors.New("unreachable code")
}

// getDefaultWaspNode calls log.Fatal and is deprecated in favor of getDefaultWaspNodeE
func getDefaultWaspNode() string {
	nodeName, _, err := getDefaultWaspNodeCore()
	if err != nil {
		log.Fatalf(err.Error())
	}
	return nodeName
}

// getDefaultWaspNodeE returns the default wasp node or an error
func getDefaultWaspNodeE() (string, error) {
	nodeName, _, err := getDefaultWaspNodeCore()
	return nodeName, err
}
