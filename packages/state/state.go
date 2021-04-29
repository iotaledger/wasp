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
	variables *buffered.BufferedKVStore
	stateHash hashing.HashValue
}

func NewVirtualState(db kvstore.KVStore, chainID *coretypes.ChainID) *virtualState {
	ret := &virtualState{
		db:        db,
		variables: buffered.NewBufferedKVStore(kv.NewHiveKVStoreReader(subRealm(db, []byte{dbprovider.ObjectTypeStateVariable}))),
		empty:     true,
	}
	if chainID != nil {
		ret.chainID = *chainID
	}
	return ret
}

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
		chainID:   vs.chainID,
		db:        vs.db,
		stateHash: vs.stateHash,
		empty:     vs.empty,
		variables: vs.variables.Clone(),
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
	return vs.variables
}

func (vs *virtualState) KVStoreReader() kv.KVStoreReader {
	return vs.variables
}

func (vs *virtualState) BlockIndex() uint32 {
	blockIndexBin, err := vs.variables.Get(kv.Key(coreutil.StatePrefixBlockIndex))
	if err != nil {
		panic(err)
	}
	blockIndex, err := util.Uint32From4Bytes(blockIndexBin)
	if err != nil {
		panic(err)
	}
	return blockIndex
}

func (vs *virtualState) Timestamp() time.Time {
	tsBin, err := vs.variables.Get(kv.Key(coreutil.StatePrefixTimestamp))
	if err != nil {
		panic(err)
	}
	ts, ok, err := codec.DecodeTime(tsBin)
	if err != nil {
		panic(err)
	}
	if !ok {
		panic(xerrors.New("VistualState: timestamp not found"))
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
	for k, v := range vs.variables.Mutations().Sets {
		if err := batch.Set(dbkeyStateVariable(k), v); err != nil {
			return err
		}
	}
	for k := range vs.variables.Mutations().Dels {
		if err := batch.Delete(dbkeyStateVariable(k)); err != nil {
			return err
		}
	}

	if err := batch.Commit(); err != nil {
		return err
	}
	vs.variables.ClearMutations()
	return nil
}

// endregion ////////////////////////////////////////////////////////////////////

func LoadSolidState(dbp *dbprovider.DBProvider, chainID *coretypes.ChainID) (VirtualState, bool, error) {
	partition := dbp.GetPartition(chainID)

	var stateHash hashing.HashValue
	var stateIndex uint32
	{
		// read state hash
		v, err := partition.Get(dbprovider.MakeKey(dbprovider.ObjectTypeStateHash))
		if err == kvstore.ErrKeyNotFound {
			// state does not exist
			return nil, false, nil
		}
		if err != nil {
			return nil, false, err
		}
		stateHash, err = hashing.HashValueFromBytes(v)
		if err != nil {
			return nil, false, err
		}
	}
	{
		// read state index
		v, err := partition.Get(dbprovider.MakeKey(dbprovider.ObjectTypeStateIndex))
		if err != nil {
			return nil, false, err
		}
		stateIndex, err = util.Uint32From4Bytes(v)
		if err != nil {
			return nil, false, err
		}
	}
	vs := NewVirtualState(partition, chainID)
	if vs.BlockIndex() != stateIndex {
		return nil, false, xerrors.New("LoadSolidState: state index inconsistent with state")
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

func NewStateReader(dbp *dbprovider.DBProvider, chainID *coretypes.ChainID) *stateReader {
	partition := dbp.GetPartition(chainID)
	return &stateReader{
		chainPartition: partition,
		chainState:     kv.NewHiveKVStoreReader(subRealm(partition, []byte{dbprovider.ObjectTypeStateVariable})),
	}
}

func (r *stateReader) BlockIndex() uint32 {
	blockIndex, err := getBlockIndexFromState(r.chainState)
	if err != nil {
		panic(xerrors.Errorf("stateReader.BlockIndex: %v", err))
	}
	return blockIndex
}

func (r *stateReader) Timestamp() time.Time {
	ts, exist, err := getTimestampFromState(r.chainState)
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

func getBlockIndexFromState(chainState kv.KVStoreReader) (uint32, bool, error) {
	blockIndexBin, err := chainState.Get(kv.Key(coreutil.StatePrefixBlockIndex))
	if err == kvstore.ErrKeyNotFound {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	if blockIndexBin == nil {
		return 0, false, nil
	}
	blockIndex, err := util.Uint32From4Bytes(blockIndexBin)
	if err != nil {
		return 0, false, err
	}
	return blockIndex, true, nil
}

func getTimestampFromState(chainState kv.KVStoreReader) (time.Time, bool, error) {
	tsBin, err := chainState.Get(kv.Key(coreutil.StatePrefixTimestamp))
	if err == kvstore.ErrKeyNotFound {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, err
	}
	ts, ok, err := codec.DecodeTime(tsBin)
	if err != nil {
		return time.Time{}, false, err
	}
	if !ok {
		return time.Time{}, false, xerrors.New("stateReader: timestamp not found")
	}
	return ts, true, nil
}

// endregion ////////////////////////////////////////////////////////////////////
