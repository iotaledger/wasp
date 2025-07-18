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
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

func initTrustCmd() *cobra.Command {
	var node string

	cmd := &cobra.Command{
		Use:   "trust <name> <pubKey> <peeringURL>",
		Short: "Trust the specified wasp node as a peer.",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			pubKey := args[1]
			peeringURL := args[2]
			node = waspcmd.DefaultWaspNodeFallback(node)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			_, err := cryptolib.DecodeHex(pubKey) // Assert it can be decoded.
			log.Check(err)
			log.Check(peering.CheckPeeringURL(peeringURL))

			_, err = client.NodeAPI.TrustPeer(ctx).PeeringTrustRequest(apiclient.PeeringTrustRequest{
				Name:       name,
				PeeringURL: peeringURL,
				PublicKey:  pubKey,
			}).Execute()
			log.Check(err)
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	return cmd
}
