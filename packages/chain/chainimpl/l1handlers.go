package chainimpl

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
)

type txInclusionStateMsg struct {
	txID  iotago.TransactionID
	state string
}

func (c *chainObj) handleStateInclusion(txID iotago.TransactionID, inclusionState string) {
	if c.consensus != nil {
		c.nodeConn.GetMetrics().GetInTxInclusionState().CountLastMessage(&txInclusionStateMsg{
			txID:  txID,
			state: inclusionState,
		})
		c.consensus.EnqueueTxInclusionsStateMsg(txID, inclusionState)
	}
}

func (c *chainObj) handleMilestone(milestonePointer *nodeclient.MilestoneInfo) {
	if c.consensus != nil {
		c.consensus.SetTimeData(time.Unix(int64(milestonePointer.Timestamp), 0))
	}
}

func (c *chainObj) stateOutputHandler(outputID iotago.OutputID, output iotago.Output) {
	c.nodeConn.GetMetrics().GetInStateOutput().CountLastMessage(struct {
		OutputID iotago.OutputID
		Output   iotago.Output
	}{
		OutputID: outputID,
		Output:   output,
	})
	outputIDUTXO := outputID.UTXOInput()
	outputIDstring := isc.OID(outputIDUTXO)
	c.log.Debugf("handling state output ID %v", outputIDstring)
	aliasOutput, ok := output.(*iotago.AliasOutput)
	if !ok {
		c.log.Panicf("unexpected output ID %v type %T received as state update to chain ID %s; alias output expected",
			outputIDstring, output, c.chainID)
	}
	if aliasOutput.AliasID.Empty() && aliasOutput.StateIndex != 0 {
		c.log.Panicf("unexpected output ID %v index %v with empty alias ID received as state update to chain ID %s; alias ID may be empty for initial alias output only",
			outputIDstring, aliasOutput.StateIndex, c.chainID)
	}
	if !util.AliasIDFromAliasOutput(aliasOutput, outputID).ToAddress().Equal(c.chainID.AsAddress()) {
		c.log.Panicf("unexpected output ID %v address %s index %v received as state update to chain ID %s, address %s",
			outputIDstring, aliasOutput.AliasID.ToAddress(), aliasOutput.StateIndex, c.chainID, c.chainID.AsAddress())
	}
	c.log.Debugf("handling state output ID %v: writing alias output to channel", outputIDstring)
	c.nodeConn.GetMetrics().GetInAliasOutput().CountLastMessage(aliasOutput)
	c.EnqueueAliasOutput(isc.NewAliasOutputWithID(aliasOutput, outputIDUTXO))
	c.log.Debugf("handling state output ID %v: alias output handled", outputIDstring)
}

func (c *chainObj) outputHandler(outputID iotago.OutputID, output iotago.Output) {
	c.nodeConn.GetMetrics().GetInOutput().CountLastMessage(struct {
		OutputID iotago.OutputID
		Output   iotago.Output
	}{
		OutputID: outputID,
		Output:   output,
	})
	outputIDUTXO := outputID.UTXOInput()
	outputIDstring := isc.OID(outputIDUTXO)
	c.log.Debugf("handling output ID %v", outputIDstring)
	onLedgerRequest, err := isc.OnLedgerFromUTXO(output, outputIDUTXO)
	if err != nil {
		c.log.Warnf("handling output ID %v: unknown output type; ignoring it", outputIDstring)
		return
	}
	c.log.Debugf("handling output ID %v: writing on ledger request to channel", outputIDstring)
	c.nodeConn.GetMetrics().GetInOnLedgerRequest().CountLastMessage(onLedgerRequest)
	c.mempool.ReceiveRequest(onLedgerRequest)
	c.log.Debugf("handling output ID %v: on ledger request handled", outputIDstring)
}

func (c *chainObj) PublishStateTransaction(stateIndex uint32, tx *iotago.Transaction) error {
	c.nodeConn.GetMetrics().GetOutPublishStateTransaction().CountLastMessage(struct {
		StateIndex  uint32
		Transaction *iotago.Transaction
	}{
		StateIndex:  stateIndex,
		Transaction: tx,
	})
	return c.nodeConn.PublishTransaction(c.chainID, tx)
}

func (c *chainObj) PublishGovernanceTransaction(tx *iotago.Transaction) error {
	c.nodeConn.GetMetrics().GetOutPublishGovernanceTransaction().CountLastMessage(tx)
	return c.nodeConn.PublishTransaction(c.chainID, tx)
}

func (c *chainObj) PullLatestOutput() {
	c.nodeConn.GetMetrics().GetOutPullLatestOutput().CountLastMessage(nil)
	c.nodeConn.PullLatestOutput(c.chainID)
}

// func (c *chainObj) PullTxInclusionState(txID iotago.TransactionID) {
// 	c.nodeConn.GetMetrics().GetOutPullTxInclusionState().CountLastMessage(txID)
// 	c.nodeConn.PullTxInclusionState(c.chainID, txID)
// }

func (c *chainObj) PullStateOutputByID(outputID *iotago.UTXOInput) {
	c.nodeConn.GetMetrics().GetOutPullOutputByID().CountLastMessage(outputID)
	c.nodeConn.PullStateOutputByID(c.chainID, outputID)
}
