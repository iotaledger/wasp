// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Node info.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			client := cliclients.WaspClientForIndex()
			info, _, err := client.NodeApi.GetPeeringIdentity(context.Background()).Execute()
			log.Check(err)

			model := &InfoModel{PubKey: info.PublicKey, NetID: info.NetId}
			log.PrintCLIOutput(model)
		},
	}
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
