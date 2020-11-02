package state

import (
	"testing"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/assert"
)

func TestVariableStateBasic(t *testing.T) {
	chainID := coretypes.ChainID{1, 3, 3, 7}
	vs1 := NewVirtualState(mapdb.NewMapDB(), &chainID)
	h1 := vs1.Hash()
	assert.EqualValues(t, *hashing.NilHash, *h1)
	assert.Equal(t, vs1.StateIndex(), uint32(0))

	vs2 := vs1.Clone()
	h2 := vs2.Hash()
	assert.EqualValues(t, h1, h2)
	assert.EqualValues(t, vs1.StateIndex(), vs1.StateIndex())

	vs1.Variables().Codec().SetInt64("num", int64(123))
	vs1.Variables().Codec().SetString("kuku", "A")
	vs1.Variables().Codec().SetString("mumu", "B")

	vs2.Variables().Codec().SetString("mumu", "B")
	vs2.Variables().Codec().SetString("kuku", "A")
	vs2.Variables().Codec().SetInt64("num", int64(123))

	assert.EqualValues(t, vs1.Hash(), vs2.Hash())

	vs3 := vs1.Clone()
	vs4 := vs2.Clone()

	assert.EqualValues(t, vs3.Hash(), vs4.Hash())
}

func TestApply(t *testing.T) {
	txid1 := (transaction.ID)(*hashing.HashStrings("test string 1"))
	reqid1 := coretypes.NewRequestID(txid1, 5)
	su1 := NewStateUpdate(&reqid1)

	assert.EqualValues(t, *su1.RequestID(), reqid1)

	txid2 := (transaction.ID)(*hashing.HashStrings("test string 2"))
	reqid2 := coretypes.NewRequestID(txid2, 2)
	su2 := NewStateUpdate(&reqid2)
	suwrong := NewStateUpdate(&reqid2)

	assert.EqualValues(t, *su2.RequestID(), reqid2)

	_, err := NewBlock(nil)
	assert.Equal(t, err == nil, false)

	_, err = NewBlock([]StateUpdate{su1, su1})
	assert.Equal(t, err == nil, false)

	_, err = NewBlock([]StateUpdate{su2, suwrong})
	assert.Equal(t, err == nil, false)

	batch1, err := NewBlock([]StateUpdate{su1, su2})
	assert.NoError(t, err)
	assert.Equal(t, uint16(2), batch1.Size())

	batch2, err := NewBlock([]StateUpdate{su1, su2})
	assert.NoError(t, err)
	assert.Equal(t, uint16(2), batch2.Size())

	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	batch1.WithStateTransaction(txid1)
	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	batch2.WithStateTransaction(txid1)
	assert.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	assert.EqualValues(t, util.GetHashValue(batch1), util.GetHashValue(batch2))

	chainID := coretypes.ChainID{1, 3, 3, 7}
	db := mapdb.NewMapDB()
	vs1 := NewVirtualState(db, &chainID)
	vs2 := NewVirtualState(db, &chainID)

	err = vs1.ApplyBatch(batch1)
	assert.NoError(t, err)

	err = vs2.ApplyBatch(batch2)
	assert.NoError(t, err)
}

func TestApply2(t *testing.T) {
	txid1 := (transaction.ID)(*hashing.HashStrings("test string 1"))
	reqid1 := coretypes.NewRequestID(txid1, 0)
	reqid2 := coretypes.NewRequestID(txid1, 2)
	reqid3 := coretypes.NewRequestID(txid1, 5)

	su1 := NewStateUpdate(&reqid1)
	su2 := NewStateUpdate(&reqid2)
	su3 := NewStateUpdate(&reqid3)

	chainID := coretypes.ChainID{1, 3, 3, 7}
	db := mapdb.NewMapDB()
	vs1 := NewVirtualState(db, &chainID)
	vs2 := NewVirtualState(db, &chainID)

	batch23, err := NewBlock([]StateUpdate{su2, su3})
	assert.NoError(t, err)
	batch23.WithBlockIndex(1)

	batch3, err := NewBlock([]StateUpdate{su3})
	assert.NoError(t, err)
	batch3.WithBlockIndex(1)

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
	reqid1 := coretypes.NewRequestID(txid1, 0)
	reqid2 := coretypes.NewRequestID(txid1, 2)

	su1 := NewStateUpdate(&reqid1)
	su2 := NewStateUpdate(&reqid2)

	chainID := coretypes.ChainID{1, 3, 3, 7}
	db := mapdb.NewMapDB()
	vs1 := NewVirtualState(db, &chainID)
	vs2 := NewVirtualState(db, &chainID)

	vs1.ApplyStateUpdate(su1)
	vs1.ApplyStateUpdate(su2)
	vs1.ApplyStateIndex(0)

	batch, err := NewBlock([]StateUpdate{su1, su2})
	assert.NoError(t, err)
	err = vs2.ApplyBatch(batch)
	assert.NoError(t, err)

	assert.EqualValues(t, vs1.Hash(), vs2.Hash())
}

func TestCommit(t *testing.T) {
	tmpdb, _ := database.NewMemDB()
	db := tmpdb.NewStore()

	partition := db.WithRealm([]byte("2"))

	txid1 := (transaction.ID)(*hashing.HashStrings("test string 1"))
	reqid1 := coretypes.NewRequestID(txid1, 5)
	su1 := NewStateUpdate(&reqid1)

	su1.Mutations().Add(buffered.NewMutationSet("x", []byte{1}))

	batch1, err := NewBlock([]StateUpdate{su1})
	assert.NoError(t, err)

	chainID := coretypes.ChainID{1, 3, 3, 7}
	vs1 := NewVirtualState(partition, &chainID)
	err = vs1.ApplyBatch(batch1)
	assert.NoError(t, err)

	v, _ := vs1.Variables().Get(kv.Key([]byte("x")))
	assert.Equal(t, []byte{1}, v)

	v, _ = partition.Get(dbkeyStateVariable(kv.Key([]byte("x"))))
	assert.Nil(t, v)

	err = vs1.CommitToDb(batch1)
	assert.NoError(t, err)

	v, _ = vs1.Variables().Get(kv.Key([]byte("x")))
	assert.Equal(t, []byte{1}, v)

	v, _ = partition.Get(dbkeyStateVariable(kv.Key([]byte("x"))))
	assert.Equal(t, []byte{1}, v)

	vs1_2, batch1_2, _, err := loadSolidState(partition, &chainID)

	assert.NoError(t, err)
	assert.EqualValues(t, util.GetHashValue(batch1), util.GetHashValue(batch1_2))
	assert.EqualValues(t, vs1.Hash(), vs1_2.Hash())

	v, _ = vs1_2.Variables().Get(kv.Key([]byte("x")))
	assert.Equal(t, []byte{1}, v)

	txid2 := (transaction.ID)(*hashing.HashStrings("test string 2"))
	reqid2 := coretypes.NewRequestID(txid2, 6)
	su2 := NewStateUpdate(&reqid2)

	su2.Mutations().Add(buffered.NewMutationDel("x"))

	batch2, err := NewBlock([]StateUpdate{su2})
	assert.NoError(t, err)

	chainID = coretypes.ChainID{1, 3, 3, 8}
	vs2 := NewVirtualState(partition, &chainID)
	err = vs2.ApplyBatch(batch2)
	assert.NoError(t, err)

	v, _ = vs2.Variables().Get(kv.Key([]byte("x")))
	assert.Nil(t, v)

	v, _ = partition.Get(dbkeyStateVariable(kv.Key([]byte("x"))))
	assert.Equal(t, []byte{1}, v)

	err = vs2.CommitToDb(batch2)
	assert.NoError(t, err)

	v, _ = partition.Get(dbkeyStateVariable(kv.Key([]byte("x"))))
	assert.Nil(t, v)
}
