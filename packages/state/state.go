package state

import (
	"bytes"
	"fmt"
	"io"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/util"
)

type virtualState struct {
	chainID    coretypes.ChainID
	db         kvstore.KVStore
	blockIndex uint32
	timestamp  int64
	empty      bool
	stateHash  hashing.HashValue
	variables  *buffered.BufferedKVStore
}

func NewVirtualState(db kvstore.KVStore, chainID *coretypes.ChainID) *virtualState {
	ret := &virtualState{
		db:        db,
		variables: buffered.NewBufferedKVStore(kv.NewHiveKVStoreReader(subRealm(db, []byte{dbprovider.ObjectTypeStateVariable}))),
		empty:     true,
	}
	if chainID != nil {
		ret.chainID = *chainID
	}
	return ret
}

func NewZeroVirtualState(db kvstore.KVStore) *virtualState {
	ret := NewVirtualState(db, nil)
	originBlock := MustNewOriginBlock()
	if err := ret.ApplyBlock(originBlock); err != nil {
		panic(err)
	}
	return ret
}

const OriginStateHashBase58 = "4oBFMdSF5zGatLxggeCAPmJcTMj8xNMmxwTwGbEdTLPC"

// OriginStateHash is independent from db provider nor chainID
func OriginStateHash() hashing.HashValue {
	emptyVirtualState := NewZeroVirtualState(mapdb.NewMapDB())
	return emptyVirtualState.Hash()
}

func getChainPartition(dbp *dbprovider.DBProvider, chainID *coretypes.ChainID) kvstore.KVStore {
	return dbp.GetPartition(chainID)
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
		vs.KVStore().DangerouslyDumpToString(),
	)
}

func (vs *virtualState) KVStore() *buffered.BufferedKVStore {
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
	stateUpd.Mutations().ApplyTo(vs.KVStore())
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
	batch := vs.db.Batched()

	{
		blockData, err := util.Bytes(b)
		if err != nil {
			return err
		}
		if err = batch.Set(dbkeyBatch(b.StateIndex()), blockData); err != nil {
			return err
		}
	}

	{
		varStateData, err := util.Bytes(vs)
		if err != nil {
			return err
		}
		if err = batch.Set(dbprovider.MakeKey(dbprovider.ObjectTypeSolidState), varStateData); err != nil {
			return err
		}
	}

	{
		solidStateValue := util.Uint32To4Bytes(vs.BlockIndex())
		if err := batch.Set(dbprovider.MakeKey(dbprovider.ObjectTypeSolidStateIndex), solidStateValue); err != nil {
			return err
		}
	}

	// store processed request IDs
	// TODO store request IDs in the 'log' contract
	for _, rid := range b.RequestIDs() {
		if err := batch.Set(dbkeyRequest(rid), []byte{0}); err != nil {
			return err
		}
	}

	// store uncommitted mutations
	for k, v := range vs.variables.Mutations().Sets {
		if err := batch.Set(dbkeyStateVariable(k), v); err != nil {
			return err
		}
	}
	for k := range vs.variables.Mutations().Dels {
		if err := batch.Delete(dbkeyStateVariable(k)); err != nil {
			return err
		}
	}

	if err := batch.Commit(); err != nil {
		return err
	}
	vs.variables.ClearMutations()
	return nil
}

func LoadSolidState(dbp *dbprovider.DBProvider, chainID *coretypes.ChainID) (VirtualState, Block, bool, error) {
	return loadSolidState(getChainPartition(dbp, chainID), chainID)
}

func loadSolidState(db kvstore.KVStore, chainID *coretypes.ChainID) (VirtualState, Block, bool, error) {
	stateIndexBin, err := db.Get(dbprovider.MakeKey(dbprovider.ObjectTypeSolidStateIndex))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil, false, nil
	}
	if err != nil {
		return nil, nil, false, err
	}

	var v []byte

	vs := NewVirtualState(db, chainID)
	if v, err = db.Get(dbprovider.MakeKey(dbprovider.ObjectTypeSolidState)); err != nil {
		return nil, nil, false, err
	}
	if err = vs.Read(bytes.NewReader(v)); err != nil {
		return nil, nil, false, fmt.Errorf("loading variable state: %v", err)
	}

	if v, err = db.Get(dbkeyBatch(util.MustUint32From4Bytes(stateIndexBin))); err != nil {
		return nil, nil, false, err
	}
	batch, err := BlockFromBytes(v)
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

func dbkeyRequest(reqid coretypes.RequestID) []byte {
	return dbprovider.MakeKey(dbprovider.ObjectTypeProcessedRequestId, reqid[:])
}

func IsRequestCompleted(chainState kvstore.KVStore, reqid coretypes.RequestID) (bool, error) {
	return chainState.Has(dbkeyRequest(reqid))
}

func StoreRequestCompleted(chainState kvstore.KVStore, reqID coretypes.RequestID) error {
	return chainState.Set(dbkeyRequest(reqID), []byte{0})
}
