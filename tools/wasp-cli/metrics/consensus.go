package metrics

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var timestampNeverConst = time.Time{}

var consensusMetricsCmd = &cobra.Command{
	Use:   "consensus",
	Short: "Show current value of collected metrics of consensus",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := config.WaspClient()
		_, chainAddress, err := iotago.ParseBech32(chainAlias)
		log.Check(err)
		chid := isc.ChainIDFromAddress(chainAddress.(*iotago.AliasAddress))
		workflowStatus, err := client.GetChainConsensusWorkflowStatus(&chid)
		log.Check(err)
		pipeMetrics, err := client.GetChainConsensusPipeMetrics(&chid)
		log.Check(err)
		header := []string{"Flag name", "Value", "Last time set"}
		table := make([][]string, 15)
		table[0] = makeWorkflowTableRow("State received", workflowStatus.FlagStateReceived, time.Time{})
		table[1] = makeWorkflowTableRow("Batch proposal sent", workflowStatus.FlagBatchProposalSent, workflowStatus.TimeBatchProposalSent)
		table[2] = makeWorkflowTableRow("Consensus on batch reached", workflowStatus.FlagConsensusBatchKnown, workflowStatus.TimeConsensusBatchKnown)
		table[3] = makeWorkflowTableRow("Virtual machine started", workflowStatus.FlagVMStarted, workflowStatus.TimeVMStarted)
		table[4] = makeWorkflowTableRow("Virtual machine result signed", workflowStatus.FlagVMResultSigned, workflowStatus.TimeVMResultSigned)
		table[5] = makeWorkflowTableRow("Transaction finalized", workflowStatus.FlagTransactionFinalized, workflowStatus.TimeTransactionFinalized)
		table[6] = makeWorkflowTableRow("Transaction posted to L1", workflowStatus.FlagTransactionPosted, workflowStatus.TimeTransactionPosted) // TODO: is not meaningful, if I am not a contributor
		table[7] = makeWorkflowTableRow("Transaction seen by L1", workflowStatus.FlagTransactionSeen, workflowStatus.TimeTransactionSeen)
		table[8] = makeWorkflowTableRow("Consensus is completed", !(workflowStatus.FlagInProgress), workflowStatus.TimeCompleted)
		table[9] = makeWorkflowTableRow("Current state index", workflowStatus.CurrentStateIndex, time.Time{})
		table[10] = makeWorkflowTableRow("Event state transition message pipe size", pipeMetrics.EventStateTransitionMsgPipeSize, time.Time{})
		table[12] = makeWorkflowTableRow("Event ACS message pipe size", pipeMetrics.EventACSMsgPipeSize, time.Time{})
		table[13] = makeWorkflowTableRow("Event VM result message pipe size", pipeMetrics.EventVMResultMsgPipeSize, time.Time{})
		table[14] = makeWorkflowTableRow("Event timer message pipe size", pipeMetrics.EventTimerMsgPipeSize, time.Time{})
		log.PrintTable(header, table)
	},
}

func makeWorkflowTableRow(name string, value interface{}, timestamp time.Time) []string {
	res := make([]string, 3)
	res[0] = name
	res[1] = fmt.Sprintf("%v", value)
	if timestamp == timestampNeverConst {
		res[2] = ""
	} else {
		res[2] = timestamp.String()
	}
	return res
}
