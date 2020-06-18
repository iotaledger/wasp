package state

import (
	"fmt"
	"io"

	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/variables"
)

type stateUpdate struct {
	batchIndex uint16
	requestId  sctransaction.RequestId
	timestamp  int64
	mutations  []variables.Mutation
}

func NewStateUpdate(reqid *sctransaction.RequestId) StateUpdate {
	var req sctransaction.RequestId
	if reqid != nil {
		req = *reqid
	}
	return &stateUpdate{
		requestId: req,
		mutations: make([]variables.Mutation, 0),
	}
}

func NewStateUpdateRead(r io.Reader) (StateUpdate, error) {
	ret := NewStateUpdate(nil).(*stateUpdate)
	return ret, ret.Read(r)
}

// StateUpdate

func (su *stateUpdate) String() string {
	ret := fmt.Sprintf("reqid: %s, ts: %d", su.requestId.String(), su.Timestamp())
	for _, mut := range su.mutations {
		ret += fmt.Sprintf(" [%s]", mut.String())
	}
	return ret
}

func (su *stateUpdate) AddMutation(mut variables.Mutation) {
	su.mutations = append(su.mutations, mut)
}

func (su *stateUpdate) Timestamp() int64 {
	return su.timestamp
}

func (su *stateUpdate) WithTimestamp(ts int64) StateUpdate {
	su.timestamp = ts
	return su
}

func (su *stateUpdate) RequestId() *sctransaction.RequestId {
	return &su.requestId
}

func (su *stateUpdate) Mutations() []variables.Mutation {
	return su.mutations
}

func (su *stateUpdate) Write(w io.Writer) error {
	if _, err := w.Write(su.requestId[:]); err != nil {
		return err
	}
	n := len(su.mutations)
	if n > util.MaxUint16 {
		return fmt.Errorf("StateUpdate has too many mutations")
	}
	if err := util.WriteUint16(w, uint16(n)); err != nil {
		return err
	}
	for _, mut := range su.mutations {
		if err := mut.Write(w); err != nil {
			return err
		}
	}
	return util.WriteUint64(w, uint64(su.timestamp))
}

func (su *stateUpdate) Read(r io.Reader) error {
	if _, err := r.Read(su.requestId[:]); err != nil {
		return err
	}
	var n uint16
	if err := util.ReadUint16(r, &n); err != nil {
		return err
	}
	su.mutations = make([]variables.Mutation, n)
	for i := uint16(0); i < n; i++ {
		mut, err := variables.ReadMutation(r)
		if err != nil {
			return err
		}
		su.mutations[i] = mut
	}
	var ts uint64
	if err := util.ReadUint64(r, &ts); err != nil {
		return err
	}
	su.timestamp = int64(ts)
	return nil
}
