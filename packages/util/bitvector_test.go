// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestFixedSizeBitVector(t *testing.T) {
	bv := util.NewFixedSizeBitVector(10)
	require.Equal(t, []int{}, bv.AsInts())
	bv = bv.SetBits([]int{0, 3, 7, 8, 9})
	require.Equal(t, []int{0, 3, 7, 8, 9}, bv.AsInts())
}

func TestFixedSizeBitVectorSerialization(t *testing.T) {
	bv := util.NewFixedSizeBitVector(10).SetBits([]int{0, 3, 7, 8, 9})
	rwutil.ReadWriteTest(t, bv, util.NewFixedSizeBitVector(0))
	newBV := rwutil.BytesTest(t, bv, util.FixedSizeBitVectorFromBytes)
	require.Equal(t, bv.AsInts(), newBV.AsInts())
}
