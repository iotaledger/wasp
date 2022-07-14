// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"fmt"
	"os"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
	"github.com/spf13/cobra"
)

var rotateCmd = &cobra.Command{
	Use:   "rotate <new state controller address>",
	Short: "Issues a tx that changes the chain state controller",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		l1Client := config.L1Client()
		prefix, newStateControllerAddr, err := iotago.ParseBech32(args[0])
		log.Check(err)
		if parameters.L1.Protocol.Bech32HRP != prefix {
			log.Fatalf("unexpected prefix. expected: %s, actual: %s", parameters.L1.Protocol.Bech32HRP, prefix)
		}

		kp := wallet.Load().KeyPair

		l1Client.OutputMap(kp.Address())

		tx := transaction.NewRotateChainStateControllerTx(
			GetCurrentChainID().AsAliasID(),
			newStateControllerAddr,
		)
		err = l1Client.PostTx((tx))
		log.Check(err)

		println(newStateControllerAddr)

		// committeePubKeys := make([]string, 0)
		// for _, api := range config.CommitteeAPI(committee) {
		// 	peerInfo, err := client.NewWaspClient(api).GetPeeringSelf()
		// 	log.Check(err)
		// 	committeePubKeys = append(committeePubKeys, peerInfo.PubKey)
		// }

		// dkgInitiatorIndex := uint16(rand.Intn(len(committee)))
		// stateControllerAddr, err := apilib.RunDKG(config.CommitteeAPI(committee), committeePubKeys, uint16(quorum), dkgInitiatorIndex)
		// log.Check(err)

		txID := "" // TODO

		fmt.Fprintf(os.Stdout, "chain rotation transaction issued successfully. TXID: %s", txID)
	},
}
