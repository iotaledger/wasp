// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"strconv"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show information about the chain",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		chainInfo, err := config.WaspClient().GetChainInfo(GetCurrentChainID())
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

			info, err := SCClient(governance.Contract.Hname()).CallView(governance.FuncGetChainInfo.Name, nil)
			log.Check(err)

			description, err := codec.DecodeString(info.MustGet(governance.VarDescription), "")
			log.Check(err)
			log.Printf("Description: %s\n", description)

			recs, err := SCClient(root.Contract.Hname()).CallView(root.FuncGetContractRecords.Name, nil)
			log.Check(err)
			contracts, err := root.DecodeContractRegistry(collections.NewMapReadOnly(recs, root.StateVarContractRegistry))
			log.Check(err)
			log.Printf("#Contracts: %d\n", len(contracts))

			ownerID, err := codec.DecodeAgentID(info.MustGet(governance.VarChainOwnerID))
			log.Check(err)
			log.Printf("Owner: %s\n", ownerID.String())

			if info.MustHas(governance.VarChainOwnerIDDelegated) {
				delegated, err := codec.DecodeAgentID(info.MustGet(governance.VarChainOwnerIDDelegated))
				log.Check(err)
				log.Printf("Delegated owner: %s\n", delegated.String())
			}

			// TODO fees will be changed
			// feeColor, defaultOwnerFee, defaultValidatorFee, err := governance.GetDefaultFeeInfo(info)
			// log.Check(err)
			// log.Printf("Default owner fee: %d %s\n", defaultOwnerFee, feeColor.String())
			// log.Printf("Default validator fee: %d %s\n", defaultValidatorFee, feeColor.String())
		}
	},
}
