// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initTrustCmd() *cobra.Command {
	var nodes []int

	cmd := &cobra.Command{
		Use:   "trust <pubKey> <netID>",
		Short: "Trust the specified wasp node as a peer.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if nodes == nil {
				nodes = chain.GetAllWaspNodes()
			}
			pubKey := args[0]
			netID := args[1]
			_, err := iotago.DecodeHex(pubKey) // Assert it can be decoded.
			log.Check(err)
			log.Check(peering.CheckNetID(netID))
			for _, i := range nodes {
				_, err = config.WaspClient(config.MustWaspAPI(i)).PostPeeringTrusted(pubKey, netID)
				log.Check(err)
			}
		},
	}

	cmd.Flags().IntSliceVarP(&nodes, "nodes", "", nil, "wasp nodes to execute the command in (ex: 0,1,2,3) (default: all nodes)")

	return cmd
}
