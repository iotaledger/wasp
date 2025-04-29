package metrics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

const maxMessageLen = 80

func initNodeconnMetricsCmd() *cobra.Command {
	var node string
	cmd := &cobra.Command{
		Use:   "nodeconn",
		Short: "Show current value of collected metrics of connection to L1",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			msgsMetrics, _, err := client.MetricsAPI.GetChainMessageMetrics(ctx).Execute()
			log.Check(err)
			printChainMessagesMetrics(msgsMetrics)
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	return cmd
}

func mapMetricItem(messages uint32, timestamp time.Time, message string) *apiclient.InterfaceMetricItem {
	return &apiclient.InterfaceMetricItem{
		Timestamp:   timestamp,
		LastMessage: message,
		Messages:    messages,
	}
}

func printChainMessagesMetrics(msgsMetrics *apiclient.ChainMessageMetrics) {
	header := []string{"Message name", "", "Total", "Last time", "Last message"}

	inAnchor := mapMetricItem(msgsMetrics.InAnchor.Messages, msgsMetrics.InAnchor.Timestamp, msgsMetrics.InAnchor.LastMessage.Raw)
	inOnLedgerRequest := mapMetricItem(msgsMetrics.InOnLedgerRequest.Messages, msgsMetrics.InOnLedgerRequest.Timestamp, msgsMetrics.InOnLedgerRequest.LastMessage.Id)
	publisherStateTransaction := mapMetricItem(msgsMetrics.OutPublisherStateTransaction.Messages, msgsMetrics.OutPublisherStateTransaction.Timestamp, msgsMetrics.OutPublisherStateTransaction.LastMessage.TxDigest)

	table := [][]string{
		makeMessagesMetricsTableRow("State anchor", true, inAnchor),
		makeMessagesMetricsTableRow("On ledger request", true, inOnLedgerRequest),
		makeMessagesMetricsTableRow("Publish state transaction", false, publisherStateTransaction),
	}
	log.PrintTable(header, table)
}

func makeMessagesMetricsTableRow(name string, isIn bool, ncmm *apiclient.InterfaceMetricItem) []string {
	res := make([]string, 5)
	res[0] = name
	if isIn {
		res[1] = "IN"
	} else {
		res[1] = "OUT"
	}
	res[2] = fmt.Sprintf("%v", ncmm.Messages)
	res[3] = ncmm.Timestamp.String()
	res[4] = ncmm.LastMessage
	if len(res[4]) > maxMessageLen {
		res[4] = res[4][:maxMessageLen]
	}
	res[4] = strings.ReplaceAll(res[4], "\n", " ")
	return res
}
