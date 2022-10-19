package state

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/trie.go/trie"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
)

// valueBatch adaptor for the batch to kv.KVWriter
type valueBatch struct {
	kvstore.BatchedMutations
}

var _ kv.KVWriter = valueBatch{}

func newValueBatch(batch kvstore.BatchedMutations) valueBatch {
	return valueBatch{batch}
}

func (b valueBatch) Set(key kv.Key, value []byte) {
	k := dbkeys.MakeKey(dbkeys.ObjectTypeState, []byte(key))
	if err := b.BatchedMutations.Set(k, value); err != nil {
		panic(err)
	}
}

func (b valueBatch) Del(key kv.Key) {
	k := dbkeys.MakeKey(dbkeys.ObjectTypeState, []byte(key))
	if err := b.BatchedMutations.Delete(k); err != nil {
		panic(err)
	}
}

// trieBatch adaptor for the batch to trie.KVWriter
type trieBatch struct {
	kvstore.BatchedMutations
}

var _ trie.KVWriter = trieBatch{}

func newTrieBatch(batch kvstore.BatchedMutations) trieBatch {
	return trieBatch{batch}
}

func (b trieBatch) Set(key, value []byte) {
	k := dbkeys.MakeKey(dbkeys.ObjectTypeTrie, key)
	if len(value) == 0 {
		if err := b.BatchedMutations.Delete(k); err != nil {
			panic(err)
		}
	}
	if err := b.BatchedMutations.Set(k, value); err != nil {
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

	vs.trie.PersistMutations(newTrieBatch(batch))
	vs.kvs.Mutations().Apply(newValueBatch(batch))
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
func LoadSolidState(store kvstore.KVStore, chainID *isc.ChainID) (VirtualStateAccess, bool, error) {
	// check the existence of terminalCommitment at key ''. chainID is expected
	v, err := store.Get(dbkeys.MakeKey(dbkeys.ObjectTypeState))
	if errors.Is(err, kvstore.ErrKeyNotFound) {
		// state does not exist
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("LoadSolidState: %v", err)
	}
	chID, err := isc.ChainIDFromBytes(v)
	if err != nil {
		return nil, false, fmt.Errorf("LoadSolidState: %v", err)
	}
	if !chID.Equals(chainID) {
		return nil, false, fmt.Errorf("LoadSolidState: expected chainID: %s, got: %s", chainID, chID)
	}
	ret := NewVirtualState(store)

	// explicit use of merkle trie model. Asserting that the chainID is committed by the root at the key ''
	merkleProof := GetMerkleProof(nil, ret.trie)
	if err = ValidateMerkleProof(merkleProof, trie.RootCommitment(ret.trie), chainID.Bytes()); err != nil {
		return nil, false, fmt.Errorf("LoadSolidState: can't prove inclusion of chain ID %s in the root: %v", chainID, err)
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

// ForEachBlockIndex iterates over all indices of block saved in the DB
// Order non-deterministic
func ForEachBlockIndex(store kvstore.KVStore, fun func(blockIndex uint32) bool) error {
	var err error
	err1 := store.IterateKeys([]byte{dbkeys.ObjectTypeBlock}, func(key kvstore.Key) bool {
		var index uint32
		index, err = util.Uint32From4Bytes(key[1:])
		if err != nil {
			err = fmt.Errorf("Uint32From4Bytes: %v, data: %s", err, hexutil.Encode(key[1:]))
			return false
		}
		return fun(index)
	})
	if err1 != nil {
		return err1
	}
	return err
}
