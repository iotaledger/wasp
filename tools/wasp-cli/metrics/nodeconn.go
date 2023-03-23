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
			client := cliclients.WaspClient(node)

			if chainAlias == "" {
				msgsMetrics, _, err := client.MetricsApi.GetNodeMessageMetrics(context.Background()).Execute()
				log.Check(err)
				printNodeMessagesMetrics(msgsMetrics)
			} else {
				chainID, err := isc.ChainIDFromString(chainAlias)
				log.Check(err)
				msgsMetrics, _, err := client.MetricsApi.GetChainMessageMetrics(context.Background(), chainID.String()).Execute()
				log.Check(err)
				printChainMessagesMetrics(msgsMetrics)
			}
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

func printNodeMessagesMetrics(msgsMetrics *apiclient.NodeMessageMetrics) {
	log.Printf("Following chains are registered for L1 events:\n")
	for _, s := range msgsMetrics.RegisteredChainIDs {
		log.Printf("\t%s\n", s)
	}

	milestoneMsg := ""
	if msgsMetrics.InMilestone.LastMessage.MilestoneId != nil {
		milestoneMsg = *msgsMetrics.InMilestone.LastMessage.MilestoneId
	}

	header := []string{"Message name", "", "Total", "Last time", "Last message"}

	inMilestone := mapMetricItem(msgsMetrics.InMilestone.Messages, msgsMetrics.InMilestone.Timestamp, milestoneMsg)
	inStateOutput := mapMetricItem(msgsMetrics.InStateOutput.Messages, msgsMetrics.InStateOutput.Timestamp, msgsMetrics.InStateOutput.LastMessage.OutputId)
	inAliasOutput := mapMetricItem(msgsMetrics.InAliasOutput.Messages, msgsMetrics.InAliasOutput.Timestamp, msgsMetrics.InAliasOutput.LastMessage.Raw)
	inOutput := mapMetricItem(msgsMetrics.InOutput.Messages, msgsMetrics.InOutput.Timestamp, msgsMetrics.InOutput.LastMessage.OutputId)
	inOnLedgerRequest := mapMetricItem(msgsMetrics.InOnLedgerRequest.Messages, msgsMetrics.InOnLedgerRequest.Timestamp, msgsMetrics.InOnLedgerRequest.LastMessage.OutputId)
	inTxInclusionState := mapMetricItem(msgsMetrics.InTxInclusionState.Messages, msgsMetrics.InTxInclusionState.Timestamp, msgsMetrics.InTxInclusionState.LastMessage.TxId)
	publisherStateTransaction := mapMetricItem(msgsMetrics.OutPublisherStateTransaction.Messages, msgsMetrics.OutPublisherStateTransaction.Timestamp, msgsMetrics.OutPublisherStateTransaction.LastMessage.TxId)
	govTransaction := mapMetricItem(msgsMetrics.OutPublishGovernanceTransaction.Messages, msgsMetrics.OutPublishGovernanceTransaction.Timestamp, msgsMetrics.OutPublishGovernanceTransaction.LastMessage.TxId)
	pullLatestOutput := mapMetricItem(msgsMetrics.OutPullLatestOutput.Messages, msgsMetrics.OutPullLatestOutput.Timestamp, msgsMetrics.OutPullLatestOutput.LastMessage)
	outPullTxInclusionState := mapMetricItem(msgsMetrics.OutPullTxInclusionState.Messages, msgsMetrics.OutPullTxInclusionState.Timestamp, msgsMetrics.OutPullTxInclusionState.LastMessage.TxId)
	outPullOutputByID := mapMetricItem(msgsMetrics.OutPullOutputByID.Messages, msgsMetrics.OutPullOutputByID.Timestamp, msgsMetrics.OutPullOutputByID.LastMessage.OutputId)

	table := [][]string{
		makeMessagesMetricsTableRow("Milestone", true, inMilestone),
		makeMessagesMetricsTableRow("State output", true, inStateOutput),
		makeMessagesMetricsTableRow("Alias output", true, inAliasOutput),
		makeMessagesMetricsTableRow("Output", true, inOutput),
		makeMessagesMetricsTableRow("On ledger request", true, inOnLedgerRequest),
		makeMessagesMetricsTableRow("Tx inclusion state", true, inTxInclusionState),
		makeMessagesMetricsTableRow("Publish state transaction", false, publisherStateTransaction),
		makeMessagesMetricsTableRow("Publish governance transaction", false, govTransaction),
		makeMessagesMetricsTableRow("Pull latest output", false, pullLatestOutput),
		makeMessagesMetricsTableRow("Pull tx inclusion state", false, outPullTxInclusionState),
		makeMessagesMetricsTableRow("Pull output by ID", false, outPullOutputByID),
	}
	log.PrintTable(header, table)
}

func printChainMessagesMetrics(msgsMetrics *apiclient.ChainMessageMetrics) {
	header := []string{"Message name", "", "Total", "Last time", "Last message"}

	inStateOutput := mapMetricItem(msgsMetrics.InStateOutput.Messages, msgsMetrics.InStateOutput.Timestamp, msgsMetrics.InStateOutput.LastMessage.OutputId)
	inAliasOutput := mapMetricItem(msgsMetrics.InAliasOutput.Messages, msgsMetrics.InAliasOutput.Timestamp, msgsMetrics.InAliasOutput.LastMessage.Raw)
	inOutput := mapMetricItem(msgsMetrics.InOutput.Messages, msgsMetrics.InOutput.Timestamp, msgsMetrics.InOutput.LastMessage.OutputId)
	inOnLedgerRequest := mapMetricItem(msgsMetrics.InOnLedgerRequest.Messages, msgsMetrics.InOnLedgerRequest.Timestamp, msgsMetrics.InOnLedgerRequest.LastMessage.OutputId)
	inTxInclusionState := mapMetricItem(msgsMetrics.InTxInclusionState.Messages, msgsMetrics.InTxInclusionState.Timestamp, msgsMetrics.InTxInclusionState.LastMessage.TxId)
	publisherStateTransaction := mapMetricItem(msgsMetrics.OutPublisherStateTransaction.Messages, msgsMetrics.OutPublisherStateTransaction.Timestamp, msgsMetrics.OutPublisherStateTransaction.LastMessage.TxId)
	govTransaction := mapMetricItem(msgsMetrics.OutPublishGovernanceTransaction.Messages, msgsMetrics.OutPublishGovernanceTransaction.Timestamp, msgsMetrics.OutPublishGovernanceTransaction.LastMessage.TxId)
	pullLatestOutput := mapMetricItem(msgsMetrics.OutPullLatestOutput.Messages, msgsMetrics.OutPullLatestOutput.Timestamp, msgsMetrics.OutPullLatestOutput.LastMessage)
	outPullTxInclusionState := mapMetricItem(msgsMetrics.OutPullTxInclusionState.Messages, msgsMetrics.OutPullTxInclusionState.Timestamp, msgsMetrics.OutPullTxInclusionState.LastMessage.TxId)
	outPullOutputByID := mapMetricItem(msgsMetrics.OutPullOutputByID.Messages, msgsMetrics.OutPullOutputByID.Timestamp, msgsMetrics.OutPullOutputByID.LastMessage.OutputId)

	table := [][]string{
		makeMessagesMetricsTableRow("State output", true, inStateOutput),
		makeMessagesMetricsTableRow("Alias output", true, inAliasOutput),
		makeMessagesMetricsTableRow("Output", true, inOutput),
		makeMessagesMetricsTableRow("On ledger request", true, inOnLedgerRequest),
		makeMessagesMetricsTableRow("Tx inclusion state", true, inTxInclusionState),
		makeMessagesMetricsTableRow("Publish state transaction", false, publisherStateTransaction),
		makeMessagesMetricsTableRow("Publish governance transaction", false, govTransaction),
		makeMessagesMetricsTableRow("Pull latest output", false, pullLatestOutput),
		makeMessagesMetricsTableRow("Pull tx inclusion state", false, outPullTxInclusionState),
		makeMessagesMetricsTableRow("Pull output by ID", false, outPullOutputByID),
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
	res[4] = strings.Replace(res[4], "\n", " ", -1)
	return res
}
