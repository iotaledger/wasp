// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"time"

	"github.com/iotaledger/trie.go/common"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// stateDraft is the implementation of the StateDraft interface
type stateDraft struct {
	*buffered.BufferedKVStore
	baseState State
}

var _ StateDraft = &stateDraft{}

func newOriginStateDraft() *stateDraft {
	return &stateDraft{
		BufferedKVStore: buffered.NewBufferedKVStore(dict.Dict{}),
	}
}

func newStateDraft(baseState State, timestamp time.Time, prevL1Commitment *L1Commitment) *stateDraft {
	d := &stateDraft{
		BufferedKVStore: buffered.NewBufferedKVStore(baseState),
		baseState:       baseState,
	}
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

func (s *stateDraft) ChainID() *isc.ChainID {
	return loadChainIDFromState(s)
}

func (s *stateDraft) Timestamp() time.Time {
	ts, err := loadTimestampFromState(s)
	mustNoErr(err)
	return ts
}

func (s *stateDraft) BaseTrieRoot() common.VCommitment {
	if s.baseState == nil {
		return nil
	}
	return s.baseState.TrieRoot()
}

func (d *stateDraft) PreviousL1Commitment() *L1Commitment {
	return loadPrevL1CommitmentFromState(d)
}
