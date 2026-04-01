// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/coreutil"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/buffered"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
	"github.com/iotaledger/wasp/v2/packages/kv/dict"
	"github.com/iotaledger/wasp/v2/packages/vm/core/root"
)

// stateDraft is the implementation of the StateDraft interface
type stateDraft struct {
	*buffered.BufferedKVStore
	baseL1Commitment *L1Commitment
	baseState        State
}

var _ StateDraft = &stateDraft{}

// OriginTimestamp is the timestamp stored in the origin ISC block.
// Cannot use time.Unix(0, 0) because bcs-go would encode it as the zero time (time.Time{} != time.Unix(0, 0)).
// Also cannot use the zero time because it is before the unix epoch, and that introduces subtle bugs.
// TODO: make configurable
var OriginTimestamp = time.Unix(1, 0)

func newOriginStateDraft() *stateDraft {
	d := &stateDraft{
		BufferedKVStore: buffered.NewBufferedKVStore(dict.Dict{}),
	}
	d.Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.Encode(uint32(0)))
	d.Set(kv.Key(coreutil.StatePrefixTimestamp), codec.Encode(OriginTimestamp))
	return d
}

func newEmptyStateDraft(prevL1Commitment *L1Commitment, baseState State) *stateDraft {
	return &stateDraft{
		BufferedKVStore:  buffered.NewBufferedKVStore(baseState),
		baseL1Commitment: prevL1Commitment,
		baseState:        baseState,
	}
}

func newStateDraft(timestamp time.Time, prevL1Commitment *L1Commitment, baseState State) *stateDraft {
	if !timestamp.IsZero() && timestamp.Before(time.Unix(0, 0)) {
		panic(fmt.Sprintf("timestamp %s is before unix epoch", timestamp.String()))
	}
	d := newEmptyStateDraft(prevL1Commitment, baseState)
	d.Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.Encode(baseState.BlockIndex()+1))
	d.Set(kv.Key(coreutil.StatePrefixTimestamp), codec.Encode(timestamp))
	d.Set(kv.Key(coreutil.StatePrefixPrevL1Commitment), prevL1Commitment.Bytes())
	return d
}

func (s *stateDraft) Mutations() *buffered.Mutations {
	return s.BufferedKVStore.Mutations()
}

func (s *stateDraft) BlockIndex() uint32 {
	return loadBlockIndexFromState(s)
}

func (s *stateDraft) Timestamp() time.Time {
	ts, err := loadTimestampFromState(s)
	mustNoErr(err)
	return ts
}

func (s *stateDraft) BaseL1Commitment() *L1Commitment {
	return s.baseL1Commitment
}

func (s *stateDraft) PreviousL1Commitment() *L1Commitment {
	return loadPrevL1CommitmentFromState(s)
}

func (s *stateDraft) SchemaVersion() isc.SchemaVersion {
	return root.NewStateReaderFromChainState(s).GetSchemaVersion()
}
