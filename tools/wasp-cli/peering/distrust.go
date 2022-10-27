// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/log"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
)

var distrustCmd = &cobra.Command{
	Use:   "distrust <pubKey|netID>",
	Short: "Remove the specified node from a list of trusted nodes. All related public keys are distrusted, if netID is provided.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pubKeyOrNetID := args[0]
		waspClient := config.WaspClient()
		if peering.CheckNetID(pubKeyOrNetID) != nil {
			log.Check(waspClient.DeletePeeringTrusted(pubKeyOrNetID))
			log.Printf("# Distrusted PubKey: %v\n", pubKeyOrNetID)
			return
		}
		trustedList, err := waspClient.GetPeeringTrustedList()
		log.Check(err)
		for _, t := range trustedList {
			if t.NetID == pubKeyOrNetID {
				err := waspClient.DeletePeeringTrusted(t.PubKey)
				if err != nil {
					log.Printf("error: failed to distrust %v/%v, reason=%v\n", t.PubKey, t.NetID, err)
				} else {
					log.Printf("# Distrusted PubKey: %v\n", t.PubKey)
				}
			}
		}
	},
}
