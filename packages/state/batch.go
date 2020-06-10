package state

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/database"
	"io"
)

type batch struct {
	stateIndex   uint32
	stateTxId    valuetransaction.ID
	stateUpdates []StateUpdate
	timestamp    int64
}

// validates, enumerates and creates a batch from array of state updates
func NewBatch(stateUpdates []StateUpdate) (Batch, error) {
	if len(stateUpdates) == 0 {
		return nil, fmt.Errorf("batch can't be empty")
	}
	for i, su := range stateUpdates {
		for j := i + 1; j < len(stateUpdates); j++ {
			if *su.RequestId() == *stateUpdates[j].RequestId() {
				return nil, fmt.Errorf("duplicate request id")
			}
		}
	}
	return &batch{
		stateUpdates: stateUpdates,
	}, nil
}

func MustNewBatch(stateUpdates []StateUpdate) Batch {
	ret, err := NewBatch(stateUpdates)
	if err != nil {
		panic(err)
	}
	return ret
}

func (b *batch) StateTransactionId() valuetransaction.ID {
	return b.stateTxId
}

func (b *batch) StateIndex() uint32 {
	return b.stateIndex
}

func (b *batch) Timestamp() int64 {
	return b.timestamp
}

func (b *batch) WithStateIndex(stateIndex uint32) Batch {
	b.stateIndex = stateIndex
	return b
}

func (b *batch) WithStateTransaction(vtxid valuetransaction.ID) Batch {
	b.stateTxId = vtxid
	return b
}

func (b *batch) WithTimestamp(ts int64) Batch {
	b.timestamp = ts
	return b
}

func (b *batch) ForEach(fun func(StateUpdate) bool) {
	for _, su := range b.stateUpdates {
		if !fun(su) {
			return
		}
	}
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

// hash of all data except state transaction hash
func (b *batch) EssenceHash() *hashing.HashValue {
	var buf bytes.Buffer
	if err := b.writeEssence(&buf); err != nil {
		panic("EssenceHash")
	}
	return hashing.HashData(buf.Bytes())
}

func (b *batch) Write(w io.Writer) error {
	if err := b.writeEssence(w); err != nil {
		return err
	}
	if _, err := w.Write(b.stateTxId.Bytes()); err != nil {
		return err
	}
	return nil
}

func (b *batch) writeEssence(w io.Writer) error {
	if err := util.WriteUint32(w, b.stateIndex); err != nil {
		return err
	}
	if err := util.WriteUint64(w, uint64(b.timestamp)); err != nil {
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
	if err := b.readEssence(r); err != nil {
		return err
	}
	if _, err := r.Read(b.stateTxId[:]); err != nil {
		return err
	}
	return nil
}

func (b *batch) readEssence(r io.Reader) error {
	if err := util.ReadUint32(r, &b.stateIndex); err != nil {
		return err
	}
	var ts uint64
	if err := util.ReadUint64(r, &ts); err != nil {
		return err
	}
	b.timestamp = int64(ts)
	var size uint16
	if err := util.ReadUint16(r, &size); err != nil {
		return err
	}
	b.stateUpdates = make([]StateUpdate, size)
	var err error
	for i := range b.stateUpdates {
		b.stateUpdates[i], err = NewStateUpdateRead(r)
		if err != nil {
			return err
		}
	}
	return nil
}

func dbkeyBatch(stateIndex uint32) []byte {
	return database.MakeKey(database.ObjectTypeStateUpdateBatch, util.Uint32To4Bytes(stateIndex))
}

func LoadBatch(addr *address.Address, stateIndex uint32) (Batch, error) {
	data, err := database.GetPartition(addr).Get(dbkeyBatch(stateIndex))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return BatchFromBytes(data)
}

func BatchFromBytes(data []byte) (Batch, error) {
	ret := new(batch)
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}
