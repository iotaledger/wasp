package state

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/variables"
	"io"
)

type variableState struct {
	stateIndex uint32
	empty      bool
	stateHash  hashing.HashValue
	vars       variables.Variables
}

// VariableState

func NewVariableState(varState VariableState) VariableState {
	if varState == nil {
		return &variableState{
			vars:  variables.New(nil),
			empty: true,
		}
	}
	return &variableState{
		stateIndex: varState.StateIndex(),
		vars:       variables.New(varState.Variables()),
	}
}

func (vs *variableState) StateIndex() uint32 {
	return vs.stateIndex
}

// the only method which changes state of the variableState object
func (vs *variableState) Apply(batch Batch) error {
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
	batch.ForEach(func(stateUpd StateUpdate) bool {
		vs.Variables().Apply(stateUpd.Variables())
		return false
	})
	vs.stateHash = *hashing.HashData(vs.Hash().Bytes(), batch.EssenceHash().Bytes())
	vs.stateIndex = batch.StateIndex()
	vs.empty = false
	return nil
}

func (vs *variableState) Hash() *hashing.HashValue {
	return &vs.stateHash
}

func (vs *variableState) Variables() variables.Variables {
	return vs.vars
}

func (vs *variableState) saveToDb(addr *address.Address) error {
	dbase, err := database.GetVariableStateDB()
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   database.DbPrefixState(addr, vs.stateIndex),
		Value: hashing.MustBytes(vs),
	})
}

func (vs *variableState) Write(w io.Writer) error {
	if _, err := w.Write(util.Uint32To4Bytes(vs.stateIndex)); err != nil {
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

func (vs *variableState) Read(r io.Reader) error {
	if err := util.ReadUint32(r, &vs.stateIndex); err != nil {
		return err
	}
	if _, err := r.Read(vs.stateHash[:]); err != nil {
		return err
	}
	if err := vs.vars.Read(r); err != nil {
		return err
	}
	return nil
}

// saves variable state to db together with the batch of state updates
func (vs *variableState) Commit(addr *address.Address, b Batch) error {
	// TODO make it Badger-atomic transaction
	// TODO mark processed requests in db in separate index

	if err := b.(*batch).saveToDb(addr); err != nil {
		return err
	}
	if err := vs.saveToDb(addr); err != nil {
		return err
	}
	if err := MarkRequestsProcessed(b.RequestIds()); err != nil {
		return err
	}
	return nil
}

func StateExist(addr *address.Address) (bool, error) {
	dbase, err := database.GetVariableStateDB()
	if err != nil {
		return false, err
	}
	return dbase.Contains(database.DbKeyVariableState(addr))
}

// loads variable state and corresponding batch
func LoadVariableState(addr *address.Address) (VariableState, Batch, error) {
	dbase, err := database.GetVariableStateDB()
	if err != nil {
		return nil, nil, err
	}
	entry, err := dbase.Get(database.DbKeyVariableState(addr))
	if err != nil {
		return nil, nil, err
	}
	ret := &variableState{}
	if err = ret.Read(bytes.NewReader(entry.Value)); err != nil {
		return nil, nil, err
	}
	batch, err := LoadBatch(addr, ret.StateIndex())
	if err != nil {
		return nil, nil, err
	}
	return ret, batch, nil
}
