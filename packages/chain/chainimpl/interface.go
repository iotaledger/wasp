// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/sctransaction"

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

func (c *chainObj) Committee() chain.Committee {
	return c.committee
}

func (c *chainObj) IsOpenQueue() bool {
	if c.IsDismissed() {
		return false
	}
	if c.isOpenQueue.Load() {
		return true
	}
	c.mutexIsReady.Lock()
	defer c.mutexIsReady.Unlock()

	return c.checkReady()
}

func (c *chainObj) SetReadyStateManager() {
	if c.IsDismissed() {
		return
	}
	c.mutexIsReady.Lock()
	defer c.mutexIsReady.Unlock()

	c.isReadyStateManager = true
	c.log.Debugf("State Manager object was created")
	c.checkReady()
}

func (c *chainObj) SetReadyConsensus() {
	if c.IsDismissed() {
		return
	}
	c.mutexIsReady.Lock()
	defer c.mutexIsReady.Unlock()

	c.isReadyConsensus = true
	c.log.Debugf("consensus object was created")
	c.checkReady()
}

func (c *chainObj) SetConnectPeriodOver() {
	if c.IsDismissed() {
		return
	}
	c.mutexIsReady.Lock()
	defer c.mutexIsReady.Unlock()

	c.isConnectPeriodOver = true
	c.log.Debugf("connect period is over")
	c.checkReady()
}

func (c *chainObj) SetQuorumOfConnectionsReached() {
	if c.IsDismissed() {
		return
	}
	c.mutexIsReady.Lock()
	defer c.mutexIsReady.Unlock()

	c.isQuorumOfConnectionsReached = true
	c.log.Debugf("quorum of connections has been reached")
	c.checkReady()
}

func (c *chainObj) isReady() bool {
	return c.isReadyConsensus &&
		c.isReadyStateManager &&
		c.isConnectPeriodOver &&
		c.isQuorumOfConnectionsReached
}

func (c *chainObj) checkReady() bool {
	if c.IsDismissed() {
		panic("dismissed")
	}
	if c.isReady() {
		c.isOpenQueue.Store(true)
		c.startTimer()
		c.onActivation()

		c.log.Infof("committee now is fully initialized")
		publisher.Publish("active_committee", c.chainID.Base58())
	}
	return c.isReady()
}

func (c *chainObj) startTimer() {
	go func() {
		tick := 0
		for c.isOpenQueue.Load() {
			time.Sleep(chain.TimerTickPeriod)
			c.ReceiveMessage(chain.TimerTick(tick))
			tick++
		}
	}()
}

func (c *chainObj) Dismiss() {
	c.log.Infof("Dismiss chain %s", c.chainID)

	c.dismissOnce.Do(func() {
		c.isOpenQueue.Store(false)
		c.dismissed.Store(true)

		close(c.chMsg)

		c.committee.Close()
		c.stateMgr.Close()
		c.consensus.Close()
	})

	publisher.Publish("dismissed_chain", c.chainID.Base58())
}

func (c *chainObj) IsDismissed() bool {
	return c.dismissed.Load()
}

func (c *chainObj) ReceiveMessage(msg interface{}) {
	if c.isOpenQueue.Load() {
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
	reqs, err := sctransaction.RequestsOnLedgerFromTransaction(tx, c.chainID.AsAddress())
	if err != nil {
		c.log.Warnf("failed to parse transaction %s: %v", tx.ID().Base58(), err)
		return
	}
	for _, req := range reqs {
		c.ReceiveRequest(req)
	}
	if chainOut := sctransaction.FindAliasOutput(tx.Essence(), c.chainID.AsAddress()); chainOut != nil {
		c.ReceiveState(chainOut, tx.Essence().Timestamp())
	}
}

func (c *chainObj) ReceiveRequest(req coretypes.Request) {
	c.ReceiveMessage(req) //
}

func (c *chainObj) ReceiveState(stateOutput *ledgerstate.AliasOutput, timestamp time.Time) {
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

func (c *chainObj) BlobCache() coretypes.BlobCache {
	return c.blobProvider
}

func (c *chainObj) GetRequestProcessingStatus(reqID coretypes.RequestID) chain.RequestProcessingStatus {
	if c.IsDismissed() {
		return chain.RequestProcessingStatusUnknown
	}
	if c.consensus != nil {
		if c.consensus.IsRequestInBacklog(reqID) {
			return chain.RequestProcessingStatusBacklog
		}
	}
	processed, err := state.IsRequestCompleted(c.ID(), reqID)
	if err != nil || !processed {
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

func (c *chainObj) Mempool() chain.Mempool {
	return c.mempool
}
