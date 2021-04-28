package state

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"io"
	"time"

	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/util"
)

//

type stateUpdate struct {
	timestamp int64
	mutations *buffered.Mutations
}

func NewStateUpdate(timestamp time.Time) StateUpdate {
	ret := &stateUpdate{
		mutations: buffered.NewMutations(),
	}
	ret.mutations.Set(kv.Key(coreutil.StatePrefixTimestamp), codec.EncodeTime(timestamp))
	return ret
}

func NewStateUpdateRead(r io.Reader) (StateUpdate, error) {
	ret := NewStateUpdate().(*stateUpdate)
	return ret, ret.Read(r)
}

// StateUpdate

func (su *stateUpdate) Clone() StateUpdate {
	ret := *su
	ret.mutations = su.mutations.Clone()
	return &ret
}

func (su *stateUpdate) String() string {
	ret := fmt.Sprintf("StateUpdate:: ts: %d, muts: [%s]", su.Timestamp(), su.mutations)
	return ret
}

func (su *stateUpdate) Timestamp() int64 {
	return su.timestamp
}

func (su *stateUpdate) WithTimestamp(ts int64) StateUpdate {
	su.timestamp = ts
	return su
}

func (su *stateUpdate) Mutations() *buffered.Mutations {
	return su.mutations
}

func (su *stateUpdate) Write(w io.Writer) error {
	if err := su.mutations.Write(w); err != nil {
		return err
	}
	return util.WriteUint64(w, uint64(su.timestamp))
}

func (su *stateUpdate) Read(r io.Reader) error {
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
