package state

import (
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/require"
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

	vs1.KVStore().Set("num", codec.EncodeInt64(int64(123)))
	vs1.KVStore().Set("kuku", codec.EncodeString("A"))
	vs1.KVStore().Set("mumu", codec.EncodeString("B"))

	vs2.KVStore().Set("mumu", codec.EncodeString("B"))
	vs2.KVStore().Set("kuku", codec.EncodeString("A"))
	vs2.KVStore().Set("num", codec.EncodeInt64(int64(123)))

	require.EqualValues(t, vs1.Hash(), vs2.Hash())

	vs3 := vs1.Clone()
	vs4 := vs2.Clone()

	require.EqualValues(t, vs3.Hash(), vs4.Hash())
}

func TestApply(t *testing.T) {
	txid1 := ledgerstate.TransactionID(hashing.HashStrings("test string 1"))
	reqid1 := ledgerstate.NewOutputID(txid1, 5)
	su1 := NewStateUpdate(coretypes.RequestID(reqid1))

	require.EqualValues(t, su1.RequestID(), reqid1)

	txid2 := ledgerstate.TransactionID(hashing.HashStrings("test string 2"))
	reqid2 := ledgerstate.NewOutputID(txid2, 2)
	su2 := NewStateUpdate(coretypes.RequestID(reqid2))
	suwrong := NewStateUpdate(coretypes.RequestID(reqid2))

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

	outID1 := ledgerstate.NewOutputID(txid1, 0)

	batch1.WithApprovingOutputID(outID1)
	require.EqualValues(t, batch1.EssenceHash(), batch2.EssenceHash())

	batch2.WithApprovingOutputID(outID1)
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

	su1 := NewStateUpdate(coretypes.RequestID(reqid1))
	su2 := NewStateUpdate(coretypes.RequestID(reqid2))
	su3 := NewStateUpdate(coretypes.RequestID(reqid3))

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

	su1 := NewStateUpdate(coretypes.RequestID(reqid1))
	su2 := NewStateUpdate(coretypes.RequestID(reqid2))

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

func TestStateReader(t *testing.T) {
	log := testlogger.NewLogger(t)
	dbp := dbprovider.NewInMemoryDBProvider(log)
	vs := NewVirtualState(dbp.GetPartition(nil), nil)
	writer := vs.KVStore()
	stateReader := NewStateReader(dbp, nil)
	reader := stateReader.KVStoreReader()

	writer.Set("key1", []byte("data1"))
	writer.Set("key2", []byte("data2"))
	err := vs.CommitToDb(MustNewOriginBlock())
	require.NoError(t, err)

	back1 := reader.MustGet("key1")
	back2 := reader.MustGet("key2")
	require.EqualValues(t, "data1", string(back1))
	require.EqualValues(t, "data2", string(back2))
}

func TestOriginHash(t *testing.T) {
	origBlock := MustNewOriginBlock()
	t.Logf("origin block hash = %s", origBlock.EssenceHash().String())
	require.EqualValues(t, OriginBlockHashBase58, origBlock.EssenceHash().String())
	t.Logf("origin state hash = %s", OriginStateHash().String())
	t.Logf("zero state hash = %s", NewZeroVirtualState(mapdb.NewMapDB()).Hash().String())
	require.EqualValues(t, OriginStateHashBase58, OriginStateHash().String())

	emptyState := NewVirtualState(mapdb.NewMapDB(), nil)
	err := emptyState.ApplyBlock(origBlock)
	require.NoError(t, err)
	require.EqualValues(t, emptyState.Hash(), OriginStateHash())
}
