// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
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
				committee = getAllWaspNodes()
			}
			if quorum == 0 {
				quorum = defaultQuorum(len(committee))
			}

			committeePubKeys := make([]string, 0)
			for _, api := range config.CommitteeAPI(committee) {
				peerInfo, err := client.NewWaspClient(api).GetPeeringSelf()
				log.Check(err)
				committeePubKeys = append(committeePubKeys, peerInfo.PubKey)
			}

			dkgInitiatorIndex := uint16(rand.Intn(len(committee)))
			stateControllerAddr, err := apilib.RunDKG(config.GetToken(), config.CommitteeAPI(committee), committeePubKeys, uint16(quorum), dkgInitiatorIndex)
			log.Check(err)

			fmt.Fprintf(os.Stdout, "DKG successful, address: %s", stateControllerAddr.Bech32(parameters.L1().Protocol.Bech32HRP))
		},
	}

	cmd.Flags().IntSliceVarP(&committee, "committee", "", nil, "peers acting as committee nodes (ex: 0,1,2,3) (default: all nodes)")
	cmd.Flags().IntVarP(&quorum, "quorum", "", 0, "quorum")
	return cmd
}
