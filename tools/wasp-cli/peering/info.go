// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Node info.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		info, err := config.WaspClient(config.MustWaspAPI()).GetPeeringSelf()
		log.Check(err)

		model := &InfoModel{PubKey: info.PubKey, NetID: info.NetID}
		log.PrintCLIOutput(model)
	},
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
