package state

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/stretchr/testify/require"
	"testing"
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

//func TestApply(t *testing.T) {
//	su1 := NewStateUpdate()
//	su1.Mutations().Set("key1", []byte("data1"))
//
//	su2 := NewStateUpdate()
//
//	block1 := NewBlock(1, su1, su2)
//	block2 := NewBlock(1, su1, su2)
//	block0 := NewBlock(0)
//
//	require.EqualValues(t, util.GetHashValue(block1), util.GetHashValue(block2))
//
//	chainID := coretypes.NewChainID(ledgerstate.NewAliasAddress([]byte("dummy")))
//	db := mapdb.NewMapDB()
//	vs1 := NewVirtualState(db, chainID)
//	vs2 := NewVirtualState(db, chainID)
//
//	err := vs1.ApplyBlock(block1)
//	require.Error(t, err)
//
//	err = vs1.ApplyBlock(block0)
//	require.NoError(t, err)
//
//	err = vs2.ApplyBlock(block0)
//	require.NoError(t, err)
//
//	err = vs1.ApplyBlock(block1)
//	require.NoError(t, err)
//
//	err = vs2.ApplyBlock(block2)
//	require.NoError(t, err)
//
//	require.EqualValues(t, vs1.Hash(), vs2.Hash())
//}
//
//func TestApply2(t *testing.T) {
//
//	su1 := NewStateUpdate()
//	su2 := NewStateUpdate()
//	su3 := NewStateUpdate()
//
//	chainID := coretypes.NewChainID(ledgerstate.NewAliasAddress([]byte("dummy")))
//	db := mapdb.NewMapDB()
//	vs1 := NewVirtualState(db, chainID)
//	vs2 := NewVirtualState(db, chainID)
//
//	block23 := NewBlock(1, su2, su3)
//
//	block3 := NewBlock(1, su3)
//
//	vs1.ApplyStateUpdate(su1)
//	err := vs1.ApplyBlock(block23)
//	require.NoError(t, err)
//
//	vs2.ApplyStateUpdate(su1)
//	vs2.ApplyStateUpdate(su2)
//	err = vs2.ApplyBlock(block3)
//	require.NoError(t, err)
//
//	require.EqualValues(t, vs1.BlockIndex(), vs2.BlockIndex())
//	require.EqualValues(t, vs1.Hash(), vs2.Hash())
//}
//
//func TestApply3(t *testing.T) {
//	nowis := time.Now()
//	su1 := NewStateUpdate(nowis)
//	su2 := NewStateUpdate(nowis.Add(1 * time.Second))
//
//	chainID := coretypes.NewChainID(ledgerstate.NewAliasAddress([]byte("dummy")))
//	db := mapdb.NewMapDB()
//	vs1 := NewVirtualState(db, chainID)
//	vs2 := NewVirtualState(db, chainID)
//
//	err := vs1.ApplyBlock(NewBlock(0))
//	require.NoError(t, err)
//	err = vs2.ApplyBlock(NewBlock(0))
//	require.NoError(t, err)
//
//	vs1.ApplyStateUpdate(su1)
//	vs1.ApplyStateUpdate(su2)
//	vs1.ApplyBlockIndex(0)
//
//	block := NewBlock(1, su1, su2)
//	err = vs2.ApplyBlock(block)
//	require.NoError(t, err)
//
//	require.EqualValues(t, vs1.Hash(), vs2.Hash())
//}
//
//func TestStateReader(t *testing.T) {
//	log := testlogger.NewLogger(t)
//	dbp := dbprovider.NewInMemoryDBProvider(log)
//	vs := NewVirtualState(dbp.GetPartition(nil), nil)
//	writer := vs.KVStore()
//	stateReader := NewStateReader(dbp, nil)
//	reader := stateReader.KVStoreReader()
//
//	writer.Set("key1", []byte("data1"))
//	writer.Set("key2", []byte("data2"))
//	err := vs.CommitToDb(NewOriginBlock())
//	require.NoError(t, err)
//
//	back1 := reader.MustGet("key1")
//	back2 := reader.MustGet("key2")
//	require.EqualValues(t, "data1", string(back1))
//	require.EqualValues(t, "data2", string(back2))
//}

func TestOriginHash(t *testing.T) {
	origBlock := NewOriginBlock()
	t.Logf("origin block hash = %s", origBlock.EssenceHash().String())
	require.EqualValues(t, coreutil.OriginBlockEssenceHashBase58, origBlock.EssenceHash().String())
	t.Logf("origin state hash = %s", OriginStateHash().String())
	t.Logf("zero state hash = %s", NewZeroVirtualState(mapdb.NewMapDB()).Hash().String())
	require.EqualValues(t, coreutil.OriginStateHashBase58, OriginStateHash().String())

	emptyState := NewVirtualState(mapdb.NewMapDB(), nil)
	err := emptyState.ApplyBlock(origBlock)
	require.NoError(t, err)
	require.EqualValues(t, emptyState.Hash(), OriginStateHash())
}
