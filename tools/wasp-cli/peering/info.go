// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

func initInfoCmd() *cobra.Command {
	var node string
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Node peering info.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			node, err = waspcmd.DefaultWaspNodeFallbackE(node)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)
			info, _, err := client.NodeAPI.GetPeeringIdentity(ctx).Execute()
			if err != nil {
				return err
			}

			model := &InfoModel{PubKey: info.PublicKey, PeeringURL: info.PeeringURL}
			log.PrintCLIOutput(model)
			return nil
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	return cmd
}

type InfoModel struct {
	PubKey     string `json:"pubKey"`
	PeeringURL string `json:"peeringURL"`
}

func (i *InfoModel) AsText() (string, error) {
	infoTemplate := `PubKey: {{ .PubKey }}
PeeringURL: {{ .PeeringURL }}`
	return log.ParseCLIOutputTemplate(i, infoTemplate)
}

func (i *InfoModel) AsJSON() (string, error) {
	jsonBytes, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
