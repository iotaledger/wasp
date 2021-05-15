package consensus1imp

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
)

func (c *consensusImpl) EventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	c.eventStateTransitionMsgCh <- msg
}
func (c *consensusImpl) eventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	c.log.Debugf("eventStateTransitionMsg: state index: %d, state output: %s, timestamp: %v",
		msg.State.BlockIndex(), coretypes.OID(msg.StateOutput.ID()), msg.StateTimestamp)
	c.setNewState(msg)
	c.takeAction()
}

func (c *consensusImpl) EventResultCalculated(msg *chain.VMResultMsg) {
	c.eventResultCalculatedMsgCh <- msg
}
func (c *consensusImpl) eventResultCalculated(msg *chain.VMResultMsg) {
	c.log.Debugf("eventResultCalculated: block index: %d", msg.Task.VirtualState.BlockIndex())

	if msg.Task.ChainInput.ID() != c.stateOutput.ID() {
		c.log.Warnf("eventResultCalculated: VMResultMsg out of context")
		return
	}
	c.processVMResult(msg.Task)
	c.takeAction()
}

func (c *consensusImpl) EventSignedResultMsg(msg *chain.SignedResultMsg) {
	c.eventSignedResultMsgCh <- msg
}
func (c *consensusImpl) eventSignedResult(msg *chain.SignedResultMsg) {
	c.log.Debugf("eventSignedResult: from sender: %d", msg.SenderIndex)
	c.processSignedResult(msg)
	c.takeAction()
}

func (c *consensusImpl) EventInclusionsStateMsg(msg *chain.InclusionStateMsg) {
	c.eventInclusionStateMsgCh <- msg
}
func (c *consensusImpl) eventInclusionState(msg *chain.InclusionStateMsg) {
	c.log.Debugf("eventInclusionState:  %s: '%s'", msg.TxID.Base58(), msg.State.String())
	c.processInclusionState(msg)

	c.takeAction()
}

func (c *consensusImpl) EventAsynchronousCommonSubsetMsg(msg *chain.AsynchronousCommonSubsetMsg) {
	c.eventACSMsgCh <- msg
}
func (c *consensusImpl) eventAsynchronousCommonSubset(msg *chain.AsynchronousCommonSubsetMsg) {
	c.log.Debugf("eventAsynchronousCommonSubset: len = %d", len(msg.ProposedBatchesBin))
	c.receiveACS(msg.ProposedBatchesBin, msg.SessionID)

	c.takeAction()
}

func (c *consensusImpl) EventVMResultMsg(msg *chain.VMResultMsg) {
	c.eventVMResultMsgCh <- msg
}
func (c *consensusImpl) eventVMResultMsg(msg *chain.VMResultMsg) {
	essenceHash := hashing.HashData(msg.Task.ResultTransactionEssence.Bytes())
	c.log.Debugf("eventVMResultMsg: state index: %d state hash: %s essence hash: %s",
		msg.Task.VirtualState.BlockIndex(), msg.Task.VirtualState.Hash(), essenceHash)
	c.processVMResult(msg.Task)

	c.takeAction()
}

func (c *consensusImpl) EventTimerMsg(msg chain.TimerTick) {
	c.eventTimerMsgCh <- msg
}
func (c *consensusImpl) eventTimerMsg(msg chain.TimerTick) {
	c.lastTimerTick.Store(int64(msg))
	c.refreshConsensusInfo()
	if msg%40 == 0 {
		if snap := c.GetStatusSnapshot(); snap != nil {
			c.log.Infof("timer tick #%d state index: %d mempool = (%d, %d)",
				snap.TimerTick, snap.StateIndex, snap.Mempool.InCounter, snap.Mempool.OutCounter)
		}
	}
	c.takeAction()
}
