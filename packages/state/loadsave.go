package state

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

type mustKVStoreBatch struct {
	prefix byte
	batch  kvstore.BatchedMutations
}

func newKVStoreBatch(prefix byte, batch kvstore.BatchedMutations) *mustKVStoreBatch {
	return &mustKVStoreBatch{
		prefix: prefix,
		batch:  batch,
	}
}

func (k *mustKVStoreBatch) Set(key kv.Key, value []byte) {
	if err := k.batch.Set(dbkeys.MakeKey(k.prefix, []byte(key)), value); err != nil {
		panic(err)
	}
}

func (k *mustKVStoreBatch) Del(key kv.Key) {
	if err := k.batch.Delete(dbkeys.MakeKey(k.prefix, []byte(key))); err != nil {
		panic(err)
	}
}

// Save saves all updates collected in the virtual state together with the provided blocks (if any) in one transaction
func (vs *virtualStateAccess) Save(blocks ...Block) error {
	if vs.kvs.Mutations().IsEmpty() {
		// nothing to commit
		vs.trie.ClearCache() // clear trie cache
		return nil
	}
	vs.Commit()

	batch, err := vs.db.Batched()
	if err != nil {
		panic(fmt.Errorf("error saving state: %w", err))
	}

	vs.trie.PersistMutations(newKVStoreBatch(dbkeys.ObjectTypeTrie, batch))
	vs.kvs.Mutations().Apply(newKVStoreBatch(dbkeys.ObjectTypeState, batch))
	for _, blk := range blocks {
		key := dbkeys.MakeKey(dbkeys.ObjectTypeBlock, util.Uint32To4Bytes(blk.BlockIndex()))
		if err := batch.Set(key, blk.Bytes()); err != nil {
			return err
		}
	}
	if err := batch.Commit(); err != nil {
		return err
	}

	// call flush explicitly, because batched.Commit doesn't actually write the changes to disk
	if err := vs.db.Flush(); err != nil {
		return err
	}

	if vs.onBlockSave != nil {
		// store or trace blocks if set so
		stateCommitment := trie.RootCommitment(vs.TrieNodeStore())
		for _, blk := range blocks {
			vs.onBlockSave(stateCommitment, blk)
		}
	}

	vs.trie.ClearCache()
	vs.kvs.ClearMutations()
	vs.kvs.Mutations().ResetModified()
	return nil
}

// LoadSolidState establishes VirtualStateAccess interface with the solid state in DB.
// Checks root commitment to chainID
func LoadSolidState(store kvstore.KVStore, chainID *iscp.ChainID) (VirtualStateAccess, bool, error) {
	// check the existence of terminalCommitment at key ''. chainID is expected
	v, err := store.Get(dbkeys.MakeKey(dbkeys.ObjectTypeState))
	if errors.Is(err, kvstore.ErrKeyNotFound) {
		// state does not exist
		return nil, false, nil
	}
	if err != nil {
		return nil, false, xerrors.Errorf("LoadSolidState: %v", err)
	}
	chID, err := iscp.ChainIDFromBytes(v)
	if err != nil {
		return nil, false, xerrors.Errorf("LoadSolidState: %v", err)
	}
	if !chID.Equals(chainID) {
		return nil, false, xerrors.Errorf("LoadSolidState: expected chainID: %s, got: %s", chainID, chID)
	}
	ret := NewVirtualState(store)

	// explicit use of merkle trie model. Asserting that the chainID is committed by the root at the key ''
	merkleProof := CommitmentModel.Proof(nil, ret.trie)
	if err = merkleProof.Validate(trie.RootCommitment(ret.trie), chainID.Bytes()); err != nil {
		return nil, false, xerrors.Errorf("LoadSolidState: can't prove inclusion of chain ID %s in the root: %v", chainID, err)
	}
	ret.kvs.Mutations().ResetModified()
	return ret, true, nil
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

// SaveRawBlockClosure return closure which saves block in specified directory
func SaveRawBlockClosure(dir string, log *logger.Logger) OnBlockSaveClosure {
	return func(stateCommitment trie.VCommitment, block Block) {
		data := block.Bytes()
		h := hashing.HashData(data)
		fname := fmt.Sprintf("%d.%s.%s.mut", block.BlockIndex(), stateCommitment.String(), h.String())
		err := os.WriteFile(path.Join(dir, fname), data, 0o600)
		if err != nil {
			log.Warnf("failed to save raw block #%d to dir %s as '%s': %v", block.BlockIndex(), dir, fname, err)
		} else {
			log.Infof("saved raw block #%d to dir %s as '%s'", block.BlockIndex(), dir, fname)
		}
	}
}
