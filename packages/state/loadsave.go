package state

import (
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

// Commit saves updates collected in the virtual state to DB together with the corresponding block in one transaction
// Mutations must be non-empty otherwise it is NOP
// It assumes uncommitted stateUpdates are either empty or consistent with mutations
// If uncommitted stateUpdates are empty, it creates a state update and block from mutations (used in mocking in tests).
func (vs *virtualState) Commit() error {
	blk, err := vs.UncommittedBlock()
	if err != nil {
		return err
	}
	if blk == nil {
		// nothing to commit
		return nil
	}
	batch := vs.db.Batched()

	if err = batch.Set(dbprovider.MakeKey(dbprovider.ObjectTypeStateHash), vs.Hash().Bytes()); err != nil {
		return err
	}
	if err = batch.Set(dbprovider.MakeKey(dbprovider.ObjectTypeStateIndex), util.Uint32To4Bytes(vs.BlockIndex())); err != nil {
		return err
	}

	key := dbprovider.MakeKey(dbprovider.ObjectTypeBlock, util.Uint32To4Bytes(vs.BlockIndex()))
	if err := batch.Set(key, blk.Bytes()); err != nil {
		return err
	}

	// store mutations
	for k, v := range vs.kvs.Mutations().Sets {
		if err := batch.Set(dbprovider.MakeKey(dbprovider.ObjectTypeStateVariable, []byte(k)), v); err != nil {
			return err
		}
	}
	for k := range vs.kvs.Mutations().Dels {
		if err := batch.Delete(dbprovider.MakeKey(dbprovider.ObjectTypeStateVariable, []byte(k))); err != nil {
			return err
		}
	}

	if err := batch.Commit(); err != nil {
		return err
	}

	vs.kvs.ClearMutations()
	// please the GC
	for i := range vs.uncommittedUpdates {
		vs.uncommittedUpdates[i] = nil
	}
	vs.uncommittedUpdates = vs.uncommittedUpdates[:0]
	return nil
}

// LoadSolidState establishes VirtualState interface with the solid state in DB. Checks consistency of DB
func LoadSolidState(dbp *dbprovider.DBProvider, chainID *coretypes.ChainID) (VirtualState, bool, error) {
	partition := dbp.GetPartition(chainID)

	stateIndex, stateHash, exists, err := loadStateIndexAndHashFromDb(partition)
	if err != nil {
		return nil, exists, xerrors.Errorf("LoadSolidState: %w", err)
	}
	if !exists {
		return nil, false, nil
	}
	vs := newVirtualState(partition, chainID)
	stateIndex1, err := loadStateIndexFromState(vs.KVStoreReader())
	if err != nil {
		return nil, false, xerrors.Errorf("LoadSolidState: %w", err)
	}
	if stateIndex != stateIndex1 {
		return nil, false, xerrors.New("LoadSolidState: state index inconsistent with the state")
	}
	vs.stateHash = stateHash
	return vs, true, nil
}

// LoadBlockBytes loads block bytes of the specified block index from DB
func LoadBlockBytes(partition kvstore.KVStore, stateIndex uint32) ([]byte, error) {
	data, err := partition.Get(dbprovider.MakeKey(dbprovider.ObjectTypeBlock, util.Uint32To4Bytes(stateIndex)))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

// LoadBlock loads block from DB and decodes it
func LoadBlock(partition kvstore.KVStore, stateIndex uint32) (*block, error) {
	data, err := LoadBlockBytes(partition, stateIndex)
	if err != nil {
		return nil, err
	}
	return BlockFromBytes(data)
}
