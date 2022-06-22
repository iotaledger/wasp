// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/trie.go/trie"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"golang.org/x/xerrors"
)

// region VirtualStateAccess /////////////////////////////////////////////////

type virtualStateAccess struct {
	db   kvstore.KVStore
	kvs  *buffered.BufferedKVStoreAccess
	trie *trie.Trie
	// onBlockSave (if != nil) is called each time block is saved to the db with the state
	onBlockSave OnBlockSaveClosure
}

var _ VirtualStateAccess = &virtualStateAccess{}

// NewVirtualState creates VirtualStateAccess interface with the partition of KVStore
func NewVirtualState(db kvstore.KVStore) *virtualStateAccess { //nolint:revive
	subState := subRealm(db, []byte{dbkeys.ObjectTypeState})
	ret := &virtualStateAccess{
		db:   db,
		kvs:  buffered.NewBufferedKVStoreAccess(kv.NewHiveKVStoreReader(subState)),
		trie: NewTrie(db),
	}

	return ret
}

// CreateOriginState origin state and saves it. It assumes store is empty
func newOriginState(store kvstore.KVStore) VirtualStateAccess {
	ret := NewVirtualState(store)
	nilChainID := iscp.ChainID{}
	// state will contain chain ID at key ''. In the origin state it 'all 0'
	ret.KVStore().Set("", nilChainID.Bytes())
	ret.KVStore().Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.EncodeUint32(0))
	ret.KVStore().Set(kv.Key(coreutil.StatePrefixTimestamp), codec.EncodeTime(time.Unix(0, 0)))
	ret.Commit()
	return ret
}

// calcOriginStateCommitment is independent of db provider nor chainID. Used for testing
func calcOriginStateCommitment() trie.VCommitment {
	return trie.RootCommitment(newOriginState(mapdb.NewMapDB()).TrieNodeStore())
}

// CreateOriginState creates and saves origin state in DB
func CreateOriginState(store kvstore.KVStore, chainID *iscp.ChainID) (VirtualStateAccess, error) {
	originState := newOriginState(store)
	if err := originState.Save(); err != nil {
		return nil, err
	}
	// state will contain chain ID at key ''.
	// We set the mutation, but we do not commit yet, it will be committed with the block #1
	originState.KVStore().Set("", chainID.Bytes())
	return originState, nil
}

func subRealm(db kvstore.KVStore, realm []byte) kvstore.KVStore {
	if db == nil {
		return nil
	}
	ret, err := db.WithRealm(append(db.Realm(), realm...))
	if err != nil {
		panic(fmt.Errorf("error creating subRealm: %w", err))
	}
	return ret
}

func (vs *virtualStateAccess) WithOnBlockSave(fun OnBlockSaveClosure) {
	vs.onBlockSave = fun
}

func (vs *virtualStateAccess) Copy() VirtualStateAccess {
	ret := &virtualStateAccess{
		db:          vs.db,
		kvs:         vs.kvs.Copy(),
		trie:        vs.trie.Clone(),
		onBlockSave: vs.onBlockSave,
	}
	return ret
}

func (vs *virtualStateAccess) DangerouslyConvertToString() string {
	return fmt.Sprintf("#%d, ts: %v, state commitment: %s\n%s",
		vs.BlockIndex(),
		vs.Timestamp(),
		trie.RootCommitment(vs.TrieNodeStore()),
		vs.KVStore().DangerouslyDumpToString(),
	)
}

func (vs *virtualStateAccess) KVStore() *buffered.BufferedKVStoreAccess {
	return vs.kvs
}

func (vs *virtualStateAccess) KVStoreReader() kv.KVStoreReader {
	return vs.kvs
}

func (vs *virtualStateAccess) OptimisticStateReader(glb coreutil.ChainStateSync) OptimisticStateReader {
	return NewOptimisticStateReader(vs.db, glb)
}

func (vs *virtualStateAccess) ChainID() *iscp.ChainID {
	chainIDBin := vs.KVStoreReader().MustGet("")
	ret, err := iscp.ChainIDFromBytes(chainIDBin)
	if err != nil {
		panic(fmt.Errorf("state.ChainID: %w", err))
	}
	return ret
}

func (vs *virtualStateAccess) BlockIndex() uint32 {
	blockIndex, err := loadStateIndexFromState(vs.kvs)
	if err != nil {
		panic(fmt.Errorf("state.BlockIndex: %w", err))
	}
	return blockIndex
}

func (vs *virtualStateAccess) Timestamp() time.Time {
	ts, err := loadTimestampFromState(vs.kvs)
	if err != nil {
		panic(fmt.Errorf("state.OutputTimestamp: %w", err))
	}
	return ts
}

func (vs *virtualStateAccess) PreviousL1Commitment() *L1Commitment {
	cBin, err := vs.KVStore().Get(kv.Key(coreutil.StatePrefixPrevL1Commitment))
	if err != nil {
		panic(fmt.Errorf("state.PreviousL1Commitment: %w", err))
	}
	c, err := L1CommitmentFromBytes(cBin)
	if err != nil {
		panic(fmt.Errorf("loadPrevStateHashFromState: %w", err))
	}
	return &c
}

// ApplyBlock applies a block of state updates. Checks consistency of the block and previous state. Updates state hash
// It is not suitible for applying origin block to empty virtual state. This is done in `newZeroVirtualState`
func (vs *virtualStateAccess) ApplyBlock(b Block) error {
	if vs.BlockIndex()+1 != b.BlockIndex() {
		return fmt.Errorf("ApplyBlock: b state index #%d can't be applied to the state with index #%d",
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

// ApplyStateUpdate applies one state update
func (vs *virtualStateAccess) ApplyStateUpdate(upd Update) {
	upd.Mutations().Apply(vs.kvs)
}

func (vs *virtualStateAccess) ProofGeneric(key []byte) *trie.ProofGeneric {
	return trie.GetProofGeneric(vs.trie, dbkeys.MakeKey(dbkeys.ObjectTypeTrie, key))
}

// ExtractBlock creates a block from mutations
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

func (vs *virtualStateAccess) Commit() {
	if !vs.kvs.Mutations().IsModified() {
		return
	}
	for k, v := range vs.kvs.Mutations().Sets {
		vs.trie.Update([]byte(k), v)
	}
	for k := range vs.kvs.Mutations().Dels {
		vs.trie.Update([]byte(k), nil)
	}
	vs.trie.Commit()
	vs.kvs.Mutations().ResetModified()
}

// TrieNodeStore returns the hash of the state, calculated as a hashing of the previous (committed) state hash and the block hash.
func (vs *virtualStateAccess) TrieNodeStore() trie.NodeStore {
	return vs.trie
}

// ReconcileTrie Checks consistency of the trie with the value store. It is a heavy operation.
// Returns list of keys which cannot be proven to be consistent with the trie.
// If everything ok, it is an empty list of keys
func (vs *virtualStateAccess) ReconcileTrie() []kv.Key {
	r := vs.trie.Reconcile(valueKVStore(vs.db))
	ret := make([]kv.Key, len(r))
	for i, k := range r {
		ret[i] = kv.Key(k)
	}
	return ret
}

func loadStateIndexFromState(chainState kv.KVStoreReader) (uint32, error) {
	blockIndexBin, err := chainState.Get(kv.Key(coreutil.StatePrefixBlockIndex))
	if err != nil {
		return 0, err
	}
	if blockIndexBin == nil {
		return 0, xerrors.New("loadStateIndexFromState: not found")
	}
	blockIndex, err := codec.DecodeUint32(blockIndexBin)
	if err != nil {
		return 0, fmt.Errorf("loadStateIndexFromState: %w", err)
	}
	return blockIndex, nil
}

func loadTimestampFromState(chainState kv.KVStoreReader) (time.Time, error) {
	tsBin, err := chainState.Get(kv.Key(coreutil.StatePrefixTimestamp))
	if err != nil {
		return time.Time{}, err
	}
	ts, err := codec.DecodeTime(tsBin)
	if err != nil {
		return time.Time{}, fmt.Errorf("loadTimestampFromState: %w", err)
	}
	return ts, nil
}
