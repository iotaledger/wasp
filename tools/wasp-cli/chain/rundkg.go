// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"
	"fmt"
	"os"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initRunDKGCmd() *cobra.Command {
	var (
		node   string
		peers  []string
		quorum int
	)

	cmd := &cobra.Command{
		Use:   "rundkg --peers=...",
		Short: "Runs the DKG on specified nodes",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			doDKG(node, peers, quorum)
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	waspcmd.WithPeersFlag(cmd, &peers)
	log.Check(cmd.MarkFlagRequired("peers"))
	cmd.Flags().IntVarP(&quorum, "quorum", "", 0, "quorum (default: 3/4s of the number of committee nodes)")
	return cmd
}

func isEnoughQuorum(n, t int) (bool, int) {
	maxF := (n - 1) / 3
	return t >= (n - maxF), maxF
}

func doDKG(node string, peers []string, quorum int) iotago.Address {
	if quorum == 0 {
		quorum = defaultQuorum(len(peers) + 1)
	}

	if ok, _ := isEnoughQuorum(len(peers)+1, quorum); !ok {
		log.Fatal("quorum needs to be at least (2/3)+1 of committee size")
	}

	client := cliclients.WaspClient(node)

	// grab the peering info of the peers from the node
	filteredPeers := make([]apiclient.PeeringNodeIdentityResponse, 0)
	{
		trustedPeers, _, err := client.NodeApi.GetTrustedPeers(context.Background()).Execute() //nolint:bodyclose // false positive
		log.Check(err)

		for _, peer := range peers {
			foundPeer, exists := lo.Find(trustedPeers, func(p apiclient.PeeringNodeIdentityResponse) bool {
				return p.Name == peer
			})
			if !exists {
				log.Fatalf("peer with name {%s} not found in trusted peers", peer)
			}
			filteredPeers = append(filteredPeers, foundPeer)
		}
	}

	// construct a list that includes the own node pub key and the peers
	info, _, err := client.NodeApi.GetPeeringIdentity(context.Background()).Execute() //nolint:bodyclose // false positive
	log.Check(err)

	committeePubKeys := []string{info.PublicKey}
	for _, peer := range filteredPeers {
		committeePubKeys = append(committeePubKeys, peer.PublicKey)
	}

	stateControllerAddr, err := apilib.RunDKG(client, committeePubKeys, uint16(quorum))
	log.Check(err)

	fmt.Fprintf(os.Stdout, "DKG successful, address: %s", stateControllerAddr.Bech32(parameters.L1().Protocol.Bech32HRP))
	return stateControllerAddr
}
