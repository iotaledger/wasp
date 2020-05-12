package commiteeimpl

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/plugins/peering"
	"time"
)

func (c *committeeObj) dispatchMessage(msg interface{}) {
	if !c.isOperational.Load() {
		return
	}
	stateMgr := false

	switch msgt := msg.(type) {

	case *peering.PeerMessage:
		// receive a message from peer
		c.processPeerMessage(msgt)

	case *committee.StateUpdateMsg:
		// StateUpdateMsg may come from peer and from own consensus operator
		c.stateMgr.EventStateUpdateMsg(msgt)

	case *committee.StateTransitionMsg:
		c.operator.EventStateTransitionMsg(msgt)

	case committee.StateTransactionMsg:
		// receive state transaction message
		c.stateMgr.EventStateTransactionMsg(msgt)

	case committee.BalancesMsg:
		// outputs and balances of the address arrived
		c.operator.EventBalancesMsg(msgt)

	case *committee.RequestMsg:
		// receive request message
		c.operator.EventRequestMsg(msgt)

	case *vm.VMOutput:
		// VM finished working
		c.operator.EventResultCalculated(msgt)

	case committee.TimerTick:
		if stateMgr {
			c.stateMgr.EventTimerMsg(msgt)
		} else {
			c.operator.EventTimerMsg(msgt)
		}
		stateMgr = !stateMgr
	}
}

func (c *committeeObj) processPeerMessage(msg *peering.PeerMessage) {
	rdr := bytes.NewReader(msg.MsgData)

	switch msg.MsgType {

	case committee.MsgNotifyRequests:
		msgt := &committee.NotifyReqMsg{}
		if err := msgt.Read(rdr); err != nil {
			log.Error(err)
			return
		}
		msgt.SenderIndex = msg.SenderIndex

		if c.stateMgr.CheckSynchronizationStatus(msgt.StateIndex) {
			c.operator.EventNotifyReqMsg(msgt)
		}

	case committee.MsgStartProcessingRequest:
		msgt := &committee.StartProcessingReqMsg{}
		if err := msgt.Read(rdr); err != nil {
			log.Error(err)
			return
		}
		msgt.SenderIndex = msg.SenderIndex
		msgt.Timestamp = time.Unix(0, msg.Timestamp)

		if c.stateMgr.CheckSynchronizationStatus(msgt.StateIndex) {
			c.operator.EventStartProcessingReqMsg(msgt)
		}

	case committee.MsgSignedHash:
		msgt := &committee.SignedHashMsg{}
		if err := msgt.Read(rdr); err != nil {
			log.Error(err)
			return
		}
		msgt.SenderIndex = msg.SenderIndex
		msgt.Timestamp = time.Unix(0, msg.Timestamp)

		if c.stateMgr.CheckSynchronizationStatus(msgt.StateIndex) {
			c.operator.EventSignedHashMsg(msgt)
		}

	case committee.MsgGetStateUpdate:
		msgt := &committee.GetStateUpdateMsg{}
		if err := msgt.Read(rdr); err != nil {
			log.Error(err)
			return
		}
		msgt.SenderIndex = msg.SenderIndex

		if c.stateMgr.CheckSynchronizationStatus(msgt.StateIndex) {
			c.stateMgr.EventGetStateUpdateMsg(msgt)
		}

	case committee.MsgStateUpdate:
		msgt := &committee.StateUpdateMsg{}
		if err := msgt.Read(rdr); err != nil {
			log.Error(err)
			return
		}
		msgt.SenderIndex = msg.SenderIndex

		if c.stateMgr.CheckSynchronizationStatus(msgt.StateIndex) {
			c.stateMgr.EventStateUpdateMsg(msgt)
		}

	default:
		log.Errorf("processPeerMessage: wrong msg type")
	}
}
