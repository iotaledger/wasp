// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"
	"fmt"
	"math/rand"
	"os"

	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/multiclient"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initRunDKGCmd() *cobra.Command {
	var (
		nodes  []string
		quorum int
	)

	cmd := &cobra.Command{
		Use:   "rundkg",
		Short: "Runs the DKG on specified nodes",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			doDKG(nodes, quorum)
		},
	}

	waspcmd.WithWaspNodesFlag(cmd, &nodes)
	log.Check(cmd.MarkFlagRequired("nodes"))
	cmd.Flags().IntVarP(&quorum, "quorum", "", 0, "quorum (default: 3/4s of the number of committee nodes)")
	return cmd
}

func doDKG(nodes []string, quorum int) iotago.Address {
	if len(nodes) == 0 {
		nodes = []string{config.MustGetDefaultWaspNode()}
	}
	if quorum == 0 {
		quorum = defaultQuorum(len(nodes))
	}

	committeePubKeys := make([]string, 0)
	for _, apiIndex := range nodes {
		peerInfo, _, err := cliclients.WaspClient(apiIndex).NodeApi.GetPeeringIdentity(context.Background()).Execute()
		log.Check(err)

		committeePubKeys = append(committeePubKeys, peerInfo.PublicKey)
	}

	dkgInitiatorIndex := uint16(rand.Intn(len(nodes)))

	var clientResolver multiclient.ClientResolver = cliclients.WaspClientForHostName

	stateControllerAddr, err := apilib.RunDKG(clientResolver, config.NodeAPIURLs(nodes), committeePubKeys, uint16(quorum), dkgInitiatorIndex)
	log.Check(err)

	fmt.Fprintf(os.Stdout, "DKG successful, address: %s", stateControllerAddr.Bech32(parameters.L1().Protocol.Bech32HRP))
	return stateControllerAddr
}
