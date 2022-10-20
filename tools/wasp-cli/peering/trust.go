// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var trustCmd = &cobra.Command{
	Use:   "trust <pubKey> <netID>",
	Short: "Trust the specified wasp node as a peer.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		pubKey := args[0]
		netID := args[1]
		_, err := hexutil.Decode(pubKey) // Assert it can be decoded.
		log.Check(err)
		log.Check(peering.CheckNetID(netID))
		_, err = config.WaspClient().PostPeeringTrusted(pubKey, netID)
		log.Check(err)
	},
}
