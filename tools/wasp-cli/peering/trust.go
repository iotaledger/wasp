// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"context"

	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initTrustCmd() *cobra.Command {
	var nodes []string

	cmd := &cobra.Command{
		Use:   "trust <pubKey> <netID>",
		Short: "Trust the specified wasp node as a peer.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			pubKey := args[0]
			netID := args[1]
			_, err := iotago.DecodeHex(pubKey) // Assert it can be decoded.
			log.Check(err)
			log.Check(peering.CheckNetID(netID))

			for _, node := range nodes {
				client := cliclients.WaspClient(node)

				_, err = client.NodeApi.TrustPeer(context.Background()).PeeringTrustRequest(apiclient.PeeringTrustRequest{
					NetId:     netID,
					PublicKey: pubKey,
				}).Execute()
				log.Check(err)
			}
		},
	}

	waspcmd.WithWaspNodesFlag(cmd, &nodes)

	return cmd
}
