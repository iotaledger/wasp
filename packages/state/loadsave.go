package state

import (
	"errors"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/kv/trie_merkle"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

const maxKeyLen = 4096

type mustKVStoreBatch struct {
	batch kvstore.BatchedMutations
}

func (k *mustKVStoreBatch) Set(key kv.Key, value []byte) {
	if err := k.batch.Set(dbkeys.MakeKey(dbkeys.ObjectTypeTrie, []byte(key)), value); err != nil {
		panic(err)
	}
}

func (k *mustKVStoreBatch) Del(key kv.Key) {
	if err := k.batch.Delete(dbkeys.MakeKey(dbkeys.ObjectTypeTrie, []byte(key))); err != nil {
		panic(err)
	}
}

// Commit saves updates collected in the virtual state to DB together with the provided blocks in one transaction
// Mutations must be non-empty otherwise it is NOP
func (vs *virtualStateAccess) Save(blocks ...Block) error {
	if vs.kvs.Mutations().IsEmpty() {
		// nothing to commit
		return nil
	}
	// generate trie update
	sub := subRealm(vs.db, []byte{dbkeys.ObjectTypeTrie})
	stateTrie := trie.New(trie_merkle.CommitmentLogic, kv.NewHiveKVStoreReader(sub))
	for k, v := range vs.kvs.Mutations().Sets {
		stateTrie.Update([]byte(k), v)
	}
	for k := range vs.kvs.Mutations().Dels {
		stateTrie.Update([]byte(k), nil)
	}
	stateTrie.Commit()

	batch := vs.db.Batched()
	stateTrie.FlushCache(&mustKVStoreBatch{batch})

	// TODO ----------------------------------

	stateCommitment := vs.StateCommitment()
	if err := batch.Set(dbkeys.MakeKey(dbkeys.ObjectTypeTrie), stateCommitment.Bytes()); err != nil {
		return err
	}

	for _, blk := range blocks {
		key := dbkeys.MakeKey(dbkeys.ObjectTypeBlock, util.Uint32To4Bytes(blk.BlockIndex()))
		if err := batch.Set(key, blk.Bytes()); err != nil {
			return err
		}
	}

	// store mutations
	for k, v := range vs.kvs.Mutations().Sets {
		if err := batch.Set(dbkeys.MakeKey(dbkeys.ObjectTypeState, []byte(k)), v); err != nil {
			return err
		}
	}
	for k := range vs.kvs.Mutations().Dels {
		if err := batch.Delete(dbkeys.MakeKey(dbkeys.ObjectTypeState, []byte(k))); err != nil {
			return err
		}
	}

	if err := batch.Commit(); err != nil {
		return err
	}

	// call flush explicitly, because batched.Commit doesn't actually write the changes to disk.
	if err := vs.db.Flush(); err != nil {
		return err
	}

	vs.kvs.ClearMutations()
	vs.kvs.Mutations().ResetModified()
	vs.appliedBlockHashes = vs.appliedBlockHashes[:0]
	vs.committedHash = stateCommitment
	return nil
}

// CreateOriginState creates zero state which is the minimal consistent state.
// It is not committed it is an origin state. It has statically known hash coreutils.OriginStateHashBase58
func CreateOriginState(store kvstore.KVStore, chainID *iscp.ChainID) (VirtualStateAccess, error) {
	originState, originBlock := newZeroVirtualState(store, chainID)
	if err := originState.Save(originBlock); err != nil {
		return nil, err
	}
	return originState, nil
}

// LoadSolidState establishes VirtualStateAccess interface with the solid state in DB. Checks consistency of DB
func LoadSolidState(store kvstore.KVStore, chainID *iscp.ChainID) (VirtualStateAccess, bool, error) {
	stateHash, exists, err := loadStateHashFromDb(store)
	if err != nil {
		return nil, exists, xerrors.Errorf("LoadSolidState: %w", err)
	}
	if !exists {
		return nil, false, nil
	}
	vs := newVirtualState(store, chainID)
	vs.committedHash = stateHash

	vs.kvs.Mutations().ResetModified()
	return vs, true, nil
}

// LoadBlockBytes loads block bytes of the specified block index from DB
func LoadBlockBytes(store kvstore.KVStore, stateIndex uint32) ([]byte, error) {
	data, err := store.Get(dbkeys.MakeKey(dbkeys.ObjectTypeBlock, util.Uint32To4Bytes(stateIndex)))
	if errors.Is(err, kvstore.ErrKeyNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

// LoadBlock loads block from DB and decodes it
func LoadBlock(store kvstore.KVStore, stateIndex uint32) (Block, error) {
	data, err := LoadBlockBytes(store, stateIndex)
	if err != nil {
		return nil, err
	}
	return BlockFromBytes(data)
}
