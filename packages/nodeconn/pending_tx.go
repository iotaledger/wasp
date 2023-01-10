// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconn

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/atomic"

	iotago "github.com/iotaledger/iota.go/v3"
)

// pendingTransaction holds info about a sent transaction that is pending.
type pendingTransaction struct {
	ctx            context.Context
	ctxCancel      context.CancelFunc
	transaction    *iotago.Transaction
	consumedInputs iotago.OutputIDs
	transactionID  iotago.TransactionID
	conflicting    *atomic.Bool
	conflictReason error
	confirmed      *atomic.Bool

	blockID     iotago.BlockID
	blockIDLock sync.RWMutex
}

func newPendingTransaction(ctxPendingTransaction context.Context, cancelPendingTransaction context.CancelFunc, transaction *iotago.Transaction) (*pendingTransaction, error) {
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

	return &pendingTransaction{
		ctx:            ctxPendingTransaction,
		ctxCancel:      cancelPendingTransaction,
		transaction:    transaction,
		consumedInputs: consumedInputs,
		transactionID:  txID,
		conflicting:    atomic.NewBool(false),
		conflictReason: nil,
		confirmed:      atomic.NewBool(false),
		blockID:        iotago.EmptyBlockID(),
		blockIDLock:    sync.RWMutex{},
	}, nil
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

func (tx *pendingTransaction) Confirmed() bool {
	return tx.confirmed.Load()
}

func (tx *pendingTransaction) SetConfirmed() {
	tx.confirmed.Store(true)
	tx.ctxCancel()
}

func (tx *pendingTransaction) Conflicting() bool {
	return tx.conflicting.Load()
}

func (tx *pendingTransaction) SetConflicting(reason error) {
	tx.conflictReason = reason
	tx.conflicting.Store(true)
	tx.ctxCancel()
}

func (tx *pendingTransaction) ConflictReason() error {
	return tx.conflictReason
}

// WaitUntilConfirmed waits until a given tx Block is confirmed, it takes care of promotions/re-attachments for that Block
func (tx *pendingTransaction) WaitUntilConfirmed() error {
	// wait until the context is done
	<-tx.ctx.Done()

	if tx.Conflicting() {
		return fmt.Errorf("transaction was conflicting: %s, error: %w", tx.transactionID.ToHex(), tx.conflictReason)
	}

	if !tx.Confirmed() {
		return fmt.Errorf("context was canceled but transaction was not confirmed: %s, error: %w", tx.transactionID.ToHex(), tx.ctx.Err())
	}

	// transaction was confirmed
	return nil
}
