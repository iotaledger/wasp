package state

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"io"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/database"
)

type virtualState struct {
	chainID    coretypes.ChainID
	db         kvstore.KVStore
	blockIndex uint32
	timestamp  int64
	empty      bool
	stateHash  hashing.HashValue
	variables  buffered.BufferedKVStore
}

func NewVirtualState(db kvstore.KVStore, chainID *coretypes.ChainID) *virtualState {
	return &virtualState{
		chainID:   *chainID,
		db:        db,
		variables: buffered.NewBufferedKVStore(subRealm(db, []byte{dbprovider.ObjectTypeStateVariable})),
		empty:     true,
	}
}

func NewEmptyVirtualState(chainID *coretypes.ChainID) *virtualState {
	return NewVirtualState(getSCPartition(chainID), chainID)
}

func getSCPartition(chainID *coretypes.ChainID) kvstore.KVStore {
	return database.GetPartition(chainID)
}

func subRealm(db kvstore.KVStore, realm []byte) kvstore.KVStore {
	if db == nil {
		return nil
	}
	return db.WithRealm(append(db.Realm(), realm...))
}

func (vs *virtualState) Clone() VirtualState {
	return &virtualState{
		chainID:    vs.chainID,
		db:         vs.db,
		blockIndex: vs.blockIndex,
		timestamp:  vs.timestamp,
		empty:      vs.empty,
		stateHash:  vs.stateHash,
		variables:  vs.variables.Clone(),
	}
}

func (vs *virtualState) DangerouslyConvertToString() string {
	return fmt.Sprintf("#%d, ts: %d, hash, %s\n%s",
		vs.blockIndex,
		vs.timestamp,
		vs.stateHash.String(),
		vs.Variables().DangerouslyDumpToString(),
	)
}

func (vs *virtualState) Variables() buffered.BufferedKVStore {
	return vs.variables
}

func (vs *virtualState) BlockIndex() uint32 {
	return vs.blockIndex
}

func (vs *virtualState) ApplyBlockIndex(blockIndex uint32) {
	vh := vs.Hash()
	vs.stateHash = hashing.HashData(vh[:], util.Uint32To4Bytes(blockIndex))
	vs.empty = false
	vs.blockIndex = blockIndex
}

func (vs *virtualState) Timestamp() int64 {
	return vs.timestamp
}

// applies block of state updates. Increases state index
func (vs *virtualState) ApplyBlock(batch Block) error {
	if !vs.empty {
		if batch.StateIndex() != vs.blockIndex+1 {
			return fmt.Errorf("ApplyBlock: block state index #%d can't be applied to the state #%d",
				batch.StateIndex(), vs.blockIndex)
		}
	} else {
		if batch.StateIndex() != 0 {
			return fmt.Errorf("ApplyBlock: block state index #%d can't be applied to the empty state", batch.StateIndex())
		}
	}
	batch.ForEach(func(_ uint16, stateUpd StateUpdate) bool {
		vs.ApplyStateUpdate(stateUpd)
		return true
	})
	vs.ApplyBlockIndex(batch.StateIndex())
	return nil
}

// applies one state update. Doesn't change state index
func (vs *virtualState) ApplyStateUpdate(stateUpd StateUpdate) {
	stateUpd.Mutations().ApplyTo(vs.Variables())
	vs.timestamp = stateUpd.Timestamp()
	vh := vs.Hash()
	sh := util.GetHashValue(stateUpd)
	vs.stateHash = hashing.HashData(vh[:], sh[:], util.Uint64To8Bytes(uint64(vs.timestamp)))
	vs.empty = false
}

func (vs *virtualState) Hash() hashing.HashValue {
	return vs.stateHash
}

func (vs *virtualState) Write(w io.Writer) error {
	if _, err := w.Write(util.Uint32To4Bytes(vs.blockIndex)); err != nil {
		return err
	}
	if err := util.WriteUint64(w, uint64(vs.timestamp)); err != nil {
		return err
	}
	if _, err := w.Write(vs.stateHash[:]); err != nil {
		return err
	}
	return nil
}

func (vs *virtualState) Read(r io.Reader) error {
	if err := util.ReadUint32(r, &vs.blockIndex); err != nil {
		return err
	}
	var ts uint64
	if err := util.ReadUint64(r, &ts); err != nil {
		return err
	}
	vs.timestamp = int64(ts)
	if _, err := r.Read(vs.stateHash[:]); err != nil {
		return err
	}
	// after reading something, the state is not empty
	vs.empty = false
	return nil
}

// saves variable state to db atomically with the block of state updates and records of processed requests
func (vs *virtualState) CommitToDb(b Block) error {
	batchData, err := util.Bytes(b)
	if err != nil {
		return err
	}
	batchDbKey := dbkeyBatch(b.StateIndex())

	varStateData, err := util.Bytes(vs)
	if err != nil {
		return err
	}
	varStateDbkey := dbprovider.MakeKey(dbprovider.ObjectTypeSolidState)

	solidStateValue := util.Uint32To4Bytes(vs.BlockIndex())
	solidStateKey := dbprovider.MakeKey(dbprovider.ObjectTypeSolidStateIndex)

	keys := [][]byte{varStateDbkey, batchDbKey, solidStateKey}
	values := [][]byte{varStateData, batchData, solidStateValue}

	// store processed request IDs
	// TODO store request IDs in the 'log' contract
	for _, rid := range b.RequestIDs() {
		keys = append(keys, dbkeyRequest(&rid))
		values = append(values, []byte{0})
	}

	// store uncommitted mutations
	vs.variables.Mutations().IterateLatest(func(k kv.Key, mut buffered.Mutation) bool {
		keys = append(keys, dbkeyStateVariable(k))

		// if mutation is MutationDel, mut.Value() = nil and the key is deleted
		values = append(values, mut.Value())
		return true
	})

	err = util.DbSetMulti(vs.db, keys, values)
	if err != nil {
		return err
	}
	vs.variables.ClearMutations()
	return nil
}

func LoadSolidState(chainID *coretypes.ChainID) (VirtualState, Block, bool, error) {
	return loadSolidState(getSCPartition(chainID), chainID)
}

func loadSolidState(db kvstore.KVStore, chainID *coretypes.ChainID) (VirtualState, Block, bool, error) {
	stateIndexBin, err := db.Get(dbprovider.MakeKey(dbprovider.ObjectTypeSolidStateIndex))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil, false, nil
	}
	if err != nil {
		return nil, nil, false, err
	}
	values, err := util.DbGetMulti(db, [][]byte{
		dbprovider.MakeKey(dbprovider.ObjectTypeSolidState),
		dbkeyBatch(util.MustUint32From4Bytes(stateIndexBin)),
	})
	if err != nil {
		return nil, nil, false, err
	}

	vs := NewVirtualState(db, chainID)
	if err = vs.Read(bytes.NewReader(values[0])); err != nil {
		return nil, nil, false, fmt.Errorf("loading variable state: %v", err)
	}

	batch, err := NewBlockFromBytes(values[1])
	if err != nil {
		return nil, nil, false, fmt.Errorf("loading block: %v", err)
	}
	if vs.BlockIndex() != batch.StateIndex() {
		return nil, nil, false, fmt.Errorf("inconsistent solid state: state indices must be equal")
	}
	return vs, batch, true, nil
}

func dbkeyStateVariable(key kv.Key) []byte {
	return dbprovider.MakeKey(dbprovider.ObjectTypeStateVariable, []byte(key))
}

func dbkeyRequest(reqid *ledgerstate.OutputID) []byte {
	return dbprovider.MakeKey(dbprovider.ObjectTypeProcessedRequestId, reqid[:])
}

func IsRequestCompleted(addr *coretypes.ChainID, reqid *ledgerstate.OutputID) (bool, error) {
	return getSCPartition(addr).Has(dbkeyRequest(reqid))
}
