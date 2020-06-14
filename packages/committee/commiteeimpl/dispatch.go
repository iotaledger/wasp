package commiteeimpl

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/committee"
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

	case *committee.StateUpdateMsg:
		// StateUpdateMsg may come from peer and from own consensus operator
		c.stateMgr.EventStateUpdateMsg(msgt)

	case *committee.StateTransitionMsg:
		if c.operator != nil {
			c.operator.EventStateTransitionMsg(msgt)
		}

	case committee.PendingBatchMsg:
		c.stateMgr.EventPendingBatchMsg(msgt)

	case committee.ProcessorIsReady:
		if c.operator != nil {
			c.operator.EventProcessorReady(msgt)
		}

	case committee.StateTransactionMsg:
		// receive state transaction message
		c.stateMgr.EventStateTransactionMsg(msgt)

	case committee.RequestMsg:
		// receive request message
		if c.operator != nil {
			c.operator.EventRequestMsg(msgt)
		}

	case committee.BalancesMsg:
		if c.operator != nil {
			c.operator.EventBalancesMsg(msgt)
		}

	case *vm.VMTask:
		// VM finished working
		if c.operator != nil {
			c.operator.EventResultCalculated(msgt)
		}
	case committee.TimerTick:

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

	case committee.MsgNotifyRequests:
		msgt := &committee.NotifyReqMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		c.stateMgr.CheckSynchronizationStatus(msgt.StateIndex)

		msgt.SenderIndex = msg.SenderIndex

		if c.operator != nil {
			c.operator.EventNotifyReqMsg(msgt)
		}

	case committee.MsgStartProcessingRequest:
		msgt := &committee.StartProcessingBatchMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		c.stateMgr.CheckSynchronizationStatus(msgt.StateIndex)

		msgt.SenderIndex = msg.SenderIndex
		msgt.Timestamp = msg.Timestamp

		if c.operator != nil {
			c.operator.EventStartProcessingBatchMsg(msgt)
		}

	case committee.MsgSignedHash:
		msgt := &committee.SignedHashMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		c.stateMgr.CheckSynchronizationStatus(msgt.StateIndex)

		msgt.SenderIndex = msg.SenderIndex
		msgt.Timestamp = msg.Timestamp

		if c.operator != nil {
			c.operator.EventSignedHashMsg(msgt)
		}

	case committee.MsgGetBatch:
		msgt := &committee.GetBatchMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		c.stateMgr.CheckSynchronizationStatus(msgt.StateIndex)

		msgt.SenderIndex = msg.SenderIndex

		c.stateMgr.EventGetBatchMsg(msgt)

	case committee.MsgBatchHeader:
		msgt := &committee.BatchHeaderMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		c.stateMgr.CheckSynchronizationStatus(msgt.StateIndex)

		msgt.SenderIndex = msg.SenderIndex
		c.stateMgr.EventBatchHeaderMsg(msgt)

	case committee.MsgStateUpdate:
		msgt := &committee.StateUpdateMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		c.stateMgr.CheckSynchronizationStatus(msgt.StateIndex)

		msgt.SenderIndex = msg.SenderIndex
		c.stateMgr.EventStateUpdateMsg(msgt)

	case committee.MsgTestTrace:
		msgt := &committee.TestTraceMsg{}
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
