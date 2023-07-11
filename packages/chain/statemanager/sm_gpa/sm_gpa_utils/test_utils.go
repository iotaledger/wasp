// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sm_gpa_utils

import (
	"bytes"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
)

func CheckBlockInStore(t require.TestingT, store state.Store, origBlock state.Block) {
	blockFromStore, err := store.BlockByTrieRoot(origBlock.TrieRoot())
	require.NoError(t, err)
	CheckBlocksEqual(t, origBlock, blockFromStore)
}

// NOTE: this function should not exist. state.Block should have Equals method
func CheckBlocksEqual(t require.TestingT, block1, block2 state.Block) {
	require.Equal(t, block1.StateIndex(), block2.StateIndex())
	require.True(t, block1.PreviousL1Commitment().Equals(block2.PreviousL1Commitment()))
	require.True(t, block1.L1Commitment().Equals(block2.L1Commitment()))
	// NOTE: having separate sentences instead of require.True(t, BlocksEqual(block1, block2))
	//       to have a more precise location of error in logs.
}

func BlocksEqual(block1, block2 state.Block) bool {
	return block1.StateIndex() == block2.StateIndex() &&
		block1.PreviousL1Commitment().Equals(block2.PreviousL1Commitment()) &&
		block1.L1Commitment().Equals(block2.L1Commitment())
}

// NOTE: this function should not exist. state.Block should have Equals method
func CheckBlocksDifferent(t require.TestingT, block1, block2 state.Block) {
	require.False(t, block1.Hash().Equals(block2.Hash()))
}

// -----------------------------------------------------------------------------
func CheckStateInStores(t require.TestingT, storeOrig, storeNew state.Store, commitment *state.L1Commitment) {
	origState, err := storeOrig.StateByTrieRoot(commitment.TrieRoot())
	require.NoError(t, err)
	CheckStateInStore(t, storeNew, origState)
}

func CheckStateInStore(t require.TestingT, store state.Store, origState state.State) {
	stateFromStore, err := store.StateByTrieRoot(origState.TrieRoot())
	require.NoError(t, err)
	require.True(t, origState.TrieRoot().Equals(stateFromStore.TrieRoot()))
	require.Equal(t, origState.BlockIndex(), stateFromStore.BlockIndex())
	require.Equal(t, origState.Timestamp(), stateFromStore.Timestamp())
	require.True(t, origState.PreviousL1Commitment().Equals(stateFromStore.PreviousL1Commitment()))
	commonState := getCommonState(origState, stateFromStore)
	for _, entry := range commonState {
		require.True(t, bytes.Equal(entry.value1, entry.value2))
	}
}

// NOTE: this function should not exist. state.State should have Equals method
func StatesEqual(state1, state2 state.State) bool {
	if !state1.TrieRoot().Equals(state2.TrieRoot()) ||
		state1.BlockIndex() != state2.BlockIndex() ||
		state1.Timestamp() != state2.Timestamp() ||
		!state1.PreviousL1Commitment().Equals(state2.PreviousL1Commitment()) {
		return false
	}
	commonState := getCommonState(state1, state2)
	for _, entry := range commonState {
		if !bytes.Equal(entry.value1, entry.value2) {
			return false
		}
	}
	return true
}

type commonEntry struct {
	value1 []byte
	value2 []byte
}

func getCommonState(state1, state2 state.State) map[kv.Key]*commonEntry {
	result := make(map[kv.Key]*commonEntry)
	iterateFun := func(iterState state.State, setValueFun func(*commonEntry, []byte)) {
		iterState.Iterate(kv.EmptyPrefix, func(key kv.Key, value []byte) bool {
			entry, ok := result[key]
			if !ok {
				entry = &commonEntry{}
				result[key] = entry
			}
			setValueFun(entry, value)
			return true
		})
	}
	iterateFun(state1, func(entry *commonEntry, value []byte) { entry.value1 = value })
	iterateFun(state2, func(entry *commonEntry, value []byte) { entry.value2 = value })
	return result
}
