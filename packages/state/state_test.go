package state

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
)

func TestVariableStateBasic(t *testing.T) {
	chainID := coretypes.NewChainID(ledgerstate.NewAliasAddress([]byte("dummy")))
	vs1 := NewVirtualState(mapdb.NewMapDB(), chainID)
	h1 := vs1.Hash()
	require.EqualValues(t, hashing.NilHash, h1)
	require.Equal(t, vs1.BlockIndex(), uint32(0))

	vs2 := vs1.Clone()
	h2 := vs2.Hash()
	require.EqualValues(t, h1, h2)
	require.EqualValues(t, vs1.BlockIndex(), vs1.BlockIndex())

	vs1.Variables().Set("num", codec.EncodeInt64(int64(123)))
	vs1.Variables().Set("kuku", codec.EncodeString("A"))
	vs1.Variables().Set("mumu", codec.EncodeString("B"))

	vs2.Variables().Set("mumu", codec.EncodeString("B"))
	vs2.Variables().Set("kuku", codec.EncodeString("A"))
	vs2.Variables().Set("num", codec.EncodeInt64(int64(123)))

	require.EqualValues(t, vs1.Hash(), vs2.Hash())

	vs3 := vs1.Clone()
	vs4 := vs2.Clone()

	require.EqualValues(t, vs3.Hash(), vs4.Hash())
}

func TestApply(t *testing.T) {
	txid1 := ledgerstate.TransactionID(hashing.HashStrings("test string 1"))
	reqid1 := ledgerstate.NewOutputID(txid1, 5)
	su1 := NewStateUpdate(reqid1)

	require.EqualValues(t, su1.RequestID(), reqid1)

	txid2 := ledgerstate.TransactionID(hashing.HashStrings("test string 2"))
	reqid2 := ledgerstate.NewOutputID(txid2, 2)
	su2 := NewStateUpdate(reqid2)
	suwrong := NewStateUpdate(reqid2)

	require.EqualValues(t, su2.RequestID(), reqid2)

	_, err := NewBlock()
	require.Error(t, err)

	_, err = NewBlock(su1, su1)
	require.Error(t, err)

	_, err = NewBlock(su2, suwrong)
	require.Error(t, err)

	batch1, err := NewBlock(su1, su2)
	require.NoError(t, err)
	require.Equal(t, uint16(2), batch1.Size())

	batch2, err := NewBlock(su1, su2)
	require.NoError(t, err)
	require.Equal(t, uint16(2), batch2.Size())

	require.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	batch1.WithStateTransaction(txid1)
	require.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	batch2.WithStateTransaction(txid1)
	require.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	require.EqualValues(t, util.GetHashValue(batch1), util.GetHashValue(batch2))

	chainID := coretypes.NewChainID(ledgerstate.NewAliasAddress([]byte("dummy")))
	db := mapdb.NewMapDB()
	vs1 := NewVirtualState(db, chainID)
	vs2 := NewVirtualState(db, chainID)

	err = vs1.ApplyBlock(batch1)
	require.NoError(t, err)

	err = vs2.ApplyBlock(batch2)
	require.NoError(t, err)
}

func TestApply2(t *testing.T) {
	txid1 := ledgerstate.TransactionID(hashing.HashStrings("test string 2"))
	reqid1 := ledgerstate.NewOutputID(txid1, 0)
	reqid2 := ledgerstate.NewOutputID(txid1, 2)
	reqid3 := ledgerstate.NewOutputID(txid1, 5)

	su1 := NewStateUpdate(reqid1)
	su2 := NewStateUpdate(reqid2)
	su3 := NewStateUpdate(reqid3)

	chainID := coretypes.NewChainID(ledgerstate.NewAliasAddress([]byte("dummy")))
	db := mapdb.NewMapDB()
	vs1 := NewVirtualState(db, chainID)
	vs2 := NewVirtualState(db, chainID)

	batch23, err := NewBlock(su2, su3)
	require.NoError(t, err)
	batch23.WithBlockIndex(1)

	batch3, err := NewBlock(su3)
	require.NoError(t, err)
	batch3.WithBlockIndex(1)

	vs1.ApplyStateUpdate(su1)
	err = vs1.ApplyBlock(batch23)
	require.NoError(t, err)

	vs2.ApplyStateUpdate(su1)
	vs2.ApplyStateUpdate(su2)
	err = vs2.ApplyBlock(batch3)
	require.NoError(t, err)

	require.EqualValues(t, vs1.BlockIndex(), vs2.BlockIndex())

	require.EqualValues(t, vs1.Hash(), vs2.Hash())
}

func TestApply3(t *testing.T) {
	txid1 := ledgerstate.TransactionID(hashing.HashStrings("test string 2"))
	reqid1 := ledgerstate.NewOutputID(txid1, 0)
	reqid2 := ledgerstate.NewOutputID(txid1, 2)

	su1 := NewStateUpdate(reqid1)
	su2 := NewStateUpdate(reqid2)

	chainID := coretypes.NewChainID(ledgerstate.NewAliasAddress([]byte("dummy")))
	db := mapdb.NewMapDB()
	vs1 := NewVirtualState(db, chainID)
	vs2 := NewVirtualState(db, chainID)

	vs1.ApplyStateUpdate(su1)
	vs1.ApplyStateUpdate(su2)
	vs1.ApplyBlockIndex(0)

	batch, err := NewBlock(su1, su2)
	require.NoError(t, err)
	err = vs2.ApplyBlock(batch)
	require.NoError(t, err)

	require.EqualValues(t, vs1.Hash(), vs2.Hash())
}

func TestCommit(t *testing.T) {
	tmpdb, _ := database.NewMemDB()
	db := tmpdb.NewStore()

	partition := db.WithRealm([]byte("2"))

	txid1 := ledgerstate.TransactionID(hashing.HashStrings("test string 2"))
	reqid1 := ledgerstate.NewOutputID(txid1, 5)
	su1 := NewStateUpdate(reqid1)

	su1.Mutations().Add(buffered.NewMutationSet("x", []byte{1}))

	batch1, err := NewBlock(su1)
	require.NoError(t, err)

	chainID := coretypes.NewChainID(ledgerstate.NewAliasAddress([]byte("dummy")))
	vs1 := NewVirtualState(partition, chainID)
	err = vs1.ApplyBlock(batch1)
	require.NoError(t, err)

	v, _ := vs1.Variables().Get("x")
	require.Equal(t, []byte{1}, v)

	v, _ = partition.Get(dbkeyStateVariable("x"))
	require.Nil(t, v)

	err = vs1.CommitToDb(batch1)
	require.NoError(t, err)

	v, _ = vs1.Variables().Get("x")
	require.Equal(t, []byte{1}, v)

	v, _ = partition.Get(dbkeyStateVariable("x"))
	require.Equal(t, []byte{1}, v)

	vs1_2, batch1_2, _, err := loadSolidState(partition, chainID)

	require.NoError(t, err)
	require.EqualValues(t, util.GetHashValue(batch1), util.GetHashValue(batch1_2))
	require.EqualValues(t, vs1.Hash(), vs1_2.Hash())

	v, _ = vs1_2.Variables().Get(kv.Key([]byte("x")))
	require.Equal(t, []byte{1}, v)

	txid2 := ledgerstate.TransactionID(hashing.HashStrings("test string 2"))
	reqid2 := ledgerstate.NewOutputID(txid2, 6)
	su2 := NewStateUpdate(reqid2)

	su2.Mutations().Add(buffered.NewMutationDel("x"))

	batch2, err := NewBlock(su2)
	require.NoError(t, err)

	chainID = coretypes.NewChainID(ledgerstate.NewAliasAddress([]byte("dummy2")))
	vs2 := NewVirtualState(partition, chainID)
	err = vs2.ApplyBlock(batch2)
	require.NoError(t, err)

	v, _ = vs2.Variables().Get(kv.Key([]byte("x")))
	require.Nil(t, v)

	v, _ = partition.Get(dbkeyStateVariable(kv.Key([]byte("x"))))
	require.Equal(t, []byte{1}, v)

	err = vs2.CommitToDb(batch2)
	require.NoError(t, err)

	v, _ = partition.Get(dbkeyStateVariable(kv.Key([]byte("x"))))
	require.Nil(t, v)
}
