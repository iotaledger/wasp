package state

import (
	"fmt"
	"io"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/util"
)

type stateUpdate struct {
	batchIndex uint16
	requestID  coretypes.RequestID
	timestamp  int64
	mutations  buffered.MutationSequence
}

func NewStateUpdate(reqid *coretypes.RequestID) StateUpdate {
	var req coretypes.RequestID
	if reqid != nil {
		req = *reqid
	}
	return &stateUpdate{
		requestID: req,
		mutations: buffered.NewMutationSequence(),
	}
}

func NewStateUpdateRead(r io.Reader) (StateUpdate, error) {
	ret := NewStateUpdate(nil).(*stateUpdate)
	return ret, ret.Read(r)
}

// StateUpdate

func (su *stateUpdate) Clear() {
	su.mutations = buffered.NewMutationSequence()
}

func (su *stateUpdate) String() string {
	ret := fmt.Sprintf("reqid: %s, ts: %d, muts: [%s]", su.requestID.String(), su.Timestamp(), su.mutations)
	return ret
}

func (su *stateUpdate) Timestamp() int64 {
	return su.timestamp
}

func (su *stateUpdate) WithTimestamp(ts int64) StateUpdate {
	su.timestamp = ts
	return su
}

func (su *stateUpdate) RequestID() *coretypes.RequestID {
	return &su.requestID
}

func (su *stateUpdate) Mutations() buffered.MutationSequence {
	return su.mutations
}

func (su *stateUpdate) Write(w io.Writer) error {
	if err := su.requestID.Write(w); err != nil {
		return err
	}
	if err := su.mutations.Write(w); err != nil {
		return err
	}
	return util.WriteUint64(w, uint64(su.timestamp))
}

func (su *stateUpdate) Read(r io.Reader) error {
	if err := su.requestID.Read(r); err != nil {
		return err
	}
	if err := su.mutations.Read(r); err != nil {
		return err
	}
	var ts uint64
	if err := util.ReadUint64(r, &ts); err != nil {
		return err
	}
	su.timestamp = int64(ts)
	return nil
}
