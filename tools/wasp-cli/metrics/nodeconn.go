package metrics

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/tools/wasp-cli/chain"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
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
			chid, err := iscp.ChainIDFromString(chainIDStr)
			log.Check(err)
			msgsMetrics, err := client.GetChainNodeConnectionMetrics(chid)
			log.Check(err)
			printMessagesMetrics(msgsMetrics, [][]string{})
		}
	},
}

func printMessagesMetrics(msgsMetrics *model.NodeConnectionMessagesMetrics, additionalRows [][]string) {
	header := []string{"Message name", "", "Total", "Last time", "Last message"}
	table := make([][]string, 9+len(additionalRows))
	table[0] = makeMessagesMetricsTableRow("Publish transaction", false, msgsMetrics.OutPublishTransaction)
	table[1] = makeMessagesMetricsTableRow("Pull latest output", false, msgsMetrics.OutPullLatestOutput)
	table[2] = makeMessagesMetricsTableRow("Pull tx inclusion state", false, msgsMetrics.OutPullTxInclusionState)
	table[3] = makeMessagesMetricsTableRow("Pull output by ID", false, msgsMetrics.OutPullOutputByID)
	table[4] = makeMessagesMetricsTableRow("State output", true, msgsMetrics.InStateOutput)
	table[5] = makeMessagesMetricsTableRow("Alias output", true, msgsMetrics.InAliasOutput)
	table[6] = makeMessagesMetricsTableRow("Output", true, msgsMetrics.InOutput)
	table[7] = makeMessagesMetricsTableRow("On ledger request", true, msgsMetrics.InOnLedgerRequest)
	table[8] = makeMessagesMetricsTableRow("Tx inclusion state", true, msgsMetrics.InTxInclusionState)
	for i := range additionalRows {
		table[9+i] = additionalRows[i]
	}
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
