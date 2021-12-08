// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"errors"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/optimism"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

// region VirtualStateAccess /////////////////////////////////////////////////

type virtualStateAccess struct {
	chainID            *iscp.ChainID
	db                 kvstore.KVStore
	kvs                *buffered.BufferedKVStoreAccess
	committedHash      hashing.HashValue
	appliedBlockHashes []hashing.HashValue
}

// newVirtualState creates VirtualStateAccess interface with the partition of KVStore
func newVirtualState(db kvstore.KVStore, chainID *iscp.ChainID) *virtualStateAccess {
	sub := subRealm(db, []byte{dbkeys.ObjectTypeStateVariable})
	ret := &virtualStateAccess{
		db:                 db,
		kvs:                buffered.NewBufferedKVStoreAccess(kv.NewHiveKVStoreReader(sub)),
		appliedBlockHashes: make([]hashing.HashValue, 0),
	}
	if chainID != nil {
		ret.chainID = chainID
	}
	return ret
}

func newZeroVirtualState(db kvstore.KVStore, chainID *iscp.ChainID) (VirtualStateAccess, Block) {
	ret := newVirtualState(db, chainID)
	originBlock := newOriginBlock()
	ret.applyBlockNoCheck(originBlock)
	_, _ = ret.ExtractBlock() // clear the update log
	return ret, originBlock
}

// calcOriginStateHash is independent from db provider nor chainID. Used for testing
func calcOriginStateHash() hashing.HashValue {
	emptyVirtualState, _ := newZeroVirtualState(mapdb.NewMapDB(), nil)
	return emptyVirtualState.StateCommitment()
}

func subRealm(db kvstore.KVStore, realm []byte) kvstore.KVStore {
	if db == nil {
		return nil
	}
	return db.WithRealm(append(db.Realm(), realm...))
}

func (vs *virtualStateAccess) Copy() VirtualStateAccess {
	ret := &virtualStateAccess{
		chainID:            vs.chainID.Clone(),
		db:                 vs.db,
		committedHash:      vs.committedHash,
		appliedBlockHashes: make([]hashing.HashValue, len(vs.appliedBlockHashes)),
		kvs:                vs.kvs.Copy(),
	}
	copy(ret.appliedBlockHashes, vs.appliedBlockHashes)
	return ret
}

func (vs *virtualStateAccess) DangerouslyConvertToString() string {
	blockHashes := "["
	for _, blockHash := range vs.appliedBlockHashes {
		blockHashes += blockHash.String() + ", "
	}
	blockHashes += "]"
	return fmt.Sprintf("#%d, ts: %v, committed hash: %s, applied block hashes: %s\n%s",
		vs.BlockIndex(),
		vs.Timestamp(),
		vs.committedHash.String(),
		blockHashes,
		vs.KVStore().DangerouslyDumpToString(),
	)
}

func (vs *virtualStateAccess) KVStore() *buffered.BufferedKVStoreAccess {
	return vs.kvs
}

func (vs *virtualStateAccess) KVStoreReader() kv.KVStoreReader {
	return vs.kvs
}

func (vs *virtualStateAccess) BlockIndex() uint32 {
	blockIndex, err := loadStateIndexFromState(vs.kvs)
	if err != nil {
		panic(xerrors.Errorf("state.BlockIndex: %w", err))
	}
	return blockIndex
}

func (vs *virtualStateAccess) Timestamp() time.Time {
	ts, err := loadTimestampFromState(vs.kvs)
	if err != nil {
		panic(xerrors.Errorf("state.OutputTimestamp: %w", err))
	}
	return ts
}

func (vs *virtualStateAccess) PreviousStateHash() hashing.HashValue {
	ph, err := loadPrevStateHashFromState(vs.kvs)
	if err != nil {
		panic(xerrors.Errorf("state.PreviousStateHash: %w", err))
	}
	return ph
}

// ApplyBlock applies a block of state updates. Checks consistency of the block and previous state. Updates state hash
// It is not suitible for applying origin block to empty virtual state. This is done in `newZeroVirtualState`
func (vs *virtualStateAccess) ApplyBlock(b Block) error {
	if vs.BlockIndex()+1 != b.BlockIndex() {
		return xerrors.Errorf("ApplyBlock: b state index #%d can't be applied to the state with index #%d",
			b.BlockIndex(), vs.BlockIndex())
	}
	if vs.Timestamp().After(b.Timestamp()) {
		return xerrors.New("ApplyBlock: inconsistent timestamps")
	}
	vs.applyBlockNoCheck(b)
	return nil
}

func (vs *virtualStateAccess) applyBlockNoCheck(b Block) {
	vs.ApplyStateUpdates(b.(*blockImpl).stateUpdate)
	vs.appliedBlockHashes = append(vs.appliedBlockHashes, hashing.HashData(b.EssenceBytes()))
}

// ApplyStateUpdates applies one state update. Doesn't change the state hash: it can be changed by Apply block
func (vs *virtualStateAccess) ApplyStateUpdates(stateUpd ...StateUpdate) {
	for _, upd := range stateUpd {
		upd.Mutations().ApplyTo(vs.KVStore())
		for k, v := range upd.Mutations().Sets {
			vs.kvs.Mutations().Set(k, v)
		}
		for k := range upd.Mutations().Dels {
			vs.kvs.Mutations().Del(k)
		}
	}
}

// ExtractBlock creates a block from update log and returns it or nil if log is empty. The log is cleared
func (vs *virtualStateAccess) ExtractBlock() (Block, error) {
	ret, err := newBlock(vs.kvs.Mutations())
	if err != nil {
		return nil, err
	}
	if vs.BlockIndex() != ret.BlockIndex() {
		return nil, xerrors.New("virtualStateAccess: internal inconsistency: index of the state is not equal to the index of the extracted block")
	}
	return ret, nil
}

// StateCommitment returns the hash of the state, calculated as a hashing of the previous (committed) state hash and the block hash.
func (vs *virtualStateAccess) StateCommitment() hashing.HashValue {
	if vs.kvs.Mutations().IsEmpty() {
		return vs.committedHash
	}
	if len(vs.appliedBlockHashes) == 0 {
		block, err := vs.ExtractBlock()
		if err != nil {
			panic(xerrors.Errorf("StateCommitment: %v", err))
		}
		vs.appliedBlockHashes = append(vs.appliedBlockHashes, hashing.HashData(block.EssenceBytes()))
	}
	ret := vs.committedHash
	for i := range vs.appliedBlockHashes {
		ret = hashing.HashData(ret[:], vs.appliedBlockHashes[i][:])
	}
	return ret
}

// endregion ////////////////////////////////////////////////////////////

// region OptimisticStateReader ///////////////////////////////////////////////////

// state reader reads the chain state from db and validates it
type OptimisticStateReaderImpl struct {
	db         kvstore.KVStore
	chainState *optimism.OptimisticKVStoreReader
}

// NewOptimisticStateReader creates new optimistic read-only access to the database. It contains own read baseline
func NewOptimisticStateReader(db kvstore.KVStore, glb coreutil.ChainStateSync) *OptimisticStateReaderImpl {
	chainState := kv.NewHiveKVStoreReader(subRealm(db, []byte{dbkeys.ObjectTypeStateVariable}))
	return &OptimisticStateReaderImpl{
		db:         db,
		chainState: optimism.NewOptimisticKVStoreReader(chainState, glb.GetSolidIndexBaseline()),
	}
}

func (r *OptimisticStateReaderImpl) BlockIndex() (uint32, error) {
	blockIndex, err := loadStateIndexFromState(r.chainState)
	if err != nil {
		return 0, err
	}
	return blockIndex, nil
}

func (r *OptimisticStateReaderImpl) Timestamp() (time.Time, error) {
	ts, err := loadTimestampFromState(r.chainState)
	if err != nil {
		return time.Time{}, err
	}
	return ts, nil
}

func (r *OptimisticStateReaderImpl) Hash() (hashing.HashValue, error) {
	if !r.chainState.IsStateValid() {
		return [32]byte{}, coreutil.ErrorStateInvalidated
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
		return [32]byte{}, coreutil.ErrorStateInvalidated
	}
	return ret, nil
}

func (r *OptimisticStateReaderImpl) KVStoreReader() kv.KVStoreReader {
	return r.chainState
}

func (r *OptimisticStateReaderImpl) SetBaseline() {
	r.chainState.SetBaseline()
}

// endregion ////////////////////////////////////////////////////////

// region mustOptimisticVirtualStateAccess ////////////////////////////////

// MustOptimisticVirtualState is a virtual state wrapper with global state baseline
// Once baseline is invalidated globally any subsequent access to the mustOptimisticVirtualStateAccess
// will lead to panic(coreutil.ErrorStateInvalidated)
type mustOptimisticVirtualStateAccess struct {
	state    VirtualStateAccess
	baseline coreutil.StateBaseline
}

// WrapMustOptimisticVirtualStateAccess wraps virtual state with state baseline in on object
// Does not copy buffers
func WrapMustOptimisticVirtualStateAccess(state VirtualStateAccess, baseline coreutil.StateBaseline) VirtualStateAccess {
	return &mustOptimisticVirtualStateAccess{
		state:    state,
		baseline: baseline,
	}
}

func (s *mustOptimisticVirtualStateAccess) BlockIndex() uint32 {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.BlockIndex()
}

func (s *mustOptimisticVirtualStateAccess) Timestamp() time.Time {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.Timestamp()
}

func (s *mustOptimisticVirtualStateAccess) PreviousStateHash() hashing.HashValue {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.PreviousStateHash()
}

func (s *mustOptimisticVirtualStateAccess) StateCommitment() hashing.HashValue {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.StateCommitment()
}

func (s *mustOptimisticVirtualStateAccess) KVStoreReader() kv.KVStoreReader {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.KVStoreReader()
}

func (s *mustOptimisticVirtualStateAccess) ApplyStateUpdates(upd ...StateUpdate) {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	s.state.ApplyStateUpdates(upd...)
}

func (s *mustOptimisticVirtualStateAccess) ApplyBlock(block Block) error {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.ApplyBlock(block)
}

func (s *mustOptimisticVirtualStateAccess) ExtractBlock() (Block, error) {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.ExtractBlock()
}

func (s *mustOptimisticVirtualStateAccess) Commit(blocks ...Block) error {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.Commit(blocks...)
}

func (s *mustOptimisticVirtualStateAccess) KVStore() *buffered.BufferedKVStoreAccess {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.KVStore()
}

func (s *mustOptimisticVirtualStateAccess) Copy() VirtualStateAccess {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.Copy()
}

func (s *mustOptimisticVirtualStateAccess) DangerouslyConvertToString() string {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.DangerouslyConvertToString()
}

// endregion /////////////////////////////////////

// region helpers //////////////////////////////////////////////////

func loadStateHashFromDb(state kvstore.KVStore) (hashing.HashValue, bool, error) {
	v, err := state.Get(dbkeys.MakeKey(dbkeys.ObjectTypeStateHash))
	if errors.Is(err, kvstore.ErrKeyNotFound) {
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
	ts, err := codec.DecodeTime(tsBin)
	if err != nil {
		return time.Time{}, xerrors.Errorf("loadTimestampFromState: %w", err)
	}
	return ts, nil
}

func loadPrevStateHashFromState(chainState kv.KVStoreReader) (hashing.HashValue, error) {
	hashBin, err := chainState.Get(kv.Key(coreutil.StatePrefixPrevStateHash))
	if err != nil {
		return hashing.NilHash, err
	}
	ph, err := codec.DecodeHashValue(hashBin)
	if err != nil {
		return hashing.NilHash, xerrors.Errorf("loadPrevStateHashFromState: %w", err)
	}
	return ph, nil
}

// endregion /////////////////////////////////////////////////////////////
