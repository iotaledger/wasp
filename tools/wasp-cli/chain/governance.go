// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"

	"github.com/samber/lo"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initChangeAccessNodesCmd() *cobra.Command {
	var offLedger bool
	var node string
	var chain string

	cmd := &cobra.Command{
		Use:   "gov-change-access-nodes <action (accept|remove|drop)> <pubkey>",
		Short: "Changes the access nodes of a chain on the governance contract.",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)

			if len(args)%2 != 0 {
				log.Fatal("wrong number of arguments")
			}
			pars := make([]lo.Tuple2[*cryptolib.PublicKey, governance.ChangeAccessNodeAction], 0)

			for i := 1; i < len(args); i += 2 {
				pubkey, err := cryptolib.PublicKeyFromString(args[i])
				action := args[i-1]
				log.Check(err)

				var actionResult governance.ChangeAccessNodeAction

				switch action {
				case "accept":
					actionResult = governance.ChangeAccessNodeActionAccept
				case "remove":
					actionResult = governance.ChangeAccessNodeActionRemove
				case "drop":
					actionResult = governance.ChangeAccessNodeActionDrop
				default:
					log.Fatal("invalid action")
				}

				pars = append(pars, lo.T2(pubkey, actionResult))
			}
			postRequest(
				node,
				chain,
				governance.FuncChangeAccessNodes.Message(pars),
				chainclient.PostRequestParams{},
				offLedger,
				true,
			)
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	cmd.Flags().BoolVarP(&offLedger, "off-ledger", "o", false,
		"post an off-ledger request",
	)

	return cmd
}

func initDisableFeePolicyCmd() *cobra.Command {
	var offLedger bool
	var node string
	var chain string

	cmd := &cobra.Command{
		Use:   "disable-feepolicy",
		Short: "set token charged by each gas to free.",
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)
			client := cliclients.WaspClient(node)

			callGovView := func(viewName string) isc.CallResults {
				apiResult, _, err := client.ChainsAPI.CallView(context.Background(), config.GetChain(chain).String()).
					ContractCallViewRequest(apiclient.ContractCallViewRequest{
						ContractName: governance.Contract.Name,
						FunctionName: viewName,
					}).Execute() //nolint:bodyclose // false positive
				log.Check(err)

				result, err := apiextensions.APIResultToCallArgs(apiResult)
				log.Check(err)
				return result
			}

			r := callGovView(governance.ViewGetFeePolicy.Name)
			feePolicy, err := governance.ViewGetFeePolicy.DecodeOutput(r)
			log.Check(err)
			feePolicy.GasPerToken = util.Ratio32{}

			postRequest(
				node,
				chain,
				governance.FuncSetFeePolicy.Message(feePolicy),
				chainclient.PostRequestParams{},
				offLedger,
				true,
			)
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	cmd.Flags().BoolVarP(&offLedger, "off-ledger", "o", false,
		"post an off-ledger request",
	)

	return cmd
}
