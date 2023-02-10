// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
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

			chainID := config.GetChain(chain)
			client := cliclients.WaspClient(node)

			chainInfo, _, err := client.ChainsApi.
				GetChainInfo(context.Background(), chainID.String()).
				Execute()
			log.Check(err)

			committeeInfo, _, err := client.ChainsApi.
				GetCommitteeInfo(context.Background(), chainID.String()).
				Execute()
			log.Check(err)

			printNodesRowHdr := []string{"PubKey", "NetID", "Alive", "Committee", "Access", "AccessAPI"}
			printNodesRowFmt := func(n apiclient.CommitteeNode, isCommitteeNode, isAccessNode bool) []string {
				return []string{
					n.Node.PublicKey,
					n.Node.NetId,
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
			log.Printf("Active: %v\n", chainInfo.IsActive)

			if chainInfo.IsActive {
				log.Printf("State address: %v\n", committeeInfo.StateAddress)
				printNodes("Committee nodes", committeeInfo.CommitteeNodes, true, false)
				printNodes("Access nodes", committeeInfo.AccessNodes, false, true)
				printNodes("Candidate nodes", committeeInfo.CandidateNodes, false, false)

				log.Printf("Description: %s\n", chainInfo.Description)

				contracts, _, err := client.ChainsApi.GetContracts(context.Background(), chainID.String()).Execute()
				log.Check(err)
				log.Printf("#Contracts: %d\n", len(contracts))

				log.Printf("Owner: %s\n", chainInfo.ChainOwnerId)

				// TODO: Validate the gas fee token id logic
				if chainInfo.GasFeePolicy != nil {
					gasFeeToken := util.BaseTokenStr

					if chainInfo.GasFeePolicy.GasFeeTokenId != "" {
						decodedToken, err := iotago.DecodeHex(chainInfo.GasFeePolicy.GasFeeTokenId)
						log.Check(err)

						tokenID, err := isc.NativeTokenIDFromBytes(decodedToken)
						log.Check(err)

						if !isc.IsEmptyNativeTokenID(tokenID) {
							gasFeeToken = tokenID.String()
						}
					}

					log.Printf("Gas fee: 1 %s = %d gas units\n", gasFeeToken, chainInfo.GasFeePolicy.GasPerToken)
					log.Printf("Validator fee share: %d%%\n", chainInfo.GasFeePolicy.ValidatorFeeShare)
				}

				log.Printf("Maximum blob size: %d bytes\n", chainInfo.MaxBlobSize)
				log.Printf("Maximum event size: %d bytes\n", chainInfo.MaxEventSize)
				log.Printf("Maximum events per request: %d\n", chainInfo.MaxEventsPerReq)
			}
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	return cmd
}
