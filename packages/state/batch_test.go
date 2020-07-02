package state

import (
	"testing"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/assert"
)

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
	assert.Equal(t, uint16(2), batch1.Size())

	batch2, err := NewBatch([]StateUpdate{su1, su2})
	assert.NoError(t, err)
	batch2.WithStateIndex(2)
	assert.Equal(t, uint16(2), batch2.Size())

	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	batch1.WithStateTransaction(txid1)
	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	batch2.WithStateTransaction(txid1)
	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	assert.EqualValues(t, util.GetHashValue(batch1), util.GetHashValue(batch2))
}

func TestBatchMarshaling(t *testing.T) {
	txid1 := (transaction.ID)(*hashing.HashStrings("test string 1"))
	reqid1 := sctransaction.NewRequestId(txid1, 0)
	reqid2 := sctransaction.NewRequestId(txid1, 2)
	su1 := NewStateUpdate(&reqid1)
	su1.Mutations().Add(kv.NewMutationSet("k", []byte{1}))
	su2 := NewStateUpdate(&reqid2)
	su1.Mutations().Add(kv.NewMutationSet("k", []byte{2}))
	batch1, err := NewBatch([]StateUpdate{su1, su2})
	assert.NoError(t, err)
	batch1.WithStateIndex(2)

	b, err := util.Bytes(batch1)
	assert.NoError(t, err)

	batch2, err := BatchFromBytes(b)
	assert.NoError(t, err)

	assert.EqualValues(t, util.GetHashValue(batch1), util.GetHashValue(batch2))
}
