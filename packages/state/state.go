package state

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/variables"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/iotaledger/wasp/plugins/database"
	"io"
)

type virtualState struct {
	stateIndex uint32
	timestamp  int64
	empty      bool
	stateHash  hashing.HashValue
	vars       variables.Variables
}

// VirtualState new empty or clone
func NewVirtualState(varState VirtualState) VirtualState {
	if varState == nil {
		return &virtualState{
			vars:  variables.New(nil),
			empty: true,
		}
	}
	return &virtualState{
		timestamp:  varState.Timestamp(),
		stateIndex: varState.StateIndex(),
		stateHash:  varState.Hash(),
		vars:       variables.New(varState.Variables()),
	}
}

func (vs *virtualState) InitiatedBy(ownerAddr *address.Address) bool {
	addr, ok, err := vs.Variables().GetAddress(vmconst.VarNameOwnerAddress)
	if !ok || err != nil {
		return false
	}
	return addr == *ownerAddr
}

func (vs *virtualState) String() string {
	return fmt.Sprintf("#%d, ts: %d, hash, %s\n%s",
		vs.stateIndex, vs.timestamp, vs.stateHash.String(), vs.Variables().String())
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

func (vs *virtualState) Variables() variables.Variables {
	return vs.vars
}

func (vs *virtualState) saveToDb(addr *address.Address) error {
	data, err := util.Bytes(vs)
	if err != nil {
		return err
	}

	if err := database.GetPartition(addr).Set(database.MakeKey(database.ObjectTypeVariableState), data); err != nil {
		return err
	}

	h := vs.Hash()
	log.Debugw("state saving to db",
		"addr", addr.String(),
		"state index", vs.StateIndex(),
		"stateHash", h.String(),
	)
	return nil
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
	if err := vs.vars.Write(w); err != nil {
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
	if err := vs.vars.Read(r); err != nil {
		return err
	}
	return nil
}

// saves variable state to db atomically with the batch of state updates and records of processed requests
func (vs *virtualState) CommitToDb(addr address.Address, b Batch) error {
	batchData, err := util.Bytes(b)
	if err != nil {
		return err
	}
	batchDbKey := dbkeyBatch(b.StateIndex())

	varStateData, err := util.Bytes(vs)
	if err != nil {
		return err
	}
	varStateDbkey := database.MakeKey(database.ObjectTypeVariableState)

	solidStateValue := util.Uint32To4Bytes(vs.StateIndex())
	solidStateKey := database.MakeKey(database.ObjectTypeSolidStateIndex)

	keys := [][]byte{varStateDbkey, batchDbKey, solidStateKey}
	values := [][]byte{varStateData, batchData, solidStateValue}

	// store successfully processed request IDs
	for _, rid := range b.RequestIds() {
		keys = append(keys, dbkeyRequest(rid))
		values = append(values, []byte{0})
	}

	db := database.GetPartition(&addr)

	return util.StoreSetToDb(db, keys, values)
}

func LoadSolidState(addr *address.Address) (VirtualState, Batch, bool, error) {
	db := database.GetPartition(addr)
	stateIndexBin, err := db.Get(database.MakeKey(database.ObjectTypeSolidStateIndex))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil, false, nil
	}
	if err != nil {
		return nil, nil, false, err
	}
	values, err := util.GetSetFromDb(db, [][]byte{
		database.MakeKey(database.ObjectTypeVariableState),
		dbkeyBatch(util.Uint32From4Bytes(stateIndexBin)),
	})
	if err != nil {
		return nil, nil, false, err
	}

	varState := NewVirtualState(nil).(*virtualState)
	if err = varState.Read(bytes.NewReader(values[0])); err != nil {
		return nil, nil, false, fmt.Errorf("loading variable state: %v", err)
	}

	batch, err := BatchFromBytes(values[1])
	if err != nil {
		return nil, nil, false, fmt.Errorf("loading batch: %v", err)
	}
	if varState.StateIndex() != batch.StateIndex() {
		return nil, nil, false, fmt.Errorf("inconsistent solid state: state indices must be equal")
	}
	return varState, batch, true, nil
}

func dbkeyRequest(reqid *sctransaction.RequestId) []byte {
	return database.MakeKey(database.ObjectTypeProcessedRequestId, reqid[:])
}

func MarkRequestProcessedFailure(addr *address.Address, reqid *sctransaction.RequestId) error {
	has, err := database.GetPartition(addr).Has(dbkeyRequest(reqid))
	if err != nil {
		return err
	}
	dbkey := dbkeyRequest(reqid)
	if !has {
		return database.GetPartition(addr).Set(dbkey, []byte{1})
	}
	value, err := database.GetPartition(addr).Get(dbkey)
	if err != nil {
		return err
	}
	if len(value) != 1 {
		return fmt.Errorf("inconistency: len(value) != 1")
	}
	return database.GetPartition(addr).Set(dbkey, []byte{value[0] + 1})
}

const maxRetriesForRequest = byte(5)

// IsRequestCompleted returns true if it was completed successfully or number of retries reached maximum
func IsRequestCompleted(addr *address.Address, reqid *sctransaction.RequestId) (bool, error) {
	dbkey := dbkeyRequest(reqid)
	has, err := database.GetPartition(addr).Has(dbkey)
	if err != nil {
		return false, err
	}
	if !has {
		return false, nil
	}
	val, err := database.GetPartition(addr).Get(dbkey)
	if err != nil {
		return false, err
	}
	if len(val) != 1 {
		return false, fmt.Errorf("inconistency: len(val) != 1")
	}
	return val[0] == 0 || val[0] >= maxRetriesForRequest, nil
}
