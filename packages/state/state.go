package state

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/iotaledger/wasp/plugins/database"
	"io"
)

type virtualState struct {
	scAddress    address.Address
	getPartition func(*address.Address) kvstore.KVStore // for unit testing
	stateIndex   uint32
	timestamp    int64
	empty        bool
	stateHash    hashing.HashValue
	variables    kv.BufferedKVStore
}

func newVirtualState(scAddress *address.Address, getPartition func(*address.Address) kvstore.KVStore) *virtualState {
	return &virtualState{
		scAddress:    *scAddress,
		getPartition: getPartition,
		variables: kv.NewBufferedKVStoreOnSubrealm(
			func() kvstore.KVStore { return getPartition(scAddress) },
			[]byte{database.ObjectTypeStateVariable},
		),
		empty: true,
	}
}

func getSCPartition(addr *address.Address) kvstore.KVStore {
	return database.GetPartition(addr)
}

func NewEmptyVirtualState(addr *address.Address) VirtualState {
	return newVirtualState(addr, getSCPartition)
}

func (vs *virtualState) Clone() VirtualState {
	return &virtualState{
		scAddress:    vs.scAddress,
		getPartition: vs.getPartition,
		stateIndex:   vs.stateIndex,
		timestamp:    vs.timestamp,
		empty:        vs.empty,
		stateHash:    vs.stateHash,
		variables:    vs.variables.Clone(),
	}
}

func (vs *virtualState) InitiatedBy(ownerAddr *address.Address) bool {
	addr, ok, err := vs.Variables().Codec().GetAddress(vmconst.VarNameOwnerAddress)
	if !ok || err != nil {
		return false
	}
	return *addr == *ownerAddr
}

func (vs *virtualState) DangerouslyConvertToString() string {
	return fmt.Sprintf("#%d, ts: %d, hash, %s\n%s",
		vs.stateIndex,
		vs.timestamp,
		vs.stateHash.String(),
		vs.Variables().DangerouslyDumpToString(),
	)
}

func (vs *virtualState) Variables() kv.BufferedKVStore {
	return vs.variables
}

func (vs *virtualState) StateIndex() uint32 {
	return vs.stateIndex
}

func (vs *virtualState) ApplyStateIndex(stateIndex uint32) {
	vh := vs.Hash()
	vs.stateHash = *hashing.HashData(vh.Bytes(), util.Uint32To4Bytes(stateIndex))
	vs.empty = false
	vs.stateIndex = stateIndex
}

func (vs *virtualState) Timestamp() int64 {
	return vs.timestamp
}

// applies batch of state updates. Increases state index
func (vs *virtualState) ApplyBatch(batch Batch) error {
	if !vs.empty {
		if batch.StateIndex() != vs.stateIndex+1 {
			return fmt.Errorf("batch state index #%d can't be applied to the state #%d",
				batch.StateIndex(), vs.stateIndex)
		}
	} else {
		if batch.StateIndex() != 0 {
			return fmt.Errorf("batch state index #%d can't be applied to the empty state", batch.StateIndex())
		}
	}
	batch.ForEach(func(_ uint16, stateUpd StateUpdate) bool {
		vs.ApplyStateUpdate(stateUpd)
		return true
	})
	vs.ApplyStateIndex(batch.StateIndex())
	return nil
}

// applies one state update. Doesn't change state index
func (vs *virtualState) ApplyStateUpdate(stateUpd StateUpdate) {
	stateUpd.Mutations().ApplyTo(vs.Variables())
	vs.timestamp = stateUpd.Timestamp()
	vh := vs.Hash()
	sh := util.GetHashValue(stateUpd)
	vs.stateHash = *hashing.HashData(vh.Bytes(), sh.Bytes(), util.Uint64To8Bytes(uint64(vs.timestamp)))
	vs.empty = false
}

func (vs *virtualState) Hash() hashing.HashValue {
	return vs.stateHash
}

func (vs *virtualState) Write(w io.Writer) error {
	if _, err := w.Write(util.Uint32To4Bytes(vs.stateIndex)); err != nil {
		return err
	}
	if err := util.WriteUint64(w, uint64(vs.timestamp)); err != nil {
		return err
	}
	if _, err := w.Write(vs.stateHash.Bytes()); err != nil {
		return err
	}
	return nil
}

func (vs *virtualState) Read(r io.Reader) error {
	if err := util.ReadUint32(r, &vs.stateIndex); err != nil {
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
	return nil
}

// saves variable state to db atomically with the batch of state updates and records of processed requests
func (vs *virtualState) CommitToDb(b Batch) error {
	batchData, err := util.Bytes(b)
	if err != nil {
		return err
	}
	batchDbKey := dbkeyBatch(b.StateIndex())

	varStateData, err := util.Bytes(vs)
	if err != nil {
		return err
	}
	varStateDbkey := database.MakeKey(database.ObjectTypeSolidState)

	solidStateValue := util.Uint32To4Bytes(vs.StateIndex())
	solidStateKey := database.MakeKey(database.ObjectTypeSolidStateIndex)

	keys := [][]byte{varStateDbkey, batchDbKey, solidStateKey}
	values := [][]byte{varStateData, batchData, solidStateValue}

	// store successfully processed request IDs
	for _, rid := range b.RequestIds() {
		keys = append(keys, dbkeyRequest(rid))
		values = append(values, []byte{0})
	}

	// store uncommitted mutations
	vs.variables.Mutations().Iterate(func(k kv.Key, mut kv.Mutation) bool {
		keys = append(keys, dbkeyStateVariable(k))

		// if mutation is MutationDel, mut.Value() = nil and the key is deleted
		values = append(values, mut.Value())
		return true
	})

	db := vs.getPartition(&vs.scAddress)
	err = util.DbSetMulti(db, keys, values)
	if err != nil {
		return err
	}
	vs.variables.ClearMutations()
	return nil
}

func LoadSolidState(addr *address.Address) (VirtualState, Batch, bool, error) {
	return loadSolidState(addr, getSCPartition)
}

func loadSolidState(scAddress *address.Address, getPartition func(*address.Address) kvstore.KVStore) (VirtualState, Batch, bool, error) {
	db := getPartition(scAddress)

	stateIndexBin, err := db.Get(database.MakeKey(database.ObjectTypeSolidStateIndex))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil, false, nil
	}
	if err != nil {
		return nil, nil, false, err
	}
	values, err := util.DbGetMulti(db, [][]byte{
		database.MakeKey(database.ObjectTypeSolidState),
		dbkeyBatch(util.Uint32From4Bytes(stateIndexBin)),
	})
	if err != nil {
		return nil, nil, false, err
	}

	vs := newVirtualState(scAddress, getPartition)
	if err = vs.Read(bytes.NewReader(values[0])); err != nil {
		return nil, nil, false, fmt.Errorf("loading variable state: %v", err)
	}

	batch, err := BatchFromBytes(values[1])
	if err != nil {
		return nil, nil, false, fmt.Errorf("loading batch: %v", err)
	}
	if vs.StateIndex() != batch.StateIndex() {
		return nil, nil, false, fmt.Errorf("inconsistent solid state: state indices must be equal")
	}
	return vs, batch, true, nil
}

func dbkeyStateVariable(key kv.Key) []byte {
	return database.MakeKey(database.ObjectTypeStateVariable, []byte(key))
}

func dbkeyRequest(reqid *sctransaction.RequestId) []byte {
	return database.MakeKey(database.ObjectTypeProcessedRequestId, reqid[:])
}

func MarkRequestProcessedFailure(addr *address.Address, reqid *sctransaction.RequestId) error {
	has, err := getSCPartition(addr).Has(dbkeyRequest(reqid))
	if err != nil {
		return err
	}
	dbkey := dbkeyRequest(reqid)
	if !has {
		return getSCPartition(addr).Set(dbkey, []byte{1})
	}
	value, err := getSCPartition(addr).Get(dbkey)
	if err != nil {
		return err
	}
	if len(value) != 1 {
		return fmt.Errorf("inconistency: len(value) != 1")
	}
	return getSCPartition(addr).Set(dbkey, []byte{value[0] + 1})
}

const maxRetriesForRequest = byte(5)

// IsRequestCompleted returns true if it was completed successfully or number of retries reached maximum
func IsRequestCompleted(addr *address.Address, reqid *sctransaction.RequestId) (bool, error) {
	dbkey := dbkeyRequest(reqid)
	has, err := getSCPartition(addr).Has(dbkey)
	if err != nil {
		return false, err
	}
	if !has {
		return false, nil
	}
	val, err := getSCPartition(addr).Get(dbkey)
	if err != nil {
		return false, err
	}
	if len(val) != 1 {
		return false, fmt.Errorf("inconistency: len(val) != 1")
	}
	return val[0] == 0 || val[0] >= maxRetriesForRequest, nil
}
