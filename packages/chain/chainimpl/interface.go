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

func (c *chainObj) startTimer() {
	go func() {
		tick := 0
		for !c.IsDismissed() {
			time.Sleep(chain.TimerTickPeriod)
			c.ReceiveMessage(chain.TimerTick(tick))
			tick++
		}
	}()
}

func (c *chainObj) Dismiss() {
	c.log.Infof("Dismiss chain %s", c.chainID.Base58())

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
	reqs, err := sctransaction.RequestsOnLedgerFromTransaction(tx, c.chainID.AsAddress())
	if err != nil {
		c.log.Warnf("failed to parse transaction %s: %v", tx.ID().Base58(), err)
		return
	}
	for _, req := range reqs {
		c.ReceiveRequest(req, tx.Essence().Timestamp())
	}
	if chainOut := sctransaction.FindAliasOutput(tx.Essence(), c.chainID.AsAddress()); chainOut != nil {
		chainOut1 := chainOut.UpdateMintingColor().(*ledgerstate.AliasOutput)
		c.ReceiveState(chainOut1, tx.Essence().Timestamp())
	}
}

func (c *chainObj) ReceiveRequest(req coretypes.Request, timestamp time.Time) {
	c.log.Debugf("ReceiveRequest: %s", req.ID())
	c.mempool.ReceiveRequest(req, timestamp)
}

func (c *chainObj) ReceiveState(stateOutput *ledgerstate.AliasOutput, timestamp time.Time) {
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
	processed, err := state.IsRequestCompleted(c.dbProvider, c.ID(), reqID)
	if err != nil {
		panic(err)
	}
	if !processed {
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
