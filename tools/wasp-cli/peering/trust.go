// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/peering"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

func initTrustCmd() *cobra.Command {
	var node string

	cmd := &cobra.Command{
		Use:   "trust <name> <pubKey> <peeringURL>",
		Short: "Trust the specified wasp node as a peer.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			pubKey := args[1]
			peeringURL := args[2]
			var err error
			node, err = waspcmd.DefaultWaspNodeFallback(node)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			_, err = cryptolib.DecodeHex(pubKey) // Assert it can be decoded.
			if err != nil {
				return err
			}
			if err = peering.CheckPeeringURL(peeringURL); err != nil {
				return err
			}

			_, err = client.NodeAPI.TrustPeer(ctx).PeeringTrustRequest(apiclient.PeeringTrustRequest{
				Name:       name,
				PeeringURL: peeringURL,
				PublicKey:  pubKey,
			}).Execute()
			return err
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	return cmd
}
