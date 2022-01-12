package metrics

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var timestampNeverConst = time.Time{}

var consensusMetricsCmd = &cobra.Command{
	Use:   "consensus",
	Short: "Show current value of collected metrics of consensus",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := config.WaspClient()
		chid, err := iscp.ChainIDFromBase58(chainIDStr)
		log.Check(err)
		workflowStatus, err := client.GetChainConsensusWorkflowStatus(chid)
		log.Check(err)
		header := []string{"Flag name", "Value", "Last time set"}
		table := make([][]string, 9)
		table[0] = makeWorkflowTableRow("State received", workflowStatus.FlagStateReceived, time.Time{})
		table[1] = makeWorkflowTableRow("Batch proposal sent", workflowStatus.FlagBatchProposalSent, workflowStatus.TimeBatchProposalSent)
		table[2] = makeWorkflowTableRow("Consensus on batch reached", workflowStatus.FlagConsensusBatchKnown, workflowStatus.TimeConsensusBatchKnown)
		table[3] = makeWorkflowTableRow("Virtual machine started", workflowStatus.FlagVMStarted, workflowStatus.TimeVMStarted)
		table[4] = makeWorkflowTableRow("Virtual machine result signed", workflowStatus.FlagVMResultSigned, workflowStatus.TimeVMResultSigned)
		table[5] = makeWorkflowTableRow("Transaction finalized", workflowStatus.FlagTransactionFinalized, workflowStatus.TimeTransactionFinalized)
		table[6] = makeWorkflowTableRow("Transaction posted to L1", workflowStatus.FlagTransactionPosted, workflowStatus.TimeTransactionPosted) // TODO: is not meaningful, if I am not a contributor
		table[7] = makeWorkflowTableRow("Transaction seen by L1", workflowStatus.FlagTransactionSeen, workflowStatus.TimeTransactionSeen)
		table[8] = makeWorkflowTableRow("Consensus is completed", !(workflowStatus.FlagInProgress), workflowStatus.TimeCompleted)
		log.PrintTable(header, table)
	},
}

func makeWorkflowTableRow(name string, value bool, timestamp time.Time) []string {
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
