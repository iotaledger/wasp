package state

import (
	"fmt"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
	"time"
)

// region virtualState //////////////////////////////////////////////////////////////////////

type virtualState struct {
	chainID   coretypes.ChainID
	db        kvstore.KVStore
	empty     bool
	kvs       *buffered.BufferedKVStore
	stateHash hashing.HashValue
}

// NewVirtualState creates VirtualState interface with the partition of KVStore
func NewVirtualState(db kvstore.KVStore, chainID *coretypes.ChainID) *virtualState {
	ret := &virtualState{
		db:    db,
		kvs:   buffered.NewBufferedKVStore(kv.NewHiveKVStoreReader(subRealm(db, []byte{dbprovider.ObjectTypeStateVariable}))),
		empty: true,
	}
	if chainID != nil {
		ret.chainID = *chainID
	}
	return ret
}

// NewZeroVirtualState creates zero state which is the minimal consistent state.
// It is not committed it is an origin state. It has statically known hash coreutils.OriginStateHashBase58
func NewZeroVirtualState(db kvstore.KVStore) *virtualState {
	ret := NewVirtualState(db, nil)
	if err := ret.ApplyBlock(NewOriginBlock()); err != nil {
		panic(err)
	}
	return ret
}

// OriginStateHash is independent from db provider nor chainID
func OriginStateHash() hashing.HashValue {
	emptyVirtualState := NewZeroVirtualState(mapdb.NewMapDB())
	return emptyVirtualState.Hash()
}

func subRealm(db kvstore.KVStore, realm []byte) kvstore.KVStore {
	if db == nil {
		return nil
	}
	return db.WithRealm(append(db.Realm(), realm...))
}

func (vs *virtualState) Clone() VirtualState {
	return &virtualState{
		chainID:   *vs.chainID.Clone(),
		db:        vs.db,
		stateHash: vs.stateHash,
		empty:     vs.empty,
		kvs:       vs.kvs.Clone(),
	}
}

func (vs *virtualState) DangerouslyConvertToString() string {
	return fmt.Sprintf("#%d, ts: %v, hash, %s\n%s",
		vs.BlockIndex(),
		vs.Timestamp(),
		vs.stateHash.String(),
		vs.KVStore().DangerouslyDumpToString(),
	)
}

func (vs *virtualState) KVStore() *buffered.BufferedKVStore {
	return vs.kvs
}

func (vs *virtualState) KVStoreReader() kv.KVStoreReader {
	return vs.kvs
}

func (vs *virtualState) BlockIndex() uint32 {
	blockIndex, err := loadStateIndexFromState(vs.kvs)
	if err != nil {
		panic(xerrors.Errorf("state.BlockIndex: %v", err))
	}
	return blockIndex
}

func (vs *virtualState) Timestamp() time.Time {
	ts, err := loadTimestampFromState(vs.kvs)
	if err != nil {
		panic(xerrors.Errorf("state.Timestamp: %v", err))
	}
	return ts
}

// ApplyBlock applies block of state updates
func (vs *virtualState) ApplyBlock(block Block) error {
	if vs.empty && block.BlockIndex() != 0 {
		return xerrors.Errorf("ApplyBlock: block state index #%d can't be applied to the empty state", block.BlockIndex())
	}
	if vs.BlockIndex()+1 != block.BlockIndex() {
		return xerrors.Errorf("ApplyBlock: block state index #%d can't be applied to the state with index #%d",
			block.BlockIndex(), vs.BlockIndex())
	}
	block.ForEach(func(_ uint16, stateUpd StateUpdate) bool {
		vs.ApplyStateUpdate(stateUpd)
		return true
	})
	return nil
}

// applies one state update. Doesn't change state index
func (vs *virtualState) ApplyStateUpdate(stateUpd StateUpdate) {
	stateUpd.Mutations().ApplyTo(vs.KVStore())
	vh := vs.Hash()
	sh := stateUpd.Hash()
	vs.stateHash = hashing.HashData(vh[:], sh[:])
	vs.empty = false
}

// Hash return hash of the state
// TODO implement Merkle hashing
func (vs *virtualState) Hash() hashing.HashValue {
	return vs.stateHash
}

// saves variable state to db atomically with the block of state updates and records of processed requests
func (vs *virtualState) CommitToDb(b Block) error {
	batch := vs.db.Batched()

	{
		if err := batch.Set(dbkeyBlock(b.BlockIndex()), b.Bytes()); err != nil {
			return err
		}
	}
	{
		if err := batch.Set(dbprovider.MakeKey(dbprovider.ObjectTypeStateHash), vs.Hash().Bytes()); err != nil {
			return err
		}
		if err := batch.Set(dbprovider.MakeKey(dbprovider.ObjectTypeStateIndex), util.Uint32To4Bytes(vs.BlockIndex())); err != nil {
			return err
		}
	}

	// store mutations
	for k, v := range vs.kvs.Mutations().Sets {
		if err := batch.Set(dbkeyStateVariable(k), v); err != nil {
			return err
		}
	}
	for k := range vs.kvs.Mutations().Dels {
		if err := batch.Delete(dbkeyStateVariable(k)); err != nil {
			return err
		}
	}

	if err := batch.Commit(); err != nil {
		return err
	}
	vs.kvs.ClearMutations()
	return nil
}

// endregion ////////////////////////////////////////////////////////////////////

// LoadSolidState establishes VirtualState interface with the solid state in DB.
// Checks consistency of DB
func LoadSolidState(dbp *dbprovider.DBProvider, chainID *coretypes.ChainID) (VirtualState, bool, error) {
	partition := dbp.GetPartition(chainID)

	stateIndex, stateHash, exists, err := loadStateIndexAndHashFromDb(partition)
	if err != nil || !exists {
		return nil, exists, xerrors.Errorf("LoadSolidState: %w", err)
	}
	vs := NewVirtualState(partition, chainID)
	stateIndex1, err := loadStateIndexFromState(vs.KVStoreReader())
	if err != nil {
		return nil, false, xerrors.Errorf("LoadSolidState: %w", err)
	}
	if stateIndex != stateIndex1 {
		return nil, false, xerrors.New("LoadSolidState: state index inconsistent with the state")
	}
	vs.stateHash = stateHash
	return vs, true, nil
}

func dbkeyStateVariable(key kv.Key) []byte {
	return dbprovider.MakeKey(dbprovider.ObjectTypeStateVariable, []byte(key))
}

// region stateReader /////////////////////////////////////////////////////////

type stateReader struct {
	chainPartition kvstore.KVStore
	chainState     kv.KVStoreReader
}

func NewStateReader(dbp *dbprovider.DBProvider, chainID *coretypes.ChainID) (*stateReader, error) {
	partition := dbp.GetPartition(chainID)

	stateIndex, _, exists, err := loadStateIndexAndHashFromDb(partition)
	if err != nil || !exists {
		return nil, xerrors.Errorf("NewStateReader: %w", err)
	}
	ret := &stateReader{
		chainPartition: partition,
		chainState:     kv.NewHiveKVStoreReader(subRealm(partition, []byte{dbprovider.ObjectTypeStateVariable})),
	}
	stateIndex1, err := loadStateIndexFromState(ret.chainState)
	if err != nil {
		return nil, xerrors.Errorf("NewStateReader: %w", err)
	}
	if stateIndex != stateIndex1 {
		return nil, xerrors.New("NewStateReader: state index inconsistent with the state")
	}
	return ret, nil
}

func (r *stateReader) BlockIndex() uint32 {
	blockIndex, err := loadStateIndexFromState(r.chainState)
	if err != nil {
		panic(xerrors.Errorf("stateReader.BlockIndex: %v", err))
	}
	return blockIndex
}

func (r *stateReader) Timestamp() time.Time {
	ts, err := loadTimestampFromState(r.chainState)
	if err != nil {
		panic(xerrors.Errorf("stateReader.Timestamp: %v", err))
	}
	return ts
}

func (r *stateReader) Hash() hashing.HashValue {
	hashBIn, err := r.chainPartition.Get(dbprovider.MakeKey(dbprovider.ObjectTypeStateHash))
	if err != nil {
		panic(err)
	}
	ret, err := hashing.HashValueFromBytes(hashBIn)
	if err != nil {
		panic(err)
	}
	return ret
}

func (r *stateReader) KVStoreReader() kv.KVStoreReader {
	return r.chainState
}

// endregion ///////////////////////////////////////////////////////////////////

// region util ///////////////////////////////////////////////////////////////////
func loadStateIndexAndHashFromDb(partition kvstore.KVStore) (uint32, hashing.HashValue, bool, error) {
	v, err := partition.Get(dbprovider.MakeKey(dbprovider.ObjectTypeStateHash))
	if err == kvstore.ErrKeyNotFound {
		return 0, hashing.HashValue{}, false, nil
	}
	if err != nil {
		return 0, hashing.HashValue{}, false, err
	}
	stateHash, err := hashing.HashValueFromBytes(v)
	if err != nil {
		return 0, hashing.HashValue{}, false, err
	}
	v, err = partition.Get(dbprovider.MakeKey(dbprovider.ObjectTypeStateIndex))
	if err == kvstore.ErrKeyNotFound {
		return 0, hashing.HashValue{}, false, nil
	}
	if err != nil {
		return 0, hashing.HashValue{}, false, err
	}
	stateIndex, err := util.Uint32From4Bytes(v)
	if err != nil {
		return 0, hashing.HashValue{}, false, err
	}
	return stateIndex, stateHash, true, nil
}

func loadStateIndexFromState(chainState kv.KVStoreReader) (uint32, error) {
	blockIndexBin, err := chainState.Get(kv.Key(coreutil.StatePrefixBlockIndex))
	if err != nil {
		return 0, xerrors.Errorf("loadStateIndexFromState: %w", err)
	}
	if blockIndexBin == nil {
		return 0, xerrors.New("loadStateIndexFromState: not found")
	}
	blockIndex, err := util.Uint32From4Bytes(blockIndexBin)
	if err != nil {
		return 0, xerrors.Errorf("loadStateIndexFromState: %w", err)
	}
	return blockIndex, nil
}

func loadTimestampFromState(chainState kv.KVStoreReader) (time.Time, error) {
	tsBin, err := chainState.Get(kv.Key(coreutil.StatePrefixTimestamp))
	if err != nil {
		return time.Time{}, xerrors.Errorf("loadTimestampFromState: %w", err)
	}
	ts, ok, err := codec.DecodeTime(tsBin)
	if err != nil {
		return time.Time{}, xerrors.Errorf("loadTimestampFromState: %w", err)
	}
	if !ok {
		return time.Time{}, xerrors.New("loadTimestampFromState: timestamp not found")
	}
	return ts, nil
}

// endregion ////////////////////////////////////////////////////////////////////
