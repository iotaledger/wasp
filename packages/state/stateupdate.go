package state

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"golang.org/x/xerrors"
)

// stateUpdateImpl implement Update interface
type stateUpdateImpl struct {
	mutations *buffered.Mutations
}

// NewStateUpdate creates a state update with timestamp mutation, if provided
func NewStateUpdate(timestamp ...time.Time) *stateUpdateImpl { //nolint:revive
	ret := &stateUpdateImpl{
		mutations: buffered.NewMutations(),
	}
	if len(timestamp) > 0 {
		ret.setTimestampMutation(timestamp[0])
	}
	return ret
}

func NewStateUpdateWithBlockLogValues(blockIndex uint32, timestamp time.Time, prevL1Commitment *L1Commitment) Update {
	ret := &stateUpdateImpl{
		mutations: buffered.NewMutations(),
	}
	ret.setBlockIndexMutation(blockIndex)
	ret.setTimestampMutation(timestamp)
	ret.setPrevL1CommitmentMutation(prevL1Commitment)
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

// Update

func (su *stateUpdateImpl) Clone() Update {
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
	ret, err := codec.DecodeUint32(blockIndexBin)
	if err != nil {
		return 0, false, err
	}
	return ret, true, nil
}

func (su *stateUpdateImpl) timestampMutation() (time.Time, bool, error) {
	timeBin, ok := su.mutations.Get(kv.Key(coreutil.StatePrefixTimestamp))
	if !ok {
		return time.Time{}, false, nil
	}
	ret, err := codec.DecodeTime(timeBin)
	if err != nil {
		return time.Time{}, false, err
	}
	return ret, true, nil
}

func (su *stateUpdateImpl) previousL1CommitmentMutation() (*L1Commitment, bool, error) {
	data, ok := su.mutations.Get(kv.Key(coreutil.StatePrefixPrevL1Commitment))
	if !ok {
		return nil, false, nil
	}

	ret, err := L1CommitmentFromBytes(data)
	if err != nil {
		return nil, false, err
	}
	return &ret, true, nil
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
		ts = t.String()
	}
	bi := none
	idx, ok, err := su.stateIndexMutation()
	if err != nil {
		bi = fmt.Sprintf("(%v)", err)
	} else if ok {
		bi = fmt.Sprintf("%d", idx)
	}
	return fmt.Sprintf("Update:: ts: %s, blockIndex: %s muts: [%+v]", ts, bi, su.mutations)
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

// findPrevStateCommitmentMutation goes backward and searches for the 'set' mutation of the previous state hash
// Return NilHash if not found
func findPrevStateCommitmentMutation(stateUpdate *stateUpdateImpl) (*L1Commitment, error) {
	ret, exists, err := stateUpdate.previousL1CommitmentMutation()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	return ret, nil
}

func (su *stateUpdateImpl) setTimestampMutation(ts time.Time) {
	su.mutations.Set(kv.Key(coreutil.StatePrefixTimestamp), codec.EncodeTime(ts))
}

func (su *stateUpdateImpl) setBlockIndexMutation(blockIndex uint32) {
	su.mutations.Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.EncodeUint32(blockIndex))
}

func (su *stateUpdateImpl) setPrevL1CommitmentMutation(prevL1Commitment *L1Commitment) {
	su.mutations.Set(kv.Key(coreutil.StatePrefixPrevL1Commitment), prevL1Commitment.Bytes())
}
