// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"errors"
	"fmt"
	"math"
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
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

// region VirtualStateAccess /////////////////////////////////////////////////

type virtualStateAccess struct {
	chainID *iscp.ChainID
	db      kvstore.KVStore
	kvs     *buffered.BufferedKVStoreAccess
}

// newVirtualState creates VirtualStateAccess interface with the partition of KVStore
func newVirtualState(db kvstore.KVStore, chainID *iscp.ChainID) *virtualStateAccess {
	sub := subRealm(db, []byte{dbkeys.ObjectTypeState})
	ret := &virtualStateAccess{
		db:  db,
		kvs: buffered.NewBufferedKVStoreAccess(kv.NewHiveKVStoreReader(sub)),
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

// calcOriginStateHash is independent of db provider nor chainID. Used for testing
func calcOriginStateHash() Commitment {
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
		chainID: vs.chainID,
		db:      vs.db,
		kvs:     vs.kvs.Copy(),
	}
	return ret
}

func (vs *virtualStateAccess) DangerouslyConvertToString() string {
	return fmt.Sprintf("#%d, ts: %v, committed hash: %s, applied block hashes: %s\n%s",
		vs.BlockIndex(),
		vs.Timestamp(),
		vs.StateCommitment().String(),
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
	vs.ApplyStateUpdate(b.(*blockImpl).stateUpdate)
}

// ApplyStateUpdates applies one state update. Doesn't change the state hash: it can be changed by Apply block
func (vs *virtualStateAccess) ApplyStateUpdate(stateUpd StateUpdate) {
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
func (vs *virtualStateAccess) StateCommitment() Commitment {
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

func loadStateHashFromDb(state kvstore.KVStore) (hashing.HashValue, bool, error) {
	v, err := state.Get(dbkeys.MakeKey(dbkeys.ObjectTypeTrie))
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
	if int(blockIndex) > math.MaxUint32 {
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
