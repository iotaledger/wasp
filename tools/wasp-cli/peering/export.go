package peering

import (
	"context"
	"encoding/json"
	"os"
	"slices"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

func initExportTrustedJSONCmd() *cobra.Command {
	var node string
	var peers []string
	var outputFile string

	cmd := &cobra.Command{
		Use:   "export-trusted",
		Short: "List trusted wasp nodes.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)

			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)
			trustedList, _, err := client.NodeAPI.GetTrustedPeers(context.Background()).Execute()
			log.Check(err)

			var filteredList []apiclient.PeeringNodeIdentityResponse

			// filter the list according to user specification
			if len(peers) == 0 {
				filteredList = trustedList
			} else {
				filteredList = lo.Filter(trustedList, func(peer apiclient.PeeringNodeIdentityResponse, _ int) bool {
					return slices.Contains(peers, peer.Name)
				})
			}

			// warn user if exporting untrusted peers, or there are peers missing
			for _, peer := range filteredList {
				if !peer.IsTrusted {
					log.Printf("WARN: untrusted peer {%s} was exported\n", peer.Name)
				}
			}
			if len(filteredList) != len(peers) {
				for _, peerName := range peers {
					exported := !lo.ContainsBy(filteredList, func(filteredPeer apiclient.PeeringNodeIdentityResponse) bool {
						return filteredPeer.Name == peerName
					})
					if !exported {
						log.Printf("WARN: unknown peer {%s}, won't be exported \n", peerName)
					}
				}
			}

			data, err := json.Marshal(filteredList)
			log.Check(err)
			if outputFile == "" {
				log.Printf("%s\n", data)
				return
			}

			file, err := os.Create(outputFile)
			if err != nil {
				log.Fatal(err)
			}
			_, err = file.Write(data)
			log.Check(err)
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	waspcmd.WithPeersFlag(cmd, &peers)
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "file where the exported list will be saved to")

	return cmd
}

func initImportTrustedJSONCmd() *cobra.Command {
	var node string

	cmd := &cobra.Command{
		Use:   "import-trusted <file path>",
		Short: "imports a JSON of trusted peers and makes a node trust them.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			bytes := util.ReadFile(args[0])
			var trustedList []apiclient.PeeringNodeIdentityResponse
			log.Check(json.Unmarshal(bytes, &trustedList))

			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)
			identity, _, err := client.NodeAPI.GetPeeringIdentity(ctx).Execute()
			log.Check(err)

			for _, t := range trustedList {
				if !t.IsTrusted {
					continue // avoid importing untrusted peers by mistake
				}

				if t.PublicKey == identity.PublicKey {
					continue
				}

				_, err := client.NodeAPI.TrustPeer(ctx).PeeringTrustRequest(apiclient.PeeringTrustRequest{
					Name:       t.Name,
					PeeringURL: t.PeeringURL,
					PublicKey:  t.PublicKey,
				}).Execute()
				log.Check(err)
			}
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)

	return cmd
}
