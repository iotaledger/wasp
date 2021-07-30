package state

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/iotaledger/wasp/packages/hashing"

	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

// stateUpdateImpl implement StateUpdate interface
type stateUpdateImpl struct {
	mutations *buffered.Mutations
}

// NewStateUpdate creates a state update with timestamp mutation, if provided
func NewStateUpdate(timestamp ...time.Time) *stateUpdateImpl {
	ret := &stateUpdateImpl{
		mutations: buffered.NewMutations(),
	}
	if len(timestamp) > 0 {
		ret.setTimestampMutation(timestamp[0])
	}
	return ret
}

func NewStateUpdateWithBlocklogValues(blockIndex uint32, timestamp time.Time, prevStateHash hashing.HashValue) *stateUpdateImpl {
	ret := &stateUpdateImpl{
		mutations: buffered.NewMutations(),
	}
	ret.setBlockIndexMutation(blockIndex)
	ret.setTimestampMutation(timestamp)
	ret.setPrevStateHashMutation(prevStateHash)
	return ret
}

func newStateUpdateFromReader(r io.Reader) (*stateUpdateImpl, error) {
	ret := &stateUpdateImpl{
		mutations: buffered.NewMutations(),
	}
	err := ret.Read(r)
	return ret, err
}

func (su *stateUpdateImpl) Mutations() *buffered.Mutations {
	return su.mutations
}

// StateUpdate

func (su *stateUpdateImpl) Clone() StateUpdate {
	return su.clone()
}

func (su *stateUpdateImpl) clone() *stateUpdateImpl {
	return &stateUpdateImpl{
		mutations: su.mutations.Clone(),
	}
}

func (su *stateUpdateImpl) Bytes() []byte {
	var buf bytes.Buffer
	_ = su.Write(&buf)
	return buf.Bytes()
}

func (su *stateUpdateImpl) StateIndexMutation() (uint32, bool) {
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

func (su *stateUpdateImpl) TimestampMutation() (time.Time, bool) {
	timeBin, ok := su.mutations.Get(kv.Key(coreutil.StatePrefixTimestamp))
	if !ok {
		return time.Time{}, false
	}
	ret, _, err := codec.DecodeTime(timeBin)
	if err != nil {
		return time.Time{}, false
	}
	return ret, true
}

func (su *stateUpdateImpl) PreviousStateHashMutation() (hashing.HashValue, bool) {
	hashBin, ok := su.mutations.Get(kv.Key(coreutil.StatePrefixPrevStateHash))
	if !ok {
		return hashing.NilHash, false
	}
	ret, _, err := codec.DecodeHashValue(hashBin)
	if err != nil {
		return hashing.NilHash, false
	}
	return ret, true
}

func (su *stateUpdateImpl) Write(w io.Writer) error {
	return su.mutations.Write(w)
}

func (su *stateUpdateImpl) Read(r io.Reader) error {
	return su.mutations.Read(r)
}

func (su *stateUpdateImpl) String() string {
	ts := "(none)"
	if t, ok := su.TimestampMutation(); ok {
		ts = fmt.Sprintf("%v", t)
	}
	bi := "(none)"
	if t, ok := su.StateIndexMutation(); ok {
		bi = fmt.Sprintf("%d", t)
	}
	ph := "(none)"
	if h, ok := su.PreviousStateHashMutation(); ok {
		ph = h.Base58()
	}
	return fmt.Sprintf("StateUpdate:: ts: %s, blockIndex: %s prevStateHash: %s muts: [%+v]", ts, bi, ph, su.mutations)
}

// findBlockIndexMutation goes backward and searches for the 'set' mutation of the blockIndex
func findBlockIndexMutation(stateUpdates []*stateUpdateImpl) (uint32, error) {
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

// findTimestampMutation goes backward and searches for the 'set' mutation of the timestamp
// Return zero time if not found
func findTimestampMutation(stateUpdates []*stateUpdateImpl) (time.Time, error) {
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

// findPrevStateHashMutation goes backward and searches for the 'set' mutation of the previous state hash
// Return NilHash if not found
func findPrevStateHashMutation(stateUpdates []*stateUpdateImpl) (hashing.HashValue, error) {
	if len(stateUpdates) == 0 {
		return hashing.NilHash, xerrors.New("findPrevStateHashMutation: no state updates were found")
	}
	for i := len(stateUpdates) - 1; i >= 0; i-- {
		hashBin, exists := stateUpdates[i].Mutations().Get(kv.Key(coreutil.StatePrefixPrevStateHash))
		if !exists || hashBin == nil {
			continue
		}
		prevStateHash, ok, err := codec.DecodeHashValue(hashBin)
		if !ok || err != nil {
			return hashing.NilHash, err
		}
		return prevStateHash, nil
	}
	return hashing.NilHash, nil
}

func (su *stateUpdateImpl) setTimestampMutation(ts time.Time) {
	su.mutations.Set(kv.Key(coreutil.StatePrefixTimestamp), codec.EncodeTime(ts))
}

func (su *stateUpdateImpl) setBlockIndexMutation(blockIndex uint32) {
	su.mutations.Set(kv.Key(coreutil.StatePrefixBlockIndex), util.Uint64To8Bytes(uint64(blockIndex)))
}

func (su *stateUpdateImpl) setPrevStateHashMutation(prevStateHash hashing.HashValue) {
	su.mutations.Set(kv.Key(coreutil.StatePrefixPrevStateHash), prevStateHash.Bytes())
}
