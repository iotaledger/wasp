// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"os"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
	"github.com/spf13/cobra"
)

func deployCmd() *cobra.Command {
	var (
		committee   []int
		quorum      int
		description string
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new chain",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			alias := GetChainAlias()

			if committee == nil {
				committee = []int{0, 1, 2, 3}
			}

			committeePubKeys := make([]string, 0)
			for _, api := range config.CommitteeAPI(committee) {
				peerInfo, err := client.NewWaspClient(api).GetPeeringSelf()
				log.Check(err)
				committeePubKeys = append(committeePubKeys, peerInfo.PubKey)
			}

			chainid, _, err := apilib.DeployChainWithDKG(apilib.CreateChainParams{
				Layer1Client:      config.GoshimmerClient(),
				CommitteeAPIHosts: config.CommitteeAPI(committee),
				CommitteePubKeys:  committeePubKeys,
				N:                 uint16(len(committee)),
				T:                 uint16(quorum),
				OriginatorKeyPair: wallet.Load().KeyPair,
				Description:       description,
				Textout:           os.Stdout,
			})
			log.Check(err)

			AddChainAlias(alias, chainid.String())
		},
	}

	cmd.Flags().IntSliceVarP(&committee, "committee", "", nil, "peers acting as committee nodes  (default: 0,1,2,3)")
	cmd.Flags().IntVarP(&quorum, "quorum", "", 3, "quorum")
	cmd.Flags().StringVarP(&description, "description", "", "", "description")
	return cmd
}
