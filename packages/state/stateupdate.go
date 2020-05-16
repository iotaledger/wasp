package state

import (
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/variables"
	"io"
)

type stateUpdate struct {
	stateIndex uint32
	batchIndex uint16
	stateTxId  valuetransaction.ID
	requestId  sctransaction.RequestId
	vars       variables.Variables
}

type NewStateUpdateParams struct {
	StateIndex uint32
	BatchIndex uint16
	StateTxId  valuetransaction.ID
	RequestId  sctransaction.RequestId
}

func NewStateUpdate(par NewStateUpdateParams) StateUpdate {
	return &stateUpdate{
		stateIndex: par.StateIndex,
		batchIndex: par.BatchIndex,
		stateTxId:  par.StateTxId,
		requestId:  par.RequestId,
		vars:       variables.New(nil),
	}
}

func NewStateUpdateRead(r io.Reader) (StateUpdate, error) {
	ret := stateUpdate{}
	return &ret, ret.Read(r)
}

// StateUpdate

func (se *stateUpdate) StateIndex() uint32 {
	return se.stateIndex
}

func (su *stateUpdate) RequestId() *sctransaction.RequestId {
	return &su.requestId
}

func (su *stateUpdate) BatchIndex() uint16 {
	return su.batchIndex
}

func (su *stateUpdate) Hash() *hashing.HashValue {
	return hashing.HashData(hashing.MustBytes(su))
}

func (su *stateUpdate) StateTransactionId() valuetransaction.ID {
	return su.stateTxId
}

func (su *stateUpdate) SetStateTransactionId(vtxId valuetransaction.ID) {
	su.stateTxId = vtxId
}

func (su *stateUpdate) Variables() variables.Variables {
	return su.vars
}

func (su *stateUpdate) Write(w io.Writer) error {
	if err := util.WriteUint32(w, su.stateIndex); err != nil {
		return err
	}
	if err := util.WriteUint16(w, su.batchIndex); err != nil {
		return err
	}
	if _, err := w.Write(su.stateTxId[:]); err != nil {
		return err
	}
	if _, err := w.Write(su.requestId[:]); err != nil {
		return err
	}
	if err := su.vars.Write(w); err != nil {
		return err
	}
	return nil
}

func (su *stateUpdate) Read(r io.Reader) error {
	if err := util.ReadUint32(r, &su.stateIndex); err != nil {
		return err
	}
	if err := util.ReadUint16(r, &su.batchIndex); err != nil {
		return err
	}
	if _, err := r.Read(su.stateTxId[:]); err != nil {
		return err
	}
	if _, err := r.Read(su.requestId[:]); err != nil {
		return err
	}
	if err := su.vars.Read(r); err != nil {
		return err
	}
	return nil
}
