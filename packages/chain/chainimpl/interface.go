// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/transaction"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

func (c *chainObj) ID() *coretypes.ChainID {
	return &c.chainID
}

func (c *chainObj) GetCommitteeInfo() *chain.CommitteeInfo {
	if c.committee == nil {
		return nil
	}
	return &chain.CommitteeInfo{
		Address:       c.committee.DKShare().Address,
		Size:          c.committee.Size(),
		Quorum:        c.committee.Quorum(),
		QuorumIsAlive: c.committee.QuorumIsAlive(),
		PeerStatus:    c.committee.PeerStatus(),
	}
}

func (c *chainObj) startTimer() {
	go func() {
		c.stateMgr.Ready().MustWait()
		tick := 0
		for !c.IsDismissed() {
			time.Sleep(chain.TimerTickPeriod)
			c.ReceiveMessage(chain.TimerTick(tick))
			tick++
		}
	}()
}

func (c *chainObj) Dismiss(reason string) {
	c.log.Infof("Dismiss chain. Reason: '%s'", reason)

	c.dismissOnce.Do(func() {
		c.dismissed.Store(true)

		close(c.chMsg)

		c.mempool.Close()
		c.stateMgr.Close()
		if c.committee != nil {
			c.committee.Close()
		}
		if c.consensus != nil {
			c.consensus.Close()
		}
		c.eventRequestProcessed.DetachAll()
		c.eventStateTransition.DetachAll()
		c.eventSynced.DetachAll()
	})

	publisher.Publish("dismissed_chain", c.chainID.Base58())
}

func (c *chainObj) IsDismissed() bool {
	return c.dismissed.Load()
}

func (c *chainObj) ReceiveMessage(msg interface{}) {
	if !c.IsDismissed() {
		select {
		case c.chMsg <- msg:
		default:
			c.log.Warnf("ReceiveMessage with type '%T' failed. Retrying after %s", msg, chain.ReceiveMsgChannelRetryDelay)
			go func() {
				time.Sleep(chain.ReceiveMsgChannelRetryDelay)
				c.ReceiveMessage(msg)
			}()
		}
	}
}

func (c *chainObj) ReceiveTransaction(tx *ledgerstate.Transaction) {
	c.log.Debugf("ReceiveTransaction: %s", tx.ID().Base58())
	reqs, err := request.RequestsOnLedgerFromTransaction(tx, c.chainID.AsAddress())
	if err != nil {
		c.log.Warnf("failed to parse transaction %s: %v", tx.ID().Base58(), err)
		return
	}
	for _, req := range reqs {
		c.ReceiveRequest(req)
	}
	if chainOut := transaction.GetAliasOutput(tx, c.chainID.AsAddress()); chainOut != nil {
		c.ReceiveState(chainOut, tx.Essence().Timestamp())
	}
}

func (c *chainObj) ReceiveRequest(req coretypes.Request) {
	c.log.Debugf("ReceiveRequest: %s", req.ID())
	c.mempool.ReceiveRequest(req)
}

func (c *chainObj) ReceiveState(stateOutput *ledgerstate.AliasOutput, timestamp time.Time) {
	fmt.Printf("++++++++++++ receive state %s\n", stateOutput.Address().Base58())

	c.log.Debugf("ReceiveState #%d: outputID: %s, stateAddr: %s",
		stateOutput.GetStateIndex(), coretypes.OID(stateOutput.ID()), stateOutput.GetStateAddress().Base58())
	c.ReceiveMessage(&chain.StateMsg{
		ChainOutput: stateOutput,
		Timestamp:   timestamp,
	})
}

func (c *chainObj) ReceiveInclusionState(txID ledgerstate.TransactionID, inclusionState ledgerstate.InclusionState) {
	c.ReceiveMessage(&chain.InclusionStateMsg{
		TxID:  txID,
		State: inclusionState,
	}) // TODO special entry point
}

func (c *chainObj) ReceiveOutput(output ledgerstate.Output) {
	c.stateMgr.EventOutputMsg(output)
}

func (c *chainObj) BlobCache() coretypes.BlobCache {
	return c.blobProvider
}

func (c *chainObj) GetRequestProcessingStatus(reqID coretypes.RequestID) chain.RequestProcessingStatus {
	if c.IsDismissed() {
		return chain.RequestProcessingStatusUnknown
	}
	if c.consensus != nil {
		if c.mempool.HasRequest(reqID) {
			return chain.RequestProcessingStatusBacklog
		}
	}
	stateReader, err := state.NewStateReader(c.dbProvider, &c.chainID)
	if err != nil {
		c.log.Errorf("GetRequestProcessingStatus: %v", err)
		return chain.RequestProcessingStatusUnknown
	}
	if !blocklog.IsRequestProcessed(stateReader, &reqID) {
		return chain.RequestProcessingStatusUnknown
	}
	return chain.RequestProcessingStatusCompleted
}

func (c *chainObj) Processors() *processors.ProcessorCache {
	return c.procset
}

func (c *chainObj) EventRequestProcessed() *events.Event {
	return c.eventRequestProcessed
}

func (c *chainObj) RequestProcessed() *events.Event {
	return c.eventRequestProcessed
}

func (c *chainObj) StateTransition() *events.Event {
	return c.eventStateTransition
}

func (c *chainObj) StateSynced() *events.Event {
	return c.eventSynced
}

func (c *chainObj) Events() chain.ChainEvents {
	return c
}

func (c *chainObj) Mempool() chain.Mempool {
	return c.mempool
}
