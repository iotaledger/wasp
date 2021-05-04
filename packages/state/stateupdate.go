package state

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
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

func NewStateUpdateWithBlockIndexMutation(blockIndex uint32, timestamp ...time.Time) *stateUpdate {
	ret := &stateUpdate{
		mutations: buffered.NewMutations(),
	}
	ret.setBlockIndexMutation(blockIndex)
	if len(timestamp) > 0 {
		ret.setTimestampMutation(timestamp[0])
	}
	return ret
}

func newStateUpdateFromMutations(mutations *buffered.Mutations) *stateUpdate {
	return &stateUpdate{
		mutations: mutations.Clone(),
	}
}

func newStateUpdateFromReader(r io.Reader) (*stateUpdate, error) {
	ret := &stateUpdate{
		mutations: buffered.NewMutations(),
	}
	return ret, ret.Read(r)
}

func (su *stateUpdate) Mutations() *buffered.Mutations {
	return su.mutations
}

// StateUpdate

func (su *stateUpdate) Clone() StateUpdate {
	return su.clone()
}

func (su *stateUpdate) clone() *stateUpdate {
	return &stateUpdate{
		mutations: su.mutations.Clone(),
	}
}

func (su *stateUpdate) Bytes() []byte {
	var buf bytes.Buffer
	_ = su.Write(&buf)
	return buf.Bytes()
}

func (su *stateUpdate) TimestampMutation() (time.Time, bool) {
	timeBin, ok := su.mutations.Get(kv.Key(coreutil.StatePrefixTimestamp))
	if !ok {
		return time.Time{}, false
	}
	ret, ok, err := codec.DecodeTime(timeBin)
	if err != nil {
		return time.Time{}, false
	}
	return ret, true
}

func (su *stateUpdate) StateIndexMutation() (uint32, bool) {
	blockIndexBin, ok := su.mutations.Get(kv.Key(coreutil.StatePrefixBlockIndex))
	if !ok {
		return 0, false
	}
	ret, err := util.Uint64From8Bytes(blockIndexBin)
	if err != nil {
		return 0, false
	}
	if int(ret) > util.MaxUint32 {
		return 0, false
	}
	return uint32(ret), true
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

func (su *stateUpdate) String() string {
	ts := "(none)"
	if t, ok := su.TimestampMutation(); ok {
		ts = fmt.Sprintf("%v", t)
	}
	bi := "(none)"
	if t, ok := su.StateIndexMutation(); ok {
		bi = fmt.Sprintf("%d", t)
	}

	ret := fmt.Sprintf("StateUpdate:: ts: %s, blockIndex: %s muts: [%+v]", ts, bi, su.mutations)
	return ret
}

// findBlockIndexMutation goes backward and searches for the 'set' mutation of the blockIndex
func findBlockIndexMutation(stateUpdates []*stateUpdate) (uint32, error) {
	if len(stateUpdates) == 0 {
		return 0, xerrors.New("findBlockIndexMutation: no state updates were found")
	}
	for i := len(stateUpdates) - 1; i >= 0; i-- {
		blockIndexBin, exists := stateUpdates[i].Mutations().Get(kv.Key(coreutil.StatePrefixBlockIndex))
		if !exists || blockIndexBin == nil {
			continue
		}
		blockIndex, err := util.Uint64From8Bytes(blockIndexBin)
		if err != nil {
			return 0, xerrors.Errorf("findBlockIndexMutation: %w", err)
		}
		if int(blockIndex) > util.MaxUint32 {
			return 0, xerrors.Errorf("findBlockIndexMutation: wrong block index value")
		}
		return uint32(blockIndex), nil
	}
	return 0, xerrors.Errorf("findBlockIndexMutation: state index mutation wasn't found in the block")
}

// findBlockIndexMutation goes backward and searches for the 'set' mutation of the blockIndex
// Return zero time if not found
func findTimestampMutation(stateUpdates []*stateUpdate) (time.Time, error) {
	if len(stateUpdates) == 0 {
		return time.Time{}, xerrors.New("findTimestampMutation: no state updates were found")
	}
	for i := len(stateUpdates) - 1; i >= 0; i-- {
		timestampBin, exists := stateUpdates[i].Mutations().Get(kv.Key(coreutil.StatePrefixTimestamp))
		if !exists || timestampBin == nil {
			continue
		}
		ts, ok, err := codec.DecodeTime(timestampBin)
		if !ok || err != nil {
			return time.Time{}, err
		}
		return ts, nil
	}
	return time.Time{}, nil
}

func (su *stateUpdate) setTimestampMutation(ts time.Time) {
	su.mutations.Set(kv.Key(coreutil.StatePrefixTimestamp), codec.EncodeTime(ts))
}

func (su *stateUpdate) setBlockIndexMutation(blockIndex uint32) {
	su.mutations.Set(kv.Key(coreutil.StatePrefixBlockIndex), util.Uint64To8Bytes(uint64(blockIndex)))
}
