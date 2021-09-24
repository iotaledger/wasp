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

// region VirtualState /////////////////////////////////////////////////

type virtualState struct {
	chainID             iscp.ChainID
	db                  kvstore.KVStore
	empty               bool
	kvs                 *buffered.BufferedKVStore
	stateHash           hashing.HashValue
	stateUpdate         StateUpdate
	isStateHashOutdated bool
}

// newVirtualState creates VirtualState interface with the partition of KVStore
func newVirtualState(db kvstore.KVStore, chainID *iscp.ChainID) *virtualState {
	sub := subRealm(db, []byte{dbkeys.ObjectTypeStateVariable})
	ret := &virtualState{
		db:                  db,
		kvs:                 buffered.NewBufferedKVStore(kv.NewHiveKVStoreReader(sub)),
		empty:               true,
		stateUpdate:         NewStateUpdate(),
		isStateHashOutdated: true,
	}
	if chainID != nil {
		ret.chainID = *chainID
	}
	return ret
}

func newZeroVirtualState(db kvstore.KVStore, chainID *iscp.ChainID) (VirtualState, Block) {
	ret := newVirtualState(db, chainID)
	originBlock := newOriginBlock()
	if err := ret.ApplyBlock(originBlock); err != nil {
		panic(err)
	}
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

func (vs *virtualState) Clone() VirtualState {
	ret := &virtualState{
		chainID:     *vs.chainID.Clone(),
		db:          vs.db,
		stateHash:   vs.stateHash,
		stateUpdate: vs.stateUpdate.Clone(),
		empty:       vs.empty,
		kvs:         vs.kvs.Clone(),
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
		panic(xerrors.Errorf("state.BlockIndex: %w", err))
	}
	return blockIndex
}

func (vs *virtualState) Timestamp() time.Time {
	ts, err := loadTimestampFromState(vs.kvs)
	if err != nil {
		panic(xerrors.Errorf("state.OutputTimestamp: %w", err))
	}
	return ts
}

func (vs *virtualState) PreviousStateHash() hashing.HashValue {
	ph, err := loadPrevStateHashFromState(vs.kvs)
	if err != nil {
		panic(xerrors.Errorf("state.PreviousStateHash: %w", err))
	}
	return ph
}

// ApplyBlock applies a block of state updates. Checks consistency of the block and previous state. Updates state hash
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
	vs.ApplyStateUpdates(b.(*blockImpl).stateUpdate)
	vs.empty = false
	return nil
}

// ApplyStateUpdate applies one state update. Doesn't change the state hash: it can be changed by Apply block
func (vs *virtualState) ApplyStateUpdates(stateUpd ...StateUpdate) {
	for _, upd := range stateUpd {
		upd.Mutations().ApplyTo(vs.KVStore())
		for k, v := range upd.Mutations().Sets {
			vs.stateUpdate.Mutations().Set(k, v)
		}
		for k := range upd.Mutations().Dels {
			vs.stateUpdate.Mutations().Del(k)
		}
	}
	vs.isStateHashOutdated = true
}

// ExtractBlock creates a block from update log and returns it or nil if log is empty. The log is cleared
func (vs *virtualState) ExtractBlock() (Block, error) {
	ret, err := newBlock(vs.stateUpdate)
	if err != nil {
		return nil, err
	}
	if vs.BlockIndex() != ret.BlockIndex() {
		return nil, xerrors.New("virtualState: internal inconsistency: index of the state is not equal to the index of the extracted block")
	}
	return ret, nil
}

// StateCommitment returns the hash of the state, calculated as a recursive hashing of the previous state hash and the block.
func (vs *virtualState) StateCommitment() hashing.HashValue {
	if vs.isStateHashOutdated {
		block, err := vs.ExtractBlock()
		if err != nil {
			panic(xerrors.Errorf("StateCommitment: %v", err))
		}
		vs.stateHash = hashing.HashData(vs.stateHash[:], block.Bytes())
		vs.isStateHashOutdated = false
	}
	return vs.stateHash
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

func (r *OptimisticStateReaderImpl) KVStoreReader() kv.KVStoreReader {
	return r.chainState
}

func (r *OptimisticStateReaderImpl) SetBaseline() {
	r.chainState.SetBaseline()
}

// endregion ////////////////////////////////////////////////////////

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
	ts, ok, err := codec.DecodeTime(tsBin)
	if err != nil {
		return time.Time{}, xerrors.Errorf("loadTimestampFromState: %w", err)
	}
	if !ok {
		return time.Time{}, xerrors.New("loadTimestampFromState: timestamp not found")
	}
	return ts, nil
}

func loadPrevStateHashFromState(chainState kv.KVStoreReader) (hashing.HashValue, error) {
	hashBin, err := chainState.Get(kv.Key(coreutil.StatePrefixPrevStateHash))
	if err != nil {
		return hashing.NilHash, err
	}
	ph, ok, err := codec.DecodeHashValue(hashBin)
	if err != nil {
		return hashing.NilHash, xerrors.Errorf("loadPrevStateHashFromState: %w", err)
	}
	if !ok {
		return hashing.NilHash, xerrors.New("loadPrevStateHashFromState: previous state hash not found")
	}
	return ph, nil
}

// endregion /////////////////////////////////////////////////////////////
