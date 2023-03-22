// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconn

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.uber.org/atomic"

	iotago "github.com/iotaledger/iota.go/v3"
)

// pendingTransaction holds info about a sent transaction that is pending.
type pendingTransaction struct {
	// this is the context given by the chain consensus.
	// if this context gets canceled, the tx should not be tracked by the node connection anymore.
	ctxChainConsensus context.Context

	// this context is used to signal that the transaction got referenced by a milestone.
	// it might be confirmed or conflicting, or the parent ctxChainConsensus got canceled.
	ctxConfirmed       context.Context
	cancelCtxConfirmed context.CancelFunc

	ncChain        *ncChain
	transaction    *iotago.Transaction
	consumedInputs iotago.OutputIDs
	transactionID  iotago.TransactionID
	conflicting    *atomic.Bool
	conflictReason error
	confirmed      *atomic.Bool

	blockID     iotago.BlockID
	blockIDLock sync.RWMutex

	lastPendingTx        *pendingTransaction
	chainedPendingTxLock sync.RWMutex
	chainedPendingTx     *pendingTransaction
}

func newPendingTransaction(ctxChainConsensus context.Context, ncChain *ncChain, transaction *iotago.Transaction, lastPendingTx *pendingTransaction) (*pendingTransaction, error) {
	txID, err := transaction.ID()
	if err != nil {
		return nil, err
	}

	// collect consumed inputs
	consumedInputs := make([]iotago.OutputID, 0, len(transaction.Essence.Inputs))
	for _, input := range transaction.Essence.Inputs {
		switch input.Type() {
		case iotago.InputUTXO:
			consumedInputs = append(consumedInputs, input.(*iotago.UTXOInput).ID())
		default:
			return nil, fmt.Errorf("%w: type %d", iotago.ErrUnknownInputType, input.Type())
		}
	}

	ctxConfirmed, cancelCtxConfirmed := context.WithCancel(ctxChainConsensus)

	pendingTx := &pendingTransaction{
		ctxChainConsensus:    ctxChainConsensus,
		ctxConfirmed:         ctxConfirmed,
		cancelCtxConfirmed:   cancelCtxConfirmed,
		ncChain:              ncChain,
		transaction:          transaction,
		consumedInputs:       consumedInputs,
		transactionID:        txID,
		lastPendingTx:        lastPendingTx,
		conflicting:          atomic.NewBool(false),
		conflictReason:       nil,
		confirmed:            atomic.NewBool(false),
		blockID:              iotago.EmptyBlockID(),
		blockIDLock:          sync.RWMutex{},
		chainedPendingTxLock: sync.RWMutex{},
		chainedPendingTx:     nil,
	}

	// chain the new transaction with the last pending one
	if lastPendingTx != nil {
		lastPendingTx.setChainedPendingTransaction(pendingTx)
	}

	return pendingTx, nil
}

func (tx *pendingTransaction) Cleanup() {
	tx.chainedPendingTxLock.RLock()
	defer tx.chainedPendingTxLock.RUnlock()

	tx.lastPendingTx = nil
	tx.chainedPendingTx = nil
}

func (tx *pendingTransaction) ID() iotago.TransactionID {
	return tx.transactionID
}

func (tx *pendingTransaction) Transaction() *iotago.Transaction {
	return tx.transaction
}

func (tx *pendingTransaction) ConsumedInputs() iotago.OutputIDs {
	return tx.consumedInputs
}

func (tx *pendingTransaction) BlockID() iotago.BlockID {
	tx.blockIDLock.RLock()
	defer tx.blockIDLock.RUnlock()

	return tx.blockID
}

func (tx *pendingTransaction) SetBlockID(blockID iotago.BlockID) {
	tx.blockIDLock.Lock()
	defer tx.blockIDLock.Unlock()
	tx.blockID = blockID
}

func (tx *pendingTransaction) Reattach() {
	tx.ncChain.reattachTx(tx)
}

func (tx *pendingTransaction) propagateReattach() {
	// propagate the new blockID to chained transactions.
	// we need to reattach pending transactions that reference
	// this transaction to fix the ordering of the outputs on L1.
	tx.chainedPendingTxLock.RLock()
	defer tx.chainedPendingTxLock.RUnlock()

	if tx.chainedPendingTx != nil {
		tx.chainedPendingTx.Reattach()
	}
}

func (tx *pendingTransaction) Confirmed() bool {
	return tx.confirmed.Load()
}

func (tx *pendingTransaction) SetConfirmed() {
	tx.confirmed.Store(true)
	tx.cancelCtxConfirmed()
}

func (tx *pendingTransaction) Conflicting() bool {
	return tx.conflicting.Load()
}

func (tx *pendingTransaction) SetConflicting(reason error) {
	tx.conflictReason = reason
	tx.conflicting.Store(true)
	tx.cancelCtxConfirmed()

	// propagate the conflict to chained transactions
	tx.chainedPendingTxLock.RLock()
	defer tx.chainedPendingTxLock.RUnlock()

	if tx.chainedPendingTx != nil {
		tx.chainedPendingTx.SetConflicting(errors.New("former chained transaction was conflicting"))
	}
}

func (tx *pendingTransaction) setChainedPendingTransaction(pendingTx *pendingTransaction) {
	tx.chainedPendingTxLock.Lock()
	defer tx.chainedPendingTxLock.Unlock()

	tx.chainedPendingTx = pendingTx
}

func (tx *pendingTransaction) ConflictReason() error {
	return tx.conflictReason
}

// waitUntilConfirmed waits until a given tx Block is confirmed, it takes care of promotions/re-attachments for that Block
func (tx *pendingTransaction) waitUntilConfirmed() error {
	select {
	case <-tx.ncChain.ctx.Done():
		// canceled by shutdown signal or "Chains.Deactivate"
		return fmt.Errorf("chain context was canceled but transaction was not confirmed yet: %s, error: %w", tx.transactionID.ToHex(), tx.ncChain.ctx.Err())

	case <-tx.ctxConfirmed.Done():
		// it might be confirmed or conflicting, or the parent ctxChainConsensus got canceled.

		if tx.Conflicting() {
			return fmt.Errorf("transaction was conflicting: %s, error: %w", tx.transactionID.ToHex(), tx.conflictReason)
		}

		if !tx.Confirmed() {
			ctxChainConsensusCanceled := tx.ctxChainConsensus.Err() != nil
			return fmt.Errorf("context was canceled but transaction was not confirmed: %s, ctxChainConsensusCanceled: %t, error: %w", tx.transactionID.ToHex(), ctxChainConsensusCanceled, tx.ctxConfirmed.Err())
		}

		// transaction was confirmed
		return nil
	}
}
