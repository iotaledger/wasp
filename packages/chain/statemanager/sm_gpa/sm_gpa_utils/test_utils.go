// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sm_gpa_utils

import (
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/state"
)

func CheckBlockInStore(t require.TestingT, store state.Store, origBlock state.Block) {
	blockFromStore, err := store.BlockByTrieRoot(origBlock.TrieRoot())
	require.NoError(t, err)
	require.True(t, origBlock.Equals(blockFromStore))
}

// CheckStateInStores validates state consistency across stores
func CheckStateInStores(t require.TestingT, storeOrig, storeNew state.Store, commitment *state.L1Commitment) {
	origState, err := storeOrig.StateByTrieRoot(commitment.TrieRoot())
	require.NoError(t, err)
	CheckStateInStore(t, storeNew, origState)
}

func CheckStateInStore(t require.TestingT, store state.Store, origState state.State) {
	stateFromStore, err := store.StateByTrieRoot(origState.TrieRoot())
	require.NoError(t, err)
	require.True(t, origState.Equals(stateFromStore))
}
