package state

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/iotaledger/wasp/packages/kv/buffered"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/assert"
)

func TestBatches(t *testing.T) {
	txid1 := ledgerstate.TransactionID(hashing.HashStrings("test string 1"))
	reqid1 := ledgerstate.NewOutputID(txid1, 5)
	su1 := NewStateUpdate(coretypes.RequestID(reqid1))

	assert.EqualValues(t, su1.RequestID(), reqid1)

	txid2 := ledgerstate.TransactionID(hashing.HashStrings("test string 2"))
	reqid2 := ledgerstate.NewOutputID(txid2, 2)
	su2 := NewStateUpdate(coretypes.RequestID(reqid2))

	assert.EqualValues(t, su2.RequestID(), reqid2)

	_, err := NewBlock()
	assert.Error(t, err)

	batch1, err := NewBlock(su1, su2)
	assert.NoError(t, err)
	batch1.WithBlockIndex(2)
	assert.Equal(t, uint16(2), batch1.Size())

	batch2, err := NewBlock(su1, su2)
	assert.NoError(t, err)
	batch2.WithBlockIndex(2)
	assert.Equal(t, uint16(2), batch2.Size())

	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	outID := ledgerstate.NewOutputID(txid1, 0)
	batch1.WithApprovingOutputID(outID)
	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	batch2.WithApprovingOutputID(outID)
	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	assert.EqualValues(t, util.GetHashValue(batch1), util.GetHashValue(batch2))
}

func TestBatchMarshaling(t *testing.T) {
	txid1 := ledgerstate.TransactionID(hashing.HashStrings("test string 1"))
	reqid1 := ledgerstate.NewOutputID(txid1, 0)
	reqid2 := ledgerstate.NewOutputID(txid1, 2)
	t.Logf("req1: %s", coretypes.RequestID(reqid1).String())
	t.Logf("req2: %s", coretypes.RequestID(reqid2).String())

	su1 := NewStateUpdate(coretypes.RequestID(reqid1))
	su1.Mutations().Add(buffered.NewMutationSet("k", []byte{1}))
	su2 := NewStateUpdate(coretypes.RequestID(reqid2))
	su2.Mutations().Add(buffered.NewMutationSet("k", []byte{2}))
	batch1, err := NewBlock(su1, su2)
	assert.NoError(t, err)
	batch1.WithBlockIndex(2)

	b, err := util.Bytes(batch1)
	assert.NoError(t, err)

	batch2, err := BlockFromBytes(b)
	assert.NoError(t, err)

	assert.EqualValues(t, util.GetHashValue(batch1), util.GetHashValue(batch2))
}

func TestOriginBlock(t *testing.T) {
	txid1 := ledgerstate.TransactionID{}
	outID1 := ledgerstate.NewOutputID(txid1, 0)
	txid2 := ledgerstate.TransactionID(hashing.RandomHash(nil))
	outID2 := ledgerstate.NewOutputID(txid1, 0)
	require.NotEqualValues(t, txid1, txid2)
	b1 := MustNewOriginBlock(outID1).WithBlockIndex(100)
	b2 := MustNewOriginBlock(outID2).WithBlockIndex(100)
	require.EqualValues(t, b1.EssenceHash(), b2.EssenceHash())
}
