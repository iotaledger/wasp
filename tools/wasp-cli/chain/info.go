// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

func initInfoCmd() *cobra.Command {
	var node string
	var chain string
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show information about the chain",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)

			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			chainInfo, res, err := client.ChainsAPI.
				GetChainInfo(ctx).
				Execute() //nolint:bodyclose // false positive

			if res.StatusCode == http.StatusNotFound {
				fmt.Print("No chain info available. Is the chain deployed and activated?\n")
				return
			}

			log.Check(err)

			committeeInfo, _, err := client.ChainsAPI.
				GetCommitteeInfo(ctx).
				Execute() //nolint:bodyclose // false positive
			log.Check(err)

			printNodesRowHdr := []string{"PubKey", "PeeringURL", "Alive", "Committee", "Access", "AccessAPI"}
			printNodesRowFmt := func(n apiclient.CommitteeNode, isCommitteeNode, isAccessNode bool) []string {
				return []string{
					n.Node.PublicKey,
					n.Node.PeeringURL,
					strconv.FormatBool(n.Node.IsAlive),
					strconv.FormatBool(isCommitteeNode),
					strconv.FormatBool(isAccessNode),
					n.AccessAPI,
				}
			}
			printNodes := func(label string, nodes []apiclient.CommitteeNode, isCommitteeNode, isAccessNode bool) {
				if nodes == nil {
					log.Printf("%s: N/A\n", label)
				}
				log.Printf("%s: %v\n", label, len(nodes))
				rows := make([][]string, 0)
				for _, n := range nodes {
					rows = append(rows, printNodesRowFmt(n, isCommitteeNode, isAccessNode))
				}
				log.PrintTable(printNodesRowHdr, rows)
			}

			log.Printf("Chain ID: %s\n", chainInfo.ChainID)
			log.Printf("EVM Chain ID: %d\n", chainInfo.EvmChainId)
			log.Printf("Active: %v\n", chainInfo.IsActive)

			if chainInfo.IsActive {
				log.Printf("State address: %v\n", committeeInfo.StateAddress)

				log.Printf("\n")
				printNodes("Committee nodes", committeeInfo.CommitteeNodes, true, false)
				log.Printf("\n")
				printNodes("Access nodes", committeeInfo.AccessNodes, false, true)
				log.Printf("\n")
				printNodes("Candidate nodes", committeeInfo.CandidateNodes, false, false)
				log.Printf("\n")

				contracts, _, err := client.ChainsAPI.GetContracts(ctx).Execute() //nolint:bodyclose // false positive
				log.Check(err)
				log.Printf("#Contracts: %d\n", len(contracts))

				log.Printf("Admin: %s\n", chainInfo.ChainAdmin)
				log.Printf("Common account: %s\n", accounts.CommonAccount())
				log.Printf("Gas fee: gas units * (%d/%d)\n", chainInfo.GasFeePolicy.GasPerToken.A, chainInfo.GasFeePolicy.GasPerToken.B)
				log.Printf("Validator fee share: %d%%\n", chainInfo.GasFeePolicy.ValidatorFeeShare)

				log.Printf("\n[Gas Limits]\n")
				log.Printf("MinGasPerRequest: %d\n", chainInfo.GasLimits.MinGasPerRequest)
				log.Printf("MaxGasPerRequest: %d\n", chainInfo.GasLimits.MaxGasPerRequest)
				log.Printf("MaxGasPerBlock: %d\n", chainInfo.GasLimits.MaxGasPerBlock)
				log.Printf("MaxGasExternalViewCall: %d\n", chainInfo.GasLimits.MaxGasExternalViewCall)
			}

			log.Printf("\n[Metadata]\n")
			log.Printf("Name: %s\n", chainInfo.Metadata.Name)
			log.Printf("Description: %s\n", chainInfo.Metadata.Description)
			log.Printf("Website: %s\n", chainInfo.Metadata.Website)

			log.Printf("Public API: %s\n", chainInfo.PublicURL)
			log.Printf("EVM Json RPC URL: %s\n", chainInfo.Metadata.EvmJsonRpcURL)
			log.Printf("EVM WebSocket URL: %s\n", chainInfo.Metadata.EvmWebSocketURL)
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	return cmd
}
