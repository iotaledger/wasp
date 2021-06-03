package state

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/kv/optimism"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

// region VirtualState /////////////////////////////////////////////////

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
	sub := subRealm(db, []byte{dbkeys.ObjectTypeStateVariable})
	ret := &virtualState{
		db:        db,
		kvs:       buffered.NewBufferedKVStore(kv.NewHiveKVStoreReader(sub)),
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

// endregion ////////////////////////////////////////////////////////////

// region StateReader ///////////////////////////////////////////////////

type stateReader struct {
	db         kvstore.KVStore
	chainState kv.KVStoreReader
}

// NewOptimisticStateReader creates new reader. Checks consistency
func NewStateReader(db kvstore.KVStore) (*stateReader, error) {
	_, exists, err := loadStateHashFromDb(db)
	if err != nil {
		return nil, xerrors.Errorf("NewOptimisticStateReader: %w", err)
	}
	if !exists {
		return nil, nil
	}
	return &stateReader{
		db:         db,
		chainState: kv.NewHiveKVStoreReader(subRealm(db, []byte{dbkeys.ObjectTypeStateVariable})),
	}, nil
}

func (r *stateReader) BlockIndex() (uint32, error) {
	blockIndex, err := loadStateIndexFromState(r.chainState)
	if err != nil {
		return 0, err
	}
	return blockIndex, nil
}

func (r *stateReader) Timestamp() (time.Time, error) {
	ts, err := loadTimestampFromState(r.chainState)
	if err != nil {
		return time.Time{}, err
	}
	return ts, nil
}

func (r *stateReader) Hash() (hashing.HashValue, error) {
	hashBIn, err := r.db.Get(dbkeys.MakeKey(dbkeys.ObjectTypeStateHash))
	if err != nil {
		return [32]byte{}, err
	}
	ret, err := hashing.HashValueFromBytes(hashBIn)
	if err != nil {
		return [32]byte{}, err
	}
	return ret, nil
}

func (r *stateReader) KVStoreReader() kv.KVStoreReader {
	return r.chainState
}

// endregion ////////////////////////////////////////////////////////////

// region OptimisticStateReader ///////////////////////////////////////////////////

// state reader reads the chain state from db and validates it
type optimisticStateReader struct {
	db         kvstore.KVStore
	chainState *optimism.OptimisticKVStoreReader
}

// NewOptimisticStateReader creates new reader. Checks consistency
func NewOptimisticStateReader(db kvstore.KVStore, glb coreutil.GlobalSync) (*optimisticStateReader, error) {
	sub := subRealm(db, []byte{dbkeys.ObjectTypeStateVariable})
	chainState := kv.NewHiveKVStoreReader(sub)
	_, exists, err := loadStateHashFromDb(db)
	if err != nil {
		return nil, xerrors.Errorf("NewOptimisticStateReader: %w", err)
	}
	if !exists {
		return nil, nil
	}
	return &optimisticStateReader{
		db:         db,
		chainState: optimism.NewOptimisticKVStoreReader(chainState, glb.GetSolidIndexBaseline()),
	}, nil
}

func (r *optimisticStateReader) BlockIndex() (uint32, error) {
	blockIndex, err := loadStateIndexFromState(r.chainState)
	if err != nil {
		return 0, err
	}
	return blockIndex, nil
}

func (r *optimisticStateReader) Timestamp() (time.Time, error) {
	ts, err := loadTimestampFromState(r.chainState)
	if err != nil {
		return time.Time{}, err
	}
	return ts, nil
}

func (r *optimisticStateReader) Hash() (hashing.HashValue, error) {
	if !r.chainState.IsStateValid() {
		return [32]byte{}, optimism.ErrStateHasBeenInvalidated
	}
	hashBIn, err := r.db.Get(dbkeys.MakeKey(dbkeys.ObjectTypeStateHash))
	if err != nil {
		return [32]byte{}, err
	}
	ret, err := hashing.HashValueFromBytes(hashBIn)
	if err != nil {
		return [32]byte{}, err
	}
	if !r.chainState.IsStateValid() {
		return [32]byte{}, optimism.ErrStateHasBeenInvalidated
	}
	return ret, nil
}

func (r *optimisticStateReader) KVStoreReader() kv.KVStoreReader {
	return r.chainState
}

func (r *optimisticStateReader) SetBaseline() {
	r.chainState.SetBaseline()
}

// endregion ////////////////////////////////////////////////////////

// region helpers //////////////////////////////////////////////////

func loadStateHashFromDb(state kvstore.KVStore) (hashing.HashValue, bool, error) {
	v, err := state.Get(dbkeys.MakeKey(dbkeys.ObjectTypeStateHash))
	if err == kvstore.ErrKeyNotFound {
		return hashing.HashValue{}, false, nil
	}
	if err != nil {
		return hashing.HashValue{}, false, err
	}
	stateHash, err := hashing.HashValueFromBytes(v)
	if err != nil {
		return hashing.HashValue{}, false, err
	}
	return stateHash, true, nil
}

func loadStateIndexFromState(chainState kv.KVStoreReader) (uint32, error) {
	blockIndexBin, err := chainState.Get(kv.Key(coreutil.StatePrefixBlockIndex))
	if err != nil {
		return 0, err
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
		return time.Time{}, err
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

// endregion /////////////////////////////////////////////////////////////
