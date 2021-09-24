package state

import (
	"errors"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

// Commit saves updates collected in the virtual state to DB together with the provided blocks in one transaction
// Mutations must be non-empty otherwise it is NOP
// It the log of updates is not taken into account
func (vs *virtualState) Commit(blocks ...Block) error {
	if vs.kvs.Mutations().IsEmpty() {
		// nothing to commit
		return nil
	}
	batch := vs.db.Batched()

	if err := batch.Set(dbkeys.MakeKey(dbkeys.ObjectTypeStateHash), vs.StateCommitment().Bytes()); err != nil {
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
		if err := batch.Set(dbkeys.MakeKey(dbkeys.ObjectTypeStateVariable, []byte(k)), v); err != nil {
			return err
		}
	}
	for k := range vs.kvs.Mutations().Dels {
		if err := batch.Delete(dbkeys.MakeKey(dbkeys.ObjectTypeStateVariable, []byte(k))); err != nil {
			return err
		}
	}

	if err := batch.Commit(); err != nil {
		return err
	}

	vs.kvs.ClearMutations()
	return nil
}

// CreateOriginState creates zero state which is the minimal consistent state.
// It is not committed it is an origin state. It has statically known hash coreutils.OriginStateHashBase58
func CreateOriginState(store kvstore.KVStore, chainID *iscp.ChainID) (VirtualState, error) {
	originState, originBlock := newZeroVirtualState(store, chainID)
	if err := originState.Commit(originBlock); err != nil {
		return nil, err
	}
	return originState, nil
}

// LoadSolidState establishes VirtualState interface with the solid state in DB. Checks consistency of DB
func LoadSolidState(store kvstore.KVStore, chainID *iscp.ChainID) (VirtualState, bool, error) {
	stateHash, exists, err := loadStateHashFromDb(store)
	if err != nil {
		return nil, exists, xerrors.Errorf("LoadSolidState: %w", err)
	}
	if !exists {
		return nil, false, nil
	}
	vs := newVirtualState(store, chainID)
	vs.stateHash = stateHash
	vs.isStateHashOutdated = false // NOTE: returned VirtualState doesn't contain the block mutations, so if the hash is re-computed, it won't match the block that was committed to the DB.
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
