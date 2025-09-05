// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/packages/peering"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

func initDistrustCmd() *cobra.Command {
	var node string
	cmd := &cobra.Command{
		Use:   "distrust <name|pubKey>",
		Short: "Remove the specified node from a list of trusted nodes. All related public keys are distrusted, if peeringURL is provided.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := args[0]

			var err error
			node, err = waspcmd.DefaultWaspNodeFallback(node)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			if peering.CheckPeeringURL(input) != nil {
				_, err = client.NodeAPI.DistrustPeer(ctx, input).Execute()
				if err != nil {
					return err
				}
				log.Printf("# Distrusted PubKey: %v\n", input)
				return nil
			}

			trustedList, _, err := client.NodeAPI.GetTrustedPeers(ctx).Execute()
			if err != nil {
				return err
			}

			for _, t := range trustedList {
				if t.PublicKey == input {
					_, err := client.NodeAPI.DistrustPeer(ctx, t.PublicKey).Execute()

					if err != nil {
						log.Printf("error: failed to distrust %v/%v, reason=%v\n", t.PublicKey, t.PeeringURL, err)
					} else {
						log.Printf("# Distrusted PubKey: %v\n", t.PublicKey)
					}
				}
			}
			return nil
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	return cmd
}
