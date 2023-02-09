package metrics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

const maxMessageLen = 80

func initNodeconnMetricsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "nodeconn",
		Short: "Show current value of collected metrics of connection to L1",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			client := cliclients.WaspClientForIndex()

			if chainAlias == "" {

				nodeconnMetrics, _, err := client.MetricsApi.GetL1Metrics(context.Background()).Execute()
				log.Check(err)
				log.Printf("Following chains are registered for L1 events:\n")
				for _, s := range nodeconnMetrics.RegisteredChainIDs {
					log.Printf("\t%s\n", s)
				}

				milestoneMsg := ""
				if nodeconnMetrics.InMilestone.LastMessage.MilestoneId != nil {
					milestoneMsg = *nodeconnMetrics.InMilestone.LastMessage.MilestoneId
				}

				inMilestone := mapMetricItem(nodeconnMetrics.InMilestone.Messages, nodeconnMetrics.InMilestone.Timestamp, milestoneMsg)
				printMessagesMetrics(
					nodeconnMetrics,
					[][]string{makeMessagesMetricsTableRow("Milestone", true, inMilestone)},
				)
			} else {
				chainID, err := isc.ChainIDFromString(chainAlias)
				log.Check(err)
				msgsMetrics, _, err := client.MetricsApi.GetChainMetrics(context.Background(), chainID.String()).Execute()
				log.Check(err)
				printMessagesMetrics(msgsMetrics, [][]string{})
			}
		},
	}
}

func mapMetricItem(messages uint32, timestamp time.Time, message string) *apiclient.InterfaceMetricItem {
	return &apiclient.InterfaceMetricItem{
		Timestamp:   timestamp,
		LastMessage: message,
		Messages:    messages,
	}
}

func printMessagesMetrics(msgsMetrics *apiclient.ChainMetrics, additionalRows [][]string) {
	header := []string{"Message name", "", "Total", "Last time", "Last message"}

	publisherStateTransaction := mapMetricItem(msgsMetrics.OutPublisherStateTransaction.Messages, msgsMetrics.OutPublisherStateTransaction.Timestamp, msgsMetrics.OutPublisherStateTransaction.LastMessage.TxId)
	govTransaction := mapMetricItem(msgsMetrics.OutPublishGovernanceTransaction.Messages, msgsMetrics.OutPublishGovernanceTransaction.Timestamp, msgsMetrics.OutPublishGovernanceTransaction.LastMessage.TxId)
	pullLatestOutput := mapMetricItem(msgsMetrics.OutPullLatestOutput.Messages, msgsMetrics.OutPullLatestOutput.Timestamp, msgsMetrics.OutPullLatestOutput.LastMessage)
	outPullTxInclusionState := mapMetricItem(msgsMetrics.OutPullTxInclusionState.Messages, msgsMetrics.OutPullTxInclusionState.Timestamp, msgsMetrics.OutPullTxInclusionState.LastMessage.TxId)
	outPullOutputByID := mapMetricItem(msgsMetrics.OutPullOutputByID.Messages, msgsMetrics.OutPullOutputByID.Timestamp, msgsMetrics.OutPullOutputByID.LastMessage.OutputId)
	inStateOutput := mapMetricItem(msgsMetrics.InStateOutput.Messages, msgsMetrics.InStateOutput.Timestamp, msgsMetrics.InStateOutput.LastMessage.OutputId)
	inAliasOutput := mapMetricItem(msgsMetrics.InAliasOutput.Messages, msgsMetrics.InAliasOutput.Timestamp, msgsMetrics.InAliasOutput.LastMessage.Raw)
	inOutput := mapMetricItem(msgsMetrics.InOutput.Messages, msgsMetrics.InOutput.Timestamp, msgsMetrics.InOutput.LastMessage.OutputId)
	inOnLedgerRequest := mapMetricItem(msgsMetrics.InOnLedgerRequest.Messages, msgsMetrics.InOnLedgerRequest.Timestamp, msgsMetrics.InOnLedgerRequest.LastMessage.OutputId)
	inTxInclusionState := mapMetricItem(msgsMetrics.InTxInclusionState.Messages, msgsMetrics.InTxInclusionState.Timestamp, msgsMetrics.InTxInclusionState.LastMessage.TxId)

	table := [][]string{
		makeMessagesMetricsTableRow("Publish state transaction", false, publisherStateTransaction),
		makeMessagesMetricsTableRow("Publish governance transaction", false, govTransaction),
		makeMessagesMetricsTableRow("Pull latest output", false, pullLatestOutput),
		makeMessagesMetricsTableRow("Pull tx inclusion state", false, outPullTxInclusionState),
		makeMessagesMetricsTableRow("Pull output by ID", false, outPullOutputByID),
		makeMessagesMetricsTableRow("State output", true, inStateOutput),
		makeMessagesMetricsTableRow("Alias output", true, inAliasOutput),
		makeMessagesMetricsTableRow("Output", true, inOutput),
		makeMessagesMetricsTableRow("On ledger request", true, inOnLedgerRequest),
		makeMessagesMetricsTableRow("Tx inclusion state", true, inTxInclusionState),
	}
	table = append(table, additionalRows...)
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
	res[4] = strings.Replace(res[4], "\n", " ", -1)
	return res
}
