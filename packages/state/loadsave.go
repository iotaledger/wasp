package state

import (
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dbprovider"
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
	var err error
	batch := vs.db.Batched()

	if err = batch.Set(dbprovider.MakeKey(dbprovider.ObjectTypeStateHash), vs.Hash().Bytes()); err != nil {
		return err
	}
	if err = batch.Set(dbprovider.MakeKey(dbprovider.ObjectTypeStateIndex), util.Uint32To4Bytes(vs.BlockIndex())); err != nil {
		return err
	}

	for _, blk := range blocks {
		key := dbprovider.MakeKey(dbprovider.ObjectTypeBlock, util.Uint32To4Bytes(vs.BlockIndex()))
		if err := batch.Set(key, blk.Bytes()); err != nil {
			return err
		}
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
	for i := range vs.updateLog {
		vs.updateLog[i] = nil
	}
	vs.updateLog = vs.updateLog[:0]
	return nil
}

// CreateOriginState creates zero state which is the minimal consistent state.
// It is not committed it is an origin state. It has statically known hash coreutils.OriginStateHashBase58
func CreateOriginState(dbp *dbprovider.DBProvider, chainID *coretypes.ChainID) (*virtualState, error) {
	originState, originBlock := newZeroVirtualState(dbp.GetPartition(chainID), chainID)
	if err := originState.Commit(originBlock); err != nil {
		return nil, err
	}
	return originState, nil
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
func LoadBlockBytes(dbp *dbprovider.DBProvider, chainID *coretypes.ChainID, stateIndex uint32) ([]byte, error) {
	data, err := dbp.GetPartition(chainID).Get(dbprovider.MakeKey(dbprovider.ObjectTypeBlock, util.Uint32To4Bytes(stateIndex)))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

// LoadBlock loads block from DB and decodes it
func LoadBlock(dbp *dbprovider.DBProvider, chainID *coretypes.ChainID, stateIndex uint32) (*block, error) {
	data, err := LoadBlockBytes(dbp, chainID, stateIndex)
	if err != nil {
		return nil, err
	}
	return BlockFromBytes(data)
}
