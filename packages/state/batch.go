package state

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

type batch struct {
	stateIndex   uint32
	stateTxId    valuetransaction.ID
	stateUpdates []StateUpdate
}

// validates, enumerates and creates a batch from array of state updates
func NewBatch(stateUpdates []StateUpdate, stateIndex uint32) (Batch, error) {
	if len(stateUpdates) == 0 {
		return nil, fmt.Errorf("batch can't be empty")
	}
	stateUpdatesNew := make([]StateUpdate, len(stateUpdates))

	for i, su := range stateUpdates {
		for j := i + 1; j < len(stateUpdates); j++ {
			if *su.RequestId() == *stateUpdates[j].RequestId() {
				return nil, fmt.Errorf("duplicate request id")
			}
		}
		su.SetBatchIndex(uint16(i))
		stateUpdatesNew[su.BatchIndex()] = su
	}
	return &batch{
		stateIndex:   stateIndex,
		stateUpdates: stateUpdatesNew,
	}, nil
}

func (b *batch) StateTransactionId() valuetransaction.ID {
	return b.stateTxId
}

func (b *batch) SetStateTransactionId(vtxid valuetransaction.ID) {
	b.stateTxId = vtxid
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
	return b.stateIndex
}

func (b *batch) Size() uint16 {
	return uint16(len(b.stateUpdates))
}

func (b *batch) RequestIds() []*sctransaction.RequestId {
	ret := make([]*sctransaction.RequestId, b.Size())
	for i, su := range b.stateUpdates {
		ret[i] = su.RequestId()
	}
	return ret
}

func (b *batch) Write(w io.Writer) error {
	if err := util.WriteUint32(w, b.stateIndex); err != nil {
		return err
	}
	if _, err := w.Write(b.stateTxId.Bytes()); err != nil {
		return err
	}
	if err := util.WriteUint16(w, uint16(len(b.stateUpdates))); err != nil {
		return err
	}
	for _, su := range b.stateUpdates {
		if err := su.Write(w); err != nil {
			return err
		}
	}
	return nil
}

func (b *batch) Read(r io.Reader) error {
	if err := util.ReadUint32(r, &b.stateIndex); err != nil {
		return err
	}
	if _, err := r.Read(b.stateTxId[:]); err != nil {
		return err
	}
	var size uint16
	if err := util.ReadUint16(r, &size); err != nil {
		return err
	}
	b.stateUpdates = make([]StateUpdate, size)
	for i := range b.stateUpdates {
		b.stateUpdates[i] = new(stateUpdate)
		if err := b.stateUpdates[i].Read(r); err != nil {
			return err
		}
	}
	return nil
}

func (b *batch) saveToDb(addr *address.Address) error {
	dbase, err := database.GetBatchesDB()
	if err != nil {
		return err
	}
	// write batch header as separate record
	var buf bytes.Buffer
	if err := b.Write(&buf); err != nil {
		return err
	}
	err = dbase.Set(database.Entry{
		Key:   database.DbKeyBatch(addr, b.stateIndex),
		Value: buf.Bytes(),
	})
	if err != nil {
		return err
	}
	return nil
}

func LoadBatch(addr *address.Address, stateIndex uint32) (Batch, error) {
	dbase, err := database.GetBatchesDB()
	if err != nil {
		return nil, err
	}
	entry, err := dbase.Get(database.DbKeyBatch(addr, stateIndex))
	if err != nil {
		return nil, err
	}
	ret := new(batch)
	if err := ret.Read(bytes.NewReader(entry.Value)); err != nil {
		return nil, err
	}
	return ret, nil
}
