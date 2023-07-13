// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/state"
)

func TestBlockSerialization(t *testing.T) {
	block1 := state.RandomBlock()
	b := block1.Bytes()
	block2, err := state.BlockFromBytes(b)
	require.NoError(t, err)
	require.Equal(t, block1, block2)
}
