package state

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestVariableStateBasic(t *testing.T) {
	vs1 := NewVariableState(nil)
	h1 := vs1.Hash()
	assert.Equal(t, h1 == *hashing.NilHash, true)
	assert.Equal(t, vs1.StateIndex(), uint32(0))

	vs2 := NewVariableState(vs1)
	h2 := vs2.Hash()
	assert.EqualValues(t, h1, h2)
	assert.EqualValues(t, vs1.StateIndex(), vs1.StateIndex())

	vs1.Variables().Set("num", uint16(123))
	vs1.Variables().Set("kuku", "A")
	vs1.Variables().Set("mumu", "B")

	vs2.Variables().Set("mumu", "B")
	vs2.Variables().Set("kuku", "A")
	vs2.Variables().Set("num", uint16(123))

	assert.EqualValues(t, vs1.Hash(), vs2.Hash())

	vs3 := NewVariableState(vs1)
	vs4 := NewVariableState(vs2)

	assert.EqualValues(t, vs3.Hash(), vs4.Hash())
}

func TestBatches(t *testing.T) {
	txid1 := (transaction.ID)(*hashing.HashStrings("test string 1"))
	reqid1 := sctransaction.NewRequestId(txid1, 5)
	su1 := NewStateUpdate(&reqid1)

	assert.EqualValues(t, *su1.RequestId(), reqid1)

	txid2 := (transaction.ID)(*hashing.HashStrings("test string 2"))
	reqid2 := sctransaction.NewRequestId(txid2, 2)
	su2 := NewStateUpdate(&reqid2)

	assert.EqualValues(t, *su2.RequestId(), reqid2)

	_, err := NewBatch(nil)
	assert.Equal(t, err == nil, false)

	batch1, err := NewBatch([]StateUpdate{su1, su2})
	assert.NoError(t, err)
	batch1.WithStateIndex(2)
	assert.Equal(t, batch1.Size(), uint16(2))

	batch2, err := NewBatch([]StateUpdate{su1, su2})
	assert.NoError(t, err)
	batch2.WithStateIndex(2)
	assert.Equal(t, batch1.Size(), uint16(2))

	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	batch1.WithStateTransaction(txid1)
	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	batch2.WithStateTransaction(txid1)
	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	assert.EqualValues(t, hashing.GetHashValue(batch1), hashing.GetHashValue(batch2))
}

func TestApply(t *testing.T) {
	txid1 := (transaction.ID)(*hashing.HashStrings("test string 1"))
	reqid1 := sctransaction.NewRequestId(txid1, 5)
	su1 := NewStateUpdate(&reqid1)

	assert.EqualValues(t, *su1.RequestId(), reqid1)

	txid2 := (transaction.ID)(*hashing.HashStrings("test string 2"))
	reqid2 := sctransaction.NewRequestId(txid2, 2)
	su2 := NewStateUpdate(&reqid2)
	suwrong := NewStateUpdate(&reqid2)

	assert.EqualValues(t, *su2.RequestId(), reqid2)

	_, err := NewBatch(nil)
	assert.Equal(t, err == nil, false)

	_, err = NewBatch([]StateUpdate{su1, su1})
	assert.Equal(t, err == nil, false)

	_, err = NewBatch([]StateUpdate{su2, suwrong})
	assert.Equal(t, err == nil, false)

	batch1, err := NewBatch([]StateUpdate{su1, su2})
	assert.NoError(t, err)
	assert.Equal(t, batch1.Size(), uint16(2))

	batch2, err := NewBatch([]StateUpdate{su1, su2})
	assert.NoError(t, err)
	assert.Equal(t, batch1.Size(), uint16(2))

	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	batch1.WithStateTransaction(txid1)
	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	batch2.WithStateTransaction(txid1)
	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	assert.EqualValues(t, hashing.GetHashValue(batch1), hashing.GetHashValue(batch2))

	vs1 := NewVariableState(nil)
	vs2 := NewVariableState(nil)

	err = vs1.ApplyBatch(batch1)
	assert.NoError(t, err)

	err = vs2.ApplyBatch(batch2)
	assert.NoError(t, err)
}

func TestApply2(t *testing.T) {
	txid1 := (transaction.ID)(*hashing.HashStrings("test string 1"))
	reqid1 := sctransaction.NewRequestId(txid1, 0)
	reqid2 := sctransaction.NewRequestId(txid1, 2)
	reqid3 := sctransaction.NewRequestId(txid1, 5)

	su1 := NewStateUpdate(&reqid1)
	su2 := NewStateUpdate(&reqid2)
	su3 := NewStateUpdate(&reqid3)

	vs1 := NewVariableState(nil)
	vs2 := NewVariableState(nil)

	batch23, err := NewBatch([]StateUpdate{su2, su3})
	assert.NoError(t, err)
	batch23.WithStateIndex(1)

	batch3, err := NewBatch([]StateUpdate{su3})
	assert.NoError(t, err)
	batch3.WithStateIndex(1)

	vs1.ApplyStateUpdate(su1)
	err = vs1.ApplyBatch(batch23)
	assert.NoError(t, err)

	vs2.ApplyStateUpdate(su1)
	vs2.ApplyStateUpdate(su2)
	err = vs2.ApplyBatch(batch3)
	assert.NoError(t, err)

	assert.EqualValues(t, vs1.StateIndex(), vs2.StateIndex())

	assert.EqualValues(t, vs1.Hash(), vs2.Hash())
}

func TestApply3(t *testing.T) {
	txid1 := (transaction.ID)(*hashing.HashStrings("test string 1"))
	reqid1 := sctransaction.NewRequestId(txid1, 0)
	reqid2 := sctransaction.NewRequestId(txid1, 2)

	su1 := NewStateUpdate(&reqid1)
	su2 := NewStateUpdate(&reqid2)

	vs1 := NewVariableState(nil)
	vs2 := NewVariableState(nil)

	nowis := time.Now()

	vs1.ApplyStateUpdate(su1)
	vs1.ApplyStateUpdate(su2)
	vs1.ApplyStateIndex(0)
	vs1.ApplyTimestamp(nowis)

	batch, err := NewBatch([]StateUpdate{su1, su2})
	assert.NoError(t, err)
	batch.WithStateIndex(0).WithTimestamp(nowis)
	err = vs2.ApplyBatch(batch)
	assert.NoError(t, err)

	assert.EqualValues(t, vs1.Hash(), vs2.Hash())
}
