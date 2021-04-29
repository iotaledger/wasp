package state

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
	"io"
	"time"

	"github.com/iotaledger/wasp/packages/kv/buffered"
)

// stateUpdate implement StateUpdate interface
type stateUpdate struct {
	mutations *buffered.Mutations
}

// NewStateUpdate creates a state update with timestamp mutation, if provided
func NewStateUpdate(timestamp ...time.Time) *stateUpdate {
	ret := &stateUpdate{
		mutations: buffered.NewMutations(),
	}
	if len(timestamp) > 0 {
		ret.setTimestampMutation(timestamp[0])
	}
	return ret
}

func newStateUpdateFromReader(r io.Reader) (*stateUpdate, error) {
	ret := &stateUpdate{
		mutations: buffered.NewMutations(),
	}
	return ret, ret.Read(r)
}

func (su *stateUpdate) setTimestampMutation(ts time.Time) {
	su.mutations.Set(kv.Key(coreutil.StatePrefixTimestamp), codec.EncodeTime(ts))
}

func (su *stateUpdate) setBlockIndexMutation(blockIndex uint32) {
	su.mutations.Set(kv.Key(coreutil.StatePrefixBlockIndex), util.Uint32To4Bytes(blockIndex))
}

// StateUpdate

func (su *stateUpdate) Clone() StateUpdate {
	return &stateUpdate{
		mutations: su.mutations.Clone(),
	}
}

func (su *stateUpdate) Bytes() []byte {
	var buf bytes.Buffer
	_ = su.Write(&buf)
	return buf.Bytes()
}

func (su *stateUpdate) Hash() hashing.HashValue {
	return hashing.HashData(su.Bytes())
}

func (su *stateUpdate) String() string {
	ret := fmt.Sprintf("StateUpdate:: ts: %v, muts: [%s]", su.Timestamp(), su.mutations)
	return ret
}

func (su *stateUpdate) Timestamp() time.Time {
	timeBin, ok := su.mutations.Get(kv.Key(coreutil.StatePrefixTimestamp))
	if !ok {
		return time.Time{}
	}
	ret, ok, err := codec.DecodeTime(timeBin)
	if err != nil {
		panic(err)
	}
	return ret
}

func (su *stateUpdate) Mutations() *buffered.Mutations {
	return su.mutations
}

func (su *stateUpdate) Write(w io.Writer) error {
	if err := su.mutations.Write(w); err != nil {
		return err
	}
	return nil
}

func (su *stateUpdate) Read(r io.Reader) error {
	if err := su.mutations.Read(r); err != nil {
		return err
	}
	return nil
}
