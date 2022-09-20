// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"github.com/iotaledger/wasp/packages/webapi/v1/model"
	"strconv"

	"github.com/iotaledger/wasp/tools/wasp-cli/util"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show information about the chain",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		chainInfo, err := config.WaspClient(config.MustWaspAPI()).GetChainInfo(GetCurrentChainID())
		log.Check(err)

		printNodesRowHdr := []string{"PubKey", "NetID", "Alive", "Committee", "Access", "AccessAPI"}
		printNodesRowFmt := func(n *model.ChainNodeStatus) []string {
			return []string{
				n.Node.PubKey,
				n.Node.NetID,
				strconv.FormatBool(n.Node.IsAlive),
				strconv.FormatBool(n.ForCommittee),
				strconv.FormatBool(n.ForAccess),
				n.AccessAPI,
			}
		}
		printNodes := func(label string, nodes []*model.ChainNodeStatus) {
			if nodes == nil {
				log.Printf("%s: N/A\n", label)
			}
			log.Printf("%s: %v\n", label, len(nodes))
			rows := make([][]string, 0)
			for _, n := range nodes {
				rows = append(rows, printNodesRowFmt(n))
			}
			log.PrintTable(printNodesRowHdr, rows)
		}

		log.Printf("Chain ID: %s\n", chainInfo.ChainID)
		log.Printf("Active: %v\n", chainInfo.Active)

		if chainInfo.Active {
			log.Printf("State address: %v\n", chainInfo.StateAddress)
			printNodes("Committee nodes", chainInfo.CommitteeNodes)
			printNodes("Access nodes", chainInfo.AccessNodes)
			printNodes("Candidate nodes", chainInfo.CandidateNodes)

			ret, err := SCClient(governance.Contract.Hname()).CallView(governance.ViewGetChainInfo.Name, nil)
			log.Check(err)
			govInfo, err := governance.GetChainInfo(ret)
			log.Check(err)

			log.Printf("Description: %s\n", govInfo.Description)

			recs, err := SCClient(root.Contract.Hname()).CallView(root.ViewGetContractRecords.Name, nil)
			log.Check(err)
			contracts, err := root.DecodeContractRegistry(collections.NewMapReadOnly(recs, root.StateVarContractRegistry))
			log.Check(err)
			log.Printf("#Contracts: %d\n", len(contracts))

			log.Printf("Owner: %s\n", govInfo.ChainOwnerID.String())

			if govInfo.GasFeePolicy != nil {
				gasFeeToken := util.BaseTokenStr
				if govInfo.GasFeePolicy.GasFeeTokenID != nil {
					gasFeeToken = govInfo.GasFeePolicy.GasFeeTokenID.String()
				}
				log.Printf("Gas fee: 1 %s = %d gas units\n", gasFeeToken, govInfo.GasFeePolicy.GasPerToken)
				log.Printf("Validator fee share: %d%%\n", govInfo.GasFeePolicy.ValidatorFeeShare)
			}

			log.Printf("Maximum blob size: %d bytes\n", govInfo.MaxBlobSize)
			log.Printf("Maximum event size: %d bytes\n", govInfo.MaxEventSize)
			log.Printf("Maximum events per request: %d\n", govInfo.MaxEventsPerReq)
		}
	},
}
