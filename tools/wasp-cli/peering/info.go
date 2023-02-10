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
	var peers []string
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Node info.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			for _, node := range peers {
				client := cliclients.WaspClient(node)
				info, _, err := client.NodeApi.GetPeeringIdentity(context.Background()).Execute()
				log.Check(err)

				model := &InfoModel{PubKey: info.PublicKey, PeeringURL: info.PeeringURL}
				log.PrintCLIOutput(model)
			}
		},
	}
	waspcmd.WithPeersFlag(cmd, &peers)
	return cmd
}

type InfoModel struct {
	PubKey     string
	PeeringURL string
}

func (i *InfoModel) AsText() (string, error) {
	infoTemplate := `PubKey: {{ .PubKey }}
peeringURL: {{ .peeringURL }}`
	return log.ParseCLIOutputTemplate(i, infoTemplate)
}
