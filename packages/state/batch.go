package state

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type batch struct {
	stateUpdates []StateUpdate
}

// validates and creates a batch from array of state updates. Array can be not sorted bu batchIndex
func NewBatch(stateUpdates []StateUpdate) (Batch, error) {
	if len(stateUpdates) == 0 {
		return nil, fmt.Errorf("batch can't be empty")
	}
	stateIndex := stateUpdates[0].StateIndex()
	txId := stateUpdates[0].StateTransactionId()

	stateUpdatesNew := make([]StateUpdate, len(stateUpdates))

	for i, su := range stateUpdates {
		if su.StateIndex() != stateIndex {
			return nil, fmt.Errorf("different stateIndex values")
		}
		if su.StateTransactionId() != txId {
			return nil, fmt.Errorf("different stateTransactionId values")
		}
		for j := i + 1; j < len(stateUpdates); j++ {
			if *su.RequestId() == *stateUpdates[j].RequestId() {
				return nil, fmt.Errorf("duplicate request id")
			}
		}
		if int(su.BatchIndex()) >= len(stateUpdatesNew) {
			return nil, fmt.Errorf("wrong batch index")
		}
		if stateUpdatesNew[su.BatchIndex()] != nil {
			return nil, fmt.Errorf("duplicate batch index")
		}
		stateUpdatesNew[su.BatchIndex()] = su
	}
	return &batch{stateUpdates: stateUpdatesNew}, nil
}

func (b *batch) SetStateTransactionId(vtxid valuetransaction.ID) {
	for _, su := range b.stateUpdates {
		su.SetStateTransactionId(vtxid)
	}
}

func (b *batch) ForEach(fun func(StateUpdate) bool) bool {
	for _, su := range b.stateUpdates {
		if fun(su) {
			return true
		}
	}
	return false
}

func (b *batch) StateIndex() uint32 {
	return b.stateUpdates[0].StateIndex()
}

func (b *batch) StateTransactionId() valuetransaction.ID {
	return b.stateUpdates[0].StateTransactionId()
}

func (b *batch) Size() uint16 {
	return uint16(len(b.stateUpdates))
}

func (b *batch) Hash() *hashing.HashValue {
	hashes := make([][]byte, len(b.stateUpdates))
	for i, su := range b.stateUpdates {
		hashes[i] = su.Hash().Bytes()
	}
	return hashing.HashData(hashes...)
}

func (b *batch) RequestIds() []*sctransaction.RequestId {
	ret := make([]*sctransaction.RequestId, b.Size())
	for i, su := range b.stateUpdates {
		ret[i] = su.RequestId()
	}
	return ret
}

func (b *batch) saveToDb(addr *address.Address) error {
	db, err := database.GetStateUpdateDB()
	if err != nil {
		return err
	}
	for _, su := range b.stateUpdates {
		err := db.Set(database.Entry{
			Key:   database.DbKeyStateUpdate(addr, b.StateIndex(), su.BatchIndex()),
			Value: hashing.MustBytes(su.(*stateUpdate)),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func LoadBatch(addr *address.Address, stateIndex uint32) (Batch, error) {
	db, err := database.GetStateUpdateDB()
	if err != nil {
		return nil, err
	}
	prefix := database.DbPrefixState(addr, stateIndex)

	stateUpdates := make([]StateUpdate, 0)
	err = db.ForEachPrefix(prefix, func(entry database.Entry) (stop bool) {
		if su, err := NewStateUpdateRead(bytes.NewReader(entry.Value)); err == nil {
			stateUpdates = append(stateUpdates, su)
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	return NewBatch(stateUpdates)
}
