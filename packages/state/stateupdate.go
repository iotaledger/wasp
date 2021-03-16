package state

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"io"

	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/util"
)

type stateUpdate struct {
	requestID ledgerstate.OutputID
	timestamp int64
	mutations buffered.MutationSequence
}

func NewStateUpdate(reqid ledgerstate.OutputID) StateUpdate {
	return &stateUpdate{
		requestID: reqid,
		mutations: buffered.NewMutationSequence(),
	}
}

func NewStateUpdateRead(r io.Reader) (StateUpdate, error) {
	ret := NewStateUpdate(ledgerstate.OutputID{}).(*stateUpdate)
	return ret, ret.Read(r)
}

// StateUpdate

func (su *stateUpdate) Clone() StateUpdate {
	ret := *su
	ret.mutations = su.mutations.Clone()
	return &ret
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

func (su *stateUpdate) RequestID() ledgerstate.OutputID {
	return su.requestID
}

func (su *stateUpdate) Mutations() buffered.MutationSequence {
	return su.mutations
}

func (su *stateUpdate) Write(w io.Writer) error {
	if _, err := w.Write(su.requestID.Bytes()); err != nil {
		return err
	}
	if err := su.mutations.Write(w); err != nil {
		return err
	}
	return util.WriteUint64(w, uint64(su.timestamp))
}

func (su *stateUpdate) Read(r io.Reader) error {
	var rid ledgerstate.OutputID
	if n, err := r.Read(rid[:]); err != nil || n != ledgerstate.OutputIDLength {
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
