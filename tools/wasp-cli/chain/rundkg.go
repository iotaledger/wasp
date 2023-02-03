// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"
	"fmt"
	"math/rand"
	"os"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/multiclient"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initRunDKGCmd() *cobra.Command {
	var (
		committee []int
		quorum    int
	)

	cmd := &cobra.Command{
		Use:   "rundkg",
		Short: "Runs the DKG on specified nodes",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if committee == nil {
				committee = GetAllWaspNodes()
			}
			if quorum == 0 {
				quorum = defaultQuorum(len(committee))
			}

			committeePubKeys := make([]string, 0)
			for _, apiIndex := range committee {
				peerInfo, _, err := cliclients.WaspClientForIndex(apiIndex).NodeApi.GetPeeringIdentity(context.Background()).Execute()
				log.Check(err)

				committeePubKeys = append(committeePubKeys, peerInfo.PublicKey)
			}

			dkgInitiatorIndex := uint16(rand.Intn(len(committee)))

			var clientResolver multiclient.ClientResolver = func(apiHost string) *apiclient.APIClient {
				return cliclients.WaspClientForHostName(apiHost)
			}

			stateControllerAddr, err := apilib.RunDKG(clientResolver, config.CommitteeAPIURL(committee), committeePubKeys, uint16(quorum), dkgInitiatorIndex)
			log.Check(err)

			fmt.Fprintf(os.Stdout, "DKG successful, address: %s", stateControllerAddr.Bech32(parameters.L1().Protocol.Bech32HRP))
		},
	}

	cmd.Flags().IntSliceVarP(&committee, "committee", "", nil, "peers acting as committee nodes (ex: 0,1,2,3) (default: all nodes)")
	cmd.Flags().IntVarP(&quorum, "quorum", "", 0, "quorum")
	return cmd
}
