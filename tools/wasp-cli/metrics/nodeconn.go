package metrics

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/log"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
)

const maxMessageLen = 80

var nodeconnMetricsCmd = &cobra.Command{
	Use:   "nodeconn",
	Short: "Show current value of collected metrics of connection to L1",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := config.WaspClient()
		if chainAlias == "" {
			nodeconnMetrics, err := client.GetNodeConnectionMetrics()
			log.Check(err)
			log.Printf("Following chains are registered for L1 events:\n")
			for _, s := range nodeconnMetrics.Registered {
				log.Printf("\t%s\n", s)
			}
			printMessagesMetrics(
				&nodeconnMetrics.NodeConnectionMessagesMetrics,
				[][]string{makeMessagesMetricsTableRow("Milestone", true, nodeconnMetrics.InMilestone)},
			)
		} else {
			chid, err := isc.ChainIDFromString(chainAlias)
			log.Check(err)
			msgsMetrics, err := client.GetChainNodeConnectionMetrics(chid)
			log.Check(err)
			printMessagesMetrics(msgsMetrics, [][]string{})
		}
	},
}

func printMessagesMetrics(msgsMetrics *model.NodeConnectionMessagesMetrics, additionalRows [][]string) {
	header := []string{"Message name", "", "Total", "Last time", "Last message"}
	table := [][]string{
		makeMessagesMetricsTableRow("Publish state transaction", false, msgsMetrics.OutPublishStateTransaction),
		makeMessagesMetricsTableRow("Publish governance transaction", false, msgsMetrics.OutPublishGovernanceTransaction),
		makeMessagesMetricsTableRow("Pull latest output", false, msgsMetrics.OutPullLatestOutput),
		makeMessagesMetricsTableRow("Pull tx inclusion state", false, msgsMetrics.OutPullTxInclusionState),
		makeMessagesMetricsTableRow("Pull output by ID", false, msgsMetrics.OutPullOutputByID),
		makeMessagesMetricsTableRow("State output", true, msgsMetrics.InStateOutput),
		makeMessagesMetricsTableRow("Alias output", true, msgsMetrics.InAliasOutput),
		makeMessagesMetricsTableRow("Output", true, msgsMetrics.InOutput),
		makeMessagesMetricsTableRow("On ledger request", true, msgsMetrics.InOnLedgerRequest),
		makeMessagesMetricsTableRow("Tx inclusion state", true, msgsMetrics.InTxInclusionState),
	}
	table = append(table, additionalRows...)
	log.PrintTable(header, table)
}

func makeMessagesMetricsTableRow(name string, isIn bool, ncmm *model.NodeConnectionMessageMetrics) []string {
	res := make([]string, 5)
	res[0] = name
	if isIn {
		res[1] = "IN"
	} else {
		res[1] = "OUT"
	}
	res[2] = fmt.Sprintf("%v", ncmm.Total)
	res[3] = ncmm.LastEvent.String()
	res[4] = ncmm.LastMessage
	if len(res[4]) > maxMessageLen {
		res[4] = res[4][:maxMessageLen]
	}
	res[4] = strings.Replace(res[4], "\n", " ", -1)
	return res
}
