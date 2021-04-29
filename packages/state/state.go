package state

import (
	"fmt"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

// region virtualState //////////////////////////////////////////////////////////////////////

type virtualState struct {
	chainID   coretypes.ChainID
	db        kvstore.KVStore
	stateData stateData
	empty     bool
	variables *buffered.BufferedKVStore
}

type stateData struct {
	blockIndex uint32
	timestamp  int64
	stateHash  hashing.HashValue
}

func stateDataFromBytes(data []byte) (*stateData, error) {
	mu := marshalutil.New(data)
	ret := &stateData{}
	var err error
	if ret.blockIndex, err = mu.ReadUint32(); err != nil {
		return nil, err
	}
	if ret.timestamp, err = mu.ReadInt64(); err != nil {
		return nil, err
	}
	if d, err := mu.ReadBytes(hashing.HashSize); err != nil {
		return nil, err
	} else {
		if ret.stateHash, err = hashing.HashValueFromBytes(d); err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (s *stateData) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteUint32(s.blockIndex).
		WriteInt64(s.timestamp).
		WriteBytes(s.stateHash[:])
	return mu.Bytes()
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

const OriginStateHashBase58 = "FH1PadHHVLik9Sx5SQ7e2GWYjP7TnLiwDeJjKwxMHui5"

// OriginStateHash is independent from db provider nor chainID
func OriginStateHash() hashing.HashValue {
	emptyVirtualState := NewZeroVirtualState(mapdb.NewMapDB())
	return emptyVirtualState.Hash()
}

func subRealm(db kvstore.KVStore, realm []byte) kvstore.KVStore {
	if db == nil {
		return nil
	}
	return db.WithRealm(append(db.Realm(), realm...))
}

func (vs *virtualState) Clone() VirtualState {
	return &virtualState{
		chainID:   vs.chainID,
		db:        vs.db,
		stateData: vs.stateData,
		empty:     vs.empty,
		variables: vs.variables.Clone(),
	}
}

func (vs *virtualState) DangerouslyConvertToString() string {
	return fmt.Sprintf("#%d, ts: %d, hash, %s\n%s",
		vs.stateData.blockIndex,
		vs.stateData.timestamp,
		vs.stateData.stateHash.String(),
		vs.KVStore().DangerouslyDumpToString(),
	)
}

func (vs *virtualState) KVStore() *buffered.BufferedKVStore {
	return vs.variables
}

func (vs *virtualState) KVStoreReader() kv.KVStoreReader {
	return vs.variables
}

func (vs *virtualState) BlockIndex() uint32 {
	return vs.stateData.blockIndex
}

func (vs *virtualState) ApplyBlockIndex(blockIndex uint32) {
	vh := vs.Hash()
	vs.stateData.stateHash = hashing.HashData(vh[:], util.Uint32To4Bytes(blockIndex))
	vs.empty = false
	vs.stateData.blockIndex = blockIndex
}

func (vs *virtualState) Timestamp() int64 {
	return vs.stateData.timestamp
}

// applies block of state updates. Increases state index
func (vs *virtualState) ApplyBlock(batch Block) error {
	if !vs.empty {
		if batch.StateIndex() != vs.stateData.blockIndex+1 {
			return fmt.Errorf("ApplyBlock: block state index #%d can't be applied to the state #%d",
				batch.StateIndex(), vs.stateData.blockIndex)
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
	vs.stateData.timestamp = stateUpd.Timestamp()
	vh := vs.Hash()
	sh := util.GetHashValue(stateUpd)
	vs.stateData.stateHash = hashing.HashData(vh[:], sh[:], util.Uint64To8Bytes(uint64(vs.stateData.timestamp)))
	vs.empty = false
}

func (vs *virtualState) Hash() hashing.HashValue {
	return vs.stateData.stateHash
}

// saves variable state to db atomically with the block of state updates and records of processed requests
func (vs *virtualState) CommitToDb(b Block) error {
	batch := vs.db.Batched()

	{
		blockData, err := util.Bytes(b)
		if err != nil {
			return err
		}
		if err = batch.Set(dbkeyBlock(b.StateIndex()), blockData); err != nil {
			return err
		}
	}

	{
		varStateData := vs.stateData.Bytes()
		if err := batch.Set(dbprovider.MakeKey(dbprovider.ObjectTypeSolidState), varStateData); err != nil {
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

// endregion ////////////////////////////////////////////////////////////////////

func LoadSolidState(dbp *dbprovider.DBProvider, chainID *coretypes.ChainID) (VirtualState, bool, error) {
	partition := dbp.GetPartition(chainID)
	v, err := partition.Get(dbprovider.MakeKey(dbprovider.ObjectTypeSolidState))
	if err == kvstore.ErrKeyNotFound {
		// state does not exist
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	vs := NewVirtualState(partition, chainID)
	sd, err := stateDataFromBytes(v)
	if err != nil {
		return nil, false, xerrors.Errorf("loading solid state: %w", err)
	}
	vs.stateData = *sd
	return vs, true, nil
}

func dbkeyStateVariable(key kv.Key) []byte {
	return dbprovider.MakeKey(dbprovider.ObjectTypeStateVariable, []byte(key))
}

// region StateReader /////////////////////////////////////////////////////////

type stateReader struct {
	chainPartition kvstore.KVStore
	chainState     kv.KVStoreReader
}

func NewStateReader(dbp *dbprovider.DBProvider, chainID *coretypes.ChainID) *stateReader {
	partition := dbp.GetPartition(chainID)
	return &stateReader{
		chainPartition: partition,
		chainState:     kv.NewHiveKVStoreReader(subRealm(partition, []byte{dbprovider.ObjectTypeStateVariable})),
	}
}

func (r *stateReader) BlockIndex() uint32 {
	sdBin, err := r.chainPartition.Get(dbprovider.MakeKey(dbprovider.ObjectTypeSolidState))
	if err != nil {
		panic(err)
	}
	stateData, err := stateDataFromBytes(sdBin)
	if err != nil {
		panic(err)
	}
	return stateData.blockIndex
}

func (r *stateReader) Timestamp() int64 {
	sdBin, err := r.chainPartition.Get(dbprovider.MakeKey(dbprovider.ObjectTypeSolidState))
	if err != nil {
		panic(err)
	}
	stateData, err := stateDataFromBytes(sdBin)
	if err != nil {
		panic(err)
	}
	return stateData.timestamp
}

func (r *stateReader) Hash() hashing.HashValue {
	sdBin, err := r.chainPartition.Get(dbprovider.MakeKey(dbprovider.ObjectTypeSolidState))
	if err != nil {
		panic(err)
	}
	stateData, err := stateDataFromBytes(sdBin)
	if err != nil {
		panic(err)
	}
	return stateData.stateHash
}

func (r *stateReader) KVStoreReader() kv.KVStoreReader {
	return r.chainState
}

// endregion ///////////////////////////////////////////////////////////////////
