// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/clients/apiextensions"
	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

func initChangeAccessNodesCmd() *cobra.Command {
	var offLedger bool
	var node string
	var chain string

	cmd := &cobra.Command{
		Use:   "gov-change-access-nodes <action (accept|remove|drop)> <pubkey>",
		Short: "Changes the access nodes of a chain on the governance contract.",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			node, err = waspcmd.DefaultWaspNodeFallback(node)
			if err != nil {
				return err
			}
			chain = defaultChainFallback(chain)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)
			if len(args)%2 != 0 {
				return fmt.Errorf("wrong number of arguments")
			}
			pars := make(governance.ChangeAccessNodeActions, 0)

			for i := 1; i < len(args); i += 2 {
				pubkey, err := cryptolib.PublicKeyFromString(args[i])
				action := args[i-1]
				if err != nil {
					return err
				}

				var actionResult governance.ChangeAccessNodeAction

				switch action {
				case "accept":
					actionResult = governance.ChangeAccessNodeActionAccept
				case "remove":
					actionResult = governance.ChangeAccessNodeActionRemove
				case "drop":
					actionResult = governance.ChangeAccessNodeActionDrop
				default:
					return fmt.Errorf("invalid action")
				}

				pars = append(pars, lo.T2(pubkey, actionResult))
			}

			postRequest(
				ctx,
				client,
				chain,
				governance.FuncChangeAccessNodes.Message(pars),
				chainclient.PostRequestParams{
					GasBudget: iotaclient.DefaultGasBudget,
				},
				offLedger,
			)
			return nil
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
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			node, err = waspcmd.DefaultWaspNodeFallback(node)
			if err != nil {
				return err
			}
			chain = defaultChainFallback(chain)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			callGovView := func(viewName string) (isc.CallResults, error) {
				var apiResult []string
				apiResult, _, err = client.ChainsAPI.CallView(ctx).
					ContractCallViewRequest(apiclient.ContractCallViewRequest{
						ContractName: governance.Contract.Name,
						FunctionName: viewName,
					}).Execute() //nolint:bodyclose // false positive
				if err != nil {
					return nil, err
				}
				var result isc.CallResults
				result, err = apiextensions.APIResultToCallArgs(apiResult)
				if err != nil {
					return nil, err
				}
				return result, nil
			}

			r, err := callGovView(governance.ViewGetFeePolicy.Name)
			if err != nil {
				return err
			}
			feePolicy, err := governance.ViewGetFeePolicy.DecodeOutput(r)
			if err != nil {
				return err
			}
			feePolicy.GasPerToken = util.Ratio32{}

			postRequest(
				ctx,
				client,
				chain,
				governance.FuncSetFeePolicy.Message(feePolicy),
				chainclient.PostRequestParams{
					GasBudget:   iotaclient.DefaultGasBudget,
					L2GasBudget: 1 * isc.Million,
				},
				offLedger,
			)
			return nil
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	cmd.Flags().BoolVarP(&offLedger, "off-ledger", "o", false,
		"post an off-ledger request",
	)

	return cmd
}
