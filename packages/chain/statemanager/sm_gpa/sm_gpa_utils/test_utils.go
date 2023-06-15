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
