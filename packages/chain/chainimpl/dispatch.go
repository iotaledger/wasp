package chainimpl

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/plugins/peering"
)

func (c *committeeObj) dispatchMessage(msg interface{}) {
	if !c.isOpenQueue.Load() {
		return
	}

	switch msgt := msg.(type) {

	case *peering.PeerMessage:
		// receive a message from peer
		c.processPeerMessage(msgt)

	case *chain.StateUpdateMsg:
		// StateUpdateMsg may come from peer and from own consensus operator
		c.stateMgr.EventStateUpdateMsg(msgt)

	case *chain.StateTransitionMsg:
		if c.operator != nil {
			c.operator.EventStateTransitionMsg(msgt)
		}

	case chain.PendingBatchMsg:
		c.stateMgr.EventPendingBatchMsg(msgt)

	case *chain.StateTransactionMsg:
		// receive state transaction message
		c.stateMgr.EventStateTransactionMsg(msgt)

	case *chain.TransactionInclusionLevelMsg:
		if c.operator != nil {
			c.operator.EventTransactionInclusionLevelMsg(msgt)
		}

	case *chain.RequestMsg:
		// receive request message
		if c.operator != nil {
			c.operator.EventRequestMsg(msgt)
		}

	case chain.BalancesMsg:
		if c.operator != nil {
			c.operator.EventBalancesMsg(msgt)
		}

	case *vm.VMTask:
		// VM finished working
		if c.operator != nil {
			c.operator.EventResultCalculated(msgt)
		}
	case chain.TimerTick:

		if msgt%2 == 0 {
			if c.stateMgr != nil {
				c.stateMgr.EventTimerMsg(msgt / 2)
			}
		} else {
			if c.operator != nil {
				c.operator.EventTimerMsg(msgt / 2)
			}
		}
	}
}

func (c *committeeObj) processPeerMessage(msg *peering.PeerMessage) {

	rdr := bytes.NewReader(msg.MsgData)

	switch msg.MsgType {

	case chain.MsgStateIndexPingPong:
		msgt := &chain.StateIndexPingPongMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		msgt.SenderIndex = msg.SenderIndex

		c.stateMgr.EvidenceStateIndex(msgt.StateIndex)
		c.stateMgr.EventStateIndexPingPongMsg(msgt)

	case chain.MsgNotifyRequests:
		msgt := &chain.NotifyReqMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		c.stateMgr.EvidenceStateIndex(msgt.StateIndex)

		msgt.SenderIndex = msg.SenderIndex

		if c.operator != nil {
			c.operator.EventNotifyReqMsg(msgt)
		}

	case chain.MsgNotifyFinalResultPosted:
		msgt := &chain.NotifyFinalResultPostedMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		c.stateMgr.EvidenceStateIndex(msgt.StateIndex)

		msgt.SenderIndex = msg.SenderIndex

		if c.operator != nil {
			c.operator.EventNotifyFinalResultPostedMsg(msgt)
		}

	case chain.MsgStartProcessingRequest:
		msgt := &chain.StartProcessingBatchMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		c.stateMgr.EvidenceStateIndex(msgt.StateIndex)

		msgt.SenderIndex = msg.SenderIndex
		msgt.Timestamp = msg.Timestamp

		if c.operator != nil {
			c.operator.EventStartProcessingBatchMsg(msgt)
		}

	case chain.MsgSignedHash:
		msgt := &chain.SignedHashMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		c.stateMgr.EvidenceStateIndex(msgt.StateIndex)

		msgt.SenderIndex = msg.SenderIndex
		msgt.Timestamp = msg.Timestamp

		if c.operator != nil {
			c.operator.EventSignedHashMsg(msgt)
		}

	case chain.MsgGetBatch:
		msgt := &chain.GetBatchMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}

		msgt.SenderIndex = msg.SenderIndex

		c.stateMgr.EventGetBatchMsg(msgt)

	case chain.MsgBatchHeader:
		msgt := &chain.BatchHeaderMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		c.stateMgr.EvidenceStateIndex(msgt.StateIndex)

		msgt.SenderIndex = msg.SenderIndex
		c.stateMgr.EventBatchHeaderMsg(msgt)

	case chain.MsgStateUpdate:
		msgt := &chain.StateUpdateMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		c.stateMgr.EvidenceStateIndex(msgt.StateIndex)

		msgt.SenderIndex = msg.SenderIndex
		c.stateMgr.EventStateUpdateMsg(msgt)

	case chain.MsgTestTrace:
		msgt := &chain.TestTraceMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}

		msgt.SenderIndex = msg.SenderIndex
		c.testTrace(msgt)

	default:
		c.log.Errorf("processPeerMessage: wrong msg type")
	}
}
