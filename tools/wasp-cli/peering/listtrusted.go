// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initListTrustedCmd() *cobra.Command {
	var printJSON bool
	var node string

	cmd := &cobra.Command{
		Use:   "list-trusted",
		Short: "List trusted wasp nodes.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)

			client := cliclients.WaspClient(node)
			trustedList, _, err := client.NodeApi.GetTrustedPeers(context.Background()).Execute()
			log.Check(err)

			if printJSON {
				data, err := json.Marshal(trustedList)
				log.Check(err)
				log.Printf("%s\n", data)
				return
			}
			header := []string{"PubKey", "NetID"}
			rows := make([][]string, len(trustedList))
			for i := range rows {
				rows[i] = []string{
					trustedList[i].PublicKey,
					trustedList[i].NetId,
				}
			}
			log.PrintTable(header, rows)
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	cmd.Flags().BoolVar(&printJSON, "json", false, "output in JSON")

	return cmd
}

func initImportTrustedJSONCmd() *cobra.Command {
	var node string

	cmd := &cobra.Command{
		Use:   "import-trusted",
		Short: "imports a JSON of trusted peers and makes a node trust them.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			bytes := util.ReadFile(args[0])
			var trustedList []apiclient.PeeringNodeIdentityResponse
			log.Check(json.Unmarshal(bytes, &trustedList))
			for _, t := range trustedList {
				client := cliclients.WaspClient(node)
				if !t.IsTrusted {
					continue // avoid importing untrusted peers by mistake
				}
				_, err := client.NodeApi.TrustPeer(context.Background()).PeeringTrustRequest(apiclient.PeeringTrustRequest{
					NetId:     t.NetId,
					PublicKey: t.PublicKey,
				}).Execute()
				log.Check(err)
			}
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)

	return cmd
}
