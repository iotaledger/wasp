package state

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/variables"
	"io"
)

// this is fake implementation of the VariableState and StateUpdate
// intended for testing
// VariableState is just a hashed value and stateIndex (not even needed for testing).
// State transition by applying state update to Variable state with Apply function
// is just a hashing of the previous VariableState
// the state update is empty always. The whole information about the state update is contained
// in the key: color and state index

type mockVariableState struct {
	address    *address.Address
	stateIndex uint32
	merkleHash hashing.HashValue
	vars       variables.Variables
}

type mockStateUpdate struct {
	address    *address.Address // persist in key
	stateIndex uint32           // persist in key
	stateTxId  valuetransaction.ID
	vars       variables.Variables
}

// StateUpdate

func NewStateUpdate(addr *address.Address, stateIndex uint32) StateUpdate {
	return &mockStateUpdate{
		address:    addr,
		stateIndex: stateIndex,
		vars:       variables.NewVariables(),
	}
}

// StateUpdate

func (se *mockStateUpdate) Address() *address.Address {
	return se.address
}

func (se *mockStateUpdate) StateIndex() uint32 {
	return se.stateIndex
}

func (su *mockStateUpdate) StateTransactionId() valuetransaction.ID {
	return su.stateTxId
}

func (su *mockStateUpdate) SetStateTransactionId(vtxId valuetransaction.ID) {
	su.stateTxId = vtxId
}

func (su *mockStateUpdate) Variables() variables.Variables {
	return su.vars
}

func (su *mockStateUpdate) Write(w io.Writer) error {
	if _, err := w.Write(su.address.Bytes()); err != nil {
		return err
	}
	if err := util.WriteUint32(w, su.stateIndex); err != nil {
		return err
	}
	if _, err := w.Write(su.stateTxId[:]); err != nil {
		return err
	}
	if err := su.vars.Write(w); err != nil {
		return err
	}
	return nil
}

func (su *mockStateUpdate) Read(r io.Reader) error {
	su.address = new(address.Address)
	if _, err := r.Read(su.address.Bytes()); err != nil {
		return err
	}
	if err := util.ReadUint32(r, &su.stateIndex); err != nil {
		return err
	}
	if _, err := r.Read(su.stateTxId[:]); err != nil {
		return err
	}
	if err := su.vars.Read(r); err != nil {
		return err
	}
	return nil
}

// VariableState

func NewMockVariableState(stateIndex uint32, hash hashing.HashValue) VariableState {
	return &mockVariableState{
		stateIndex: stateIndex,
		merkleHash: hash,
	}
}

func (vs *mockVariableState) StateIndex() uint32 {
	return vs.stateIndex
}

func (vs *mockVariableState) Apply(stateUpdate StateUpdate) VariableState {
	merkleHash := hashing.NilHash
	if vs != nil {
		merkleHash = hashing.HashData(vs.merkleHash.Bytes())
	}
	return NewMockVariableState(stateUpdate.StateIndex(), *merkleHash)
}

func (vs *mockVariableState) Variables() variables.Variables {
	return vs.vars
}

func (vs *mockVariableState) Write(w io.Writer) error {
	if _, err := w.Write(util.Uint32To4Bytes(vs.stateIndex)); err != nil {
		return err
	}
	if _, err := w.Write(vs.merkleHash.Bytes()); err != nil {
		return err
	}
	if err := vs.vars.Write(w); err != nil {
		return err
	}
	return nil
}

func (vs *mockVariableState) Read(r io.Reader) error {
	if err := util.ReadUint32(r, &vs.stateIndex); err != nil {
		return err
	}
	if _, err := r.Read(vs.merkleHash.Bytes()); err != nil {
		return err
	}
	if err := vs.vars.Read(r); err != nil {
		return err
	}
	return nil
}

func CreateOriginVariableState(stateUpdate StateUpdate) VariableState {
	return VariableState(nil).Apply(stateUpdate)
}
