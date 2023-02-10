// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initInfoCmd() *cobra.Command {
	var nodes []string
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Node info.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			nodes = waspcmd.DefaultNodesFallback(nodes)

			for _, node := range nodes {
				client := cliclients.WaspClient(node)
				info, _, err := client.NodeApi.GetPeeringIdentity(context.Background()).Execute()
				log.Check(err)

				model := &InfoModel{PubKey: info.PublicKey, NetID: info.NetId}
				log.PrintCLIOutput(model)
			}
		},
	}
	waspcmd.WithWaspNodesFlag(cmd, &nodes)
	return cmd
}

type InfoModel struct {
	PubKey string
	NetID  string
}

func (i *InfoModel) AsText() (string, error) {
	infoTemplate := `PubKey: {{ .PubKey }}
NetID: {{ .NetID }}`
	return log.ParseCLIOutputTemplate(i, infoTemplate)
}
