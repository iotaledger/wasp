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

type virtualState struct {
	chainID   coretypes.ChainID
	db        kvstore.KVStore
	empty     bool
	kvs       *buffered.BufferedKVStore
	stateHash hashing.HashValue
	updateLog []StateUpdate
}

// newVirtualState creates VirtualState interface with the partition of KVStore
func newVirtualState(db kvstore.KVStore, chainID *coretypes.ChainID) *virtualState {
	ret := &virtualState{
		db:        db,
		kvs:       buffered.NewBufferedKVStore(kv.NewHiveKVStoreReader(subRealm(db, []byte{dbprovider.ObjectTypeStateVariable}))),
		empty:     true,
		updateLog: make([]StateUpdate, 0),
	}
	if chainID != nil {
		ret.chainID = *chainID
	}
	return ret
}

func newZeroVirtualState(db kvstore.KVStore, chainID *coretypes.ChainID) (*virtualState, *block) {
	ret := newVirtualState(db, chainID)
	originBlock := NewOriginBlock()
	if err := ret.ApplyBlock(originBlock); err != nil {
		panic(err)
	}
	_, _ = ret.ExtractBlock() // clear the update log
	return ret, originBlock
}

// calcOriginStateHash is independent from db provider nor chainID. Used for testing
func calcOriginStateHash() hashing.HashValue {
	emptyVirtualState, _ := newZeroVirtualState(mapdb.NewMapDB(), nil)
	return emptyVirtualState.Hash()
}

func subRealm(db kvstore.KVStore, realm []byte) kvstore.KVStore {
	if db == nil {
		return nil
	}
	return db.WithRealm(append(db.Realm(), realm...))
}

func (vs *virtualState) Clone() VirtualState {
	ret := &virtualState{
		chainID:   *vs.chainID.Clone(),
		db:        vs.db,
		stateHash: vs.stateHash,
		updateLog: make([]StateUpdate, len(vs.updateLog), cap(vs.updateLog)),
		empty:     vs.empty,
		kvs:       vs.kvs.Clone(),
	}
	for i := range ret.updateLog {
		ret.updateLog[i] = vs.updateLog[i] // do not clone, just reference
	}
	return ret
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
		panic(xerrors.Errorf("state.OutputTimestamp: %v", err))
	}
	return ts
}

// ApplyBlock applies block of state updates. Checks consistency of the block and previous state. Updates state hash
func (vs *virtualState) ApplyBlock(b Block) error {
	if vs.empty && b.BlockIndex() != 0 {
		return xerrors.Errorf("ApplyBlock: b state index #%d can't be applied to the empty state", b.BlockIndex())
	}
	if !vs.empty && vs.BlockIndex()+1 != b.BlockIndex() {
		return xerrors.Errorf("ApplyBlock: b state index #%d can't be applied to the state with index #%d",
			b.BlockIndex(), vs.BlockIndex())
	}
	if !vs.empty && vs.Timestamp().After(b.Timestamp()) {
		return xerrors.New("ApplyBlock: inconsistent timestamps")
	}
	upds := make([]StateUpdate, len(b.(*block).stateUpdates))
	for i := range upds {
		upds[i] = b.(*block).stateUpdates[i]
	}
	vs.ApplyStateUpdates(upds...)
	vs.empty = false
	return nil
}

// ApplyStateUpdate applies one state update. Doesn't change state hash: it can be changed bu Apply block
func (vs *virtualState) ApplyStateUpdates(stateUpd ...StateUpdate) {
	for _, upd := range stateUpd {
		upd.Mutations().ApplyTo(vs.KVStore())
		vs.stateHash = hashing.HashData(vs.stateHash[:], upd.Bytes())
	}
	vs.updateLog = append(vs.updateLog, stateUpd...) // do not clone
}

// ExtractBlock creates a block from update log and returns it or nil if log is empty. The log is cleared
func (vs *virtualState) ExtractBlock() (Block, error) {
	if len(vs.updateLog) == 0 {
		return nil, nil
	}
	ret, err := NewBlock(vs.updateLog...)
	if err != nil {
		return nil, err
	}
	if vs.BlockIndex() != ret.BlockIndex() {
		return nil, xerrors.New("virtualState: internal inconsistency: index of the state is not equal to the index of the extracted block")
	}
	for i := range vs.updateLog {
		vs.updateLog[i] = nil // for GC
	}
	vs.updateLog = vs.updateLog[:0]
	return ret, nil
}

// TODO implement Merkle hashing

// Hash return hash of the state
func (vs *virtualState) Hash() hashing.HashValue {
	return vs.stateHash
}

type stateReader struct {
	chainPartition kvstore.KVStore
	chainState     kv.KVStoreReader
}

// NewStateReader creates new reader. Checks consistency
func NewStateReader(dbp *dbprovider.DBProvider, chainID *coretypes.ChainID) (*stateReader, error) {
	partition := dbp.GetPartition(chainID)
	stateIndex, _, exists, err := loadStateIndexAndHashFromDb(partition)
	if err != nil {
		return nil, xerrors.Errorf("NewStateReader: %w", err)
	}
	if !exists {
		return nil, xerrors.Errorf("NewStateReader: state does not exist")
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
		panic(xerrors.Errorf("stateReader.OutputTimestamp: %v", err))
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
	blockIndex, err := util.Uint64From8Bytes(blockIndexBin)
	if err != nil {
		return 0, xerrors.Errorf("loadStateIndexFromState: %w", err)
	}
	if int(blockIndex) > util.MaxUint32 {
		return 0, xerrors.Errorf("loadStateIndexFromState: wrong state index value")
	}
	return uint32(blockIndex), nil
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
