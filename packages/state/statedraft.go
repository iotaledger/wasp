// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"time"

	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// stateDraft is the implementation of the StateDraft interface
type stateDraft struct {
	*buffered.BufferedKVStore
	baseL1Commitment *L1Commitment
	baseState        State
}

var _ StateDraft = &stateDraft{}

func newOriginStateDraft() *stateDraft {
	return &stateDraft{
		BufferedKVStore: buffered.NewBufferedKVStore(dict.Dict{}),
	}
}

func newEmptyStateDraft(prevL1Commitment *L1Commitment, baseState State) *stateDraft {
	return &stateDraft{
		BufferedKVStore:  buffered.NewBufferedKVStore(baseState),
		baseL1Commitment: prevL1Commitment,
		baseState:        baseState,
	}
}

func newStateDraft(timestamp time.Time, prevL1Commitment *L1Commitment, baseState State) *stateDraft {
	d := newEmptyStateDraft(prevL1Commitment, baseState)
	d.Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.EncodeUint32(baseState.BlockIndex()+1))
	d.Set(kv.Key(coreutil.StatePrefixTimestamp), codec.EncodeTime(timestamp))
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
