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
func NewStateUpdate(timestamp ...time.Time) *stateUpdateImpl { //nolint
	ret := &stateUpdateImpl{
		mutations: buffered.NewMutations(),
	}
	if len(timestamp) > 0 {
		ret.setTimestampMutation(timestamp[0])
	}
	return ret
}

func NewStateUpdateWithBlocklogValues(blockIndex uint32, timestamp time.Time, prevStateHash hashing.HashValue) StateUpdate {
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

func (su *stateUpdateImpl) stateIndexMutation() (uint32, bool, error) {
	blockIndexBin, ok := su.mutations.Get(kv.Key(coreutil.StatePrefixBlockIndex))
	if !ok {
		return 0, false, nil
	}
	ret, err := util.Uint64From8Bytes(blockIndexBin)
	if err != nil {
		return 0, false, err
	}
	if int(ret) > util.MaxUint32 {
		return 0, false, xerrors.New("wrong state index")
	}
	return uint32(ret), true, nil
}

func (su *stateUpdateImpl) timestampMutation() (time.Time, bool, error) {
	timeBin, ok := su.mutations.Get(kv.Key(coreutil.StatePrefixTimestamp))
	if !ok {
		return time.Time{}, false, nil
	}
	ret, _, err := codec.DecodeTime(timeBin)
	if err != nil {
		return time.Time{}, false, err
	}
	return ret, true, nil
}

func (su *stateUpdateImpl) previousStateHashMutation() (hashing.HashValue, bool, error) {
	hashBin, ok := su.mutations.Get(kv.Key(coreutil.StatePrefixPrevStateHash))
	if !ok {
		return hashing.NilHash, false, nil
	}
	ret, _, err := codec.DecodeHashValue(hashBin)
	if err != nil {
		return hashing.NilHash, false, err
	}
	return ret, true, nil
}

func (su *stateUpdateImpl) Write(w io.Writer) error {
	return su.mutations.Write(w)
}

func (su *stateUpdateImpl) Read(r io.Reader) error {
	return su.mutations.Read(r)
}

const none = "(none)"

func (su *stateUpdateImpl) String() string {
	ts := none
	t, ok, err := su.timestampMutation()
	if err != nil {
		ts = fmt.Sprintf("(%v)", err)
	} else if ok {
		ts = fmt.Sprintf("%v", t)
	}
	bi := none
	idx, ok, err := su.stateIndexMutation()
	if err != nil {
		bi = fmt.Sprintf("(%v)", err)
	} else if ok {
		bi = fmt.Sprintf("%d", idx)
	}
	ph := none
	h, ok, err := su.previousStateHashMutation()
	if err != nil {
		ph = fmt.Sprintf("(%v)", err)
	} else if ok {
		ph = h.Base58()
	}
	return fmt.Sprintf("StateUpdate:: ts: %s, blockIndex: %s prevStateHash: %s muts: [%+v]", ts, bi, ph, su.mutations)
}

// findBlockIndexMutation goes backward and searches for the 'set' mutation of the blockIndex
func findBlockIndexMutation(stateUpdate *stateUpdateImpl) (uint32, error) {
	bi, exists, err := stateUpdate.stateIndexMutation()
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, xerrors.Errorf("findBlockIndexMutation: state index mutation wasn't found in the block")
	}
	return bi, nil
}

// findTimestampMutation goes backward and searches for the 'set' mutation of the timestamp
// Return zero time if not found
func findTimestampMutation(stateUpdate *stateUpdateImpl) (time.Time, error) {
	ts, exists, err := stateUpdate.timestampMutation()
	if err != nil {
		return time.Time{}, err
	}
	if !exists {
		return time.Time{}, nil
	}
	return ts, nil
}

// findPrevStateHashMutation goes backward and searches for the 'set' mutation of the previous state hash
// Return NilHash if not found
func findPrevStateHashMutation(stateUpdate *stateUpdateImpl) (hashing.HashValue, error) {
	h, exists, err := stateUpdate.previousStateHashMutation()
	if err != nil {
		return hashing.NilHash, err
	}
	if !exists {
		return hashing.NilHash, nil
	}
	return h, nil
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
