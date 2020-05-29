package state

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/stretchr/testify/assert"
	"testing"
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

	_, err := NewBatch(nil, 2)
	assert.Equal(t, err == nil, false)

	batch1, err := NewBatch([]StateUpdate{su1, su2}, 2)
	assert.NoError(t, err)
	assert.Equal(t, batch1.Size(), uint16(2))

	batch2, err := NewBatch([]StateUpdate{su1, su2}, 2)
	assert.NoError(t, err)
	assert.Equal(t, batch1.Size(), uint16(2))

	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	batch1.Commit(txid1)
	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	batch2.Commit(txid1)
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

	reqid3 := sctransaction.NewRequestId(txid2, 0)
	su3 := NewStateUpdate(&reqid3)

	assert.EqualValues(t, *su2.RequestId(), reqid2)

	_, err := NewBatch(nil, 2)
	assert.Equal(t, err == nil, false)

	_, err = NewBatch([]StateUpdate{su1, su1}, 0)
	assert.Equal(t, err == nil, false)

	_, err = NewBatch([]StateUpdate{su2, suwrong}, 0)
	assert.Equal(t, err == nil, false)

	batch1, err := NewBatch([]StateUpdate{su1, su2}, 0)
	assert.NoError(t, err)
	assert.Equal(t, batch1.Size(), uint16(2))

	batch2, err := NewBatch([]StateUpdate{su1, su2}, 0)
	assert.NoError(t, err)
	assert.Equal(t, batch1.Size(), uint16(2))

	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	batch1.Commit(txid1)
	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	batch2.Commit(txid1)
	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	assert.EqualValues(t, hashing.GetHashValue(batch1), hashing.GetHashValue(batch2))

	vs1 := NewVariableState(nil)
	vs2 := NewVariableState(nil)
	vs3 := NewVariableState(nil)

	err = vs1.ApplyBatch(batch1)
	assert.NoError(t, err)

	err = vs2.ApplyBatch(batch2)
	assert.NoError(t, err)

	assert.EqualValues(t, vs1.Hash(), vs2.Hash())

	batch3, err := NewBatch([]StateUpdate{su1}, 2)
	assert.NoError(t, err)
	err = vs3.ApplyBatch(batch3)
	assert.Equal(t, err != nil, true)

	su1.Variables().Set("kuku", "A")
	su2.Variables().Set("mumu", "B")

	assert.EqualValues(t, vs1.Hash(), vs2.Hash())

	vs1.ApplyStateUpdate(su1)
	b, err := NewBatch([]StateUpdate{su2, su3}, 1)
	assert.NoError(t, err)
	err = vs1.ApplyBatch(b)
	assert.NoError(t, err)

	batch4, err := NewBatch([]StateUpdate{su1, su2, su3}, 1)
	assert.NoError(t, err)
	err = vs2.ApplyBatch(batch4)
	assert.NoError(t, err)

	assert.EqualValues(t, vs1.StateIndex(), vs2.StateIndex())

	assert.EqualValues(t, vs1.Hash(), vs2.Hash())
}
