// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconn

import (
	"context"
	"sync"

	"go.uber.org/atomic"
	"golang.org/x/xerrors"

	iotago "github.com/iotaledger/iota.go/v3"
)

// PendingTransaction holds info about a sent transaction that is pending.
type PendingTransaction struct {
	ctx            context.Context
	CtxCancel      context.CancelFunc
	transaction    *iotago.Transaction
	consumedInputs iotago.OutputIDs
	transactionID  iotago.TransactionID
	conflicting    *atomic.Bool
	conflictReason error
	confirmed      *atomic.Bool

	blockID     iotago.BlockID
	blockIDLock sync.RWMutex
}

func NewPendingTransaction(ctxPendingTransaction context.Context, cancelPendingTransaction context.CancelFunc, transaction *iotago.Transaction) (*PendingTransaction, error) {
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
			return nil, xerrors.Errorf("%w: type %d", iotago.ErrUnknownInputType, input.Type())
		}
	}

	return &PendingTransaction{
		ctx:            ctxPendingTransaction,
		CtxCancel:      cancelPendingTransaction,
		transaction:    transaction,
		consumedInputs: consumedInputs,
		transactionID:  txID,
		conflicting:    atomic.NewBool(false),
		confirmed:      atomic.NewBool(false),
		blockID:        iotago.EmptyBlockID(),
		blockIDLock:    sync.RWMutex{},
	}, nil
}

func (tx *PendingTransaction) ID() iotago.TransactionID {
	return tx.transactionID
}

func (tx *PendingTransaction) Transaction() *iotago.Transaction {
	return tx.transaction
}

func (tx *PendingTransaction) ConsumedInputs() iotago.OutputIDs {
	return tx.consumedInputs
}

func (tx *PendingTransaction) BlockID() iotago.BlockID {
	tx.blockIDLock.RLock()
	defer tx.blockIDLock.RUnlock()

	return tx.blockID
}

func (tx *PendingTransaction) SetBlockID(blockID iotago.BlockID) {
	tx.blockIDLock.Lock()
	defer tx.blockIDLock.Unlock()

	tx.blockID = blockID
}

func (tx *PendingTransaction) Confirmed() bool {
	return tx.confirmed.Load()
}

func (tx *PendingTransaction) SetConfirmed() {
	tx.confirmed.Store(true)
	tx.CtxCancel()
}

func (tx *PendingTransaction) Conflicting() bool {
	return tx.conflicting.Load()
}

func (tx *PendingTransaction) SetConflicting(reason error) {
	tx.conflictReason = reason
	tx.conflicting.Store(true)
	tx.CtxCancel()
}

func (tx *PendingTransaction) ConflictReason() error {
	return tx.conflictReason
}

// WaitUntilConfirmed waits until a given tx Block is confirmed, it takes care of promotions/re-attachments for that Block
func (tx *PendingTransaction) WaitUntilConfirmed() error {
	// wait until the context is done
	<-tx.ctx.Done()

	if tx.conflicting.Load() {
		return xerrors.Errorf("transaction was conflicting: %s", tx.transactionID.ToHex())
	}

	if !tx.confirmed.Load() {
		return xerrors.Errorf("context was canceled but transaction was not confirmed: %s, error: %s", tx.transactionID.ToHex(), tx.ctx.Err())
	}

	// transaction was confirmed
	return nil
}
