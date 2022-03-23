package metrics

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/webapi/model"
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
		if chainIDStr == "" {
			nodeconnMetrics, err := client.GetNodeConnectionMetrics()
			log.Check(err)
			log.Printf("Following chains are registered for L1 events:\n")
			for _, s := range nodeconnMetrics.Registered {
				log.Printf("\t%s\n", s)
			}
			printMessagesMetrics(&nodeconnMetrics.NodeConnectionMessagesMetrics)
		} else {
			chid, err := iscp.ChainIDFromString(chainIDStr)
			log.Check(err)
			msgsMetrics, err := client.GetChainNodeConnectionMetrics(chid)
			log.Check(err)
			printMessagesMetrics(msgsMetrics)
		}
	},
}

func printMessagesMetrics(msgsMetrics *model.NodeConnectionMessagesMetrics) {
	header := []string{"Message name", "", "Total", "Last time", "Last message"}
	table := make([][]string, 8)
	table[0] = makeMessagesMetricsTableRow("Pull state", false, msgsMetrics.OutPullState)
	table[1] = makeMessagesMetricsTableRow("Pull tx inclusion state", false, msgsMetrics.OutPullTransactionInclusionState)
	table[2] = makeMessagesMetricsTableRow("Pull confirmed output", false, msgsMetrics.OutPullConfirmedOutput)
	table[3] = makeMessagesMetricsTableRow("Post transaction", false, msgsMetrics.OutPostTransaction)
	table[4] = makeMessagesMetricsTableRow("Transaction", true, msgsMetrics.InTransaction)
	table[5] = makeMessagesMetricsTableRow("Inclusion state", true, msgsMetrics.InInclusionState)
	table[6] = makeMessagesMetricsTableRow("Output", true, msgsMetrics.InOutput)
	table[7] = makeMessagesMetricsTableRow("Unspent alias output", true, msgsMetrics.InUnspentAliasOutput)
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
