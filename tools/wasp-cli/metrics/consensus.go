package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

var timestampNeverConst = time.Time{}

func initConsensusMetricsCmd() *cobra.Command {
	var node string
	cmd := &cobra.Command{
		Use:   "consensus",
		Short: "Show current value of collected metrics of consensus",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)
			chainAddress, err := cryptolib.NewAddressFromHexString(chainAlias)
			log.Check(err)

			chainID := isc.ChainIDFromAddress(chainAddress)
			workflowStatus, _, err := client.MetricsAPI.
				GetChainWorkflowMetrics(ctx, chainID.String()).
				Execute()
			log.Check(err)

			pipeMetrics, _, err := client.MetricsAPI.
				GetChainPipeMetrics(ctx, chainID.String()).
				Execute()
			log.Check(err)

			header := []string{"Flag name", "Value", "Last time set"}
			table := make([][]string, 15)
			table[0] = makeWorkflowTableRow("State received", workflowStatus.FlagStateReceived, time.Time{})
			table[1] = makeWorkflowTableRow("Batch proposal sent", workflowStatus.FlagBatchProposalSent, workflowStatus.TimeBatchProposalSent)
			table[2] = makeWorkflowTableRow("Consensus on batch reached", workflowStatus.FlagConsensusBatchKnown, workflowStatus.TimeConsensusBatchKnown)
			table[3] = makeWorkflowTableRow("Virtual machine started", workflowStatus.FlagVMStarted, workflowStatus.TimeVMStarted)
			table[4] = makeWorkflowTableRow("Virtual machine result signed", workflowStatus.FlagVMResultSigned, workflowStatus.TimeVMResultSigned)
			table[5] = makeWorkflowTableRow("Transaction finalized", workflowStatus.FlagTransactionFinalized, workflowStatus.TimeTransactionFinalized)
			table[6] = makeWorkflowTableRow("Transaction posted to L1", workflowStatus.FlagTransactionPosted, workflowStatus.TimeTransactionPosted)
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
	waspcmd.WithWaspNodeFlag(cmd, &node)
	return cmd
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
