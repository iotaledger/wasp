// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc/isctest"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"

	"github.com/iotaledger/wasp/packages/isc"
)

func TestTypedMempoolPoolLimit(t *testing.T) {
	waitReq := NewWaitReq(waitRequestCleanupEvery)
	poolSizeLimit := 3
	size := 0
	pool := NewTypedPool[isc.OnLedgerRequest](poolSizeLimit, waitReq, func(newSize int) { size = newSize }, func(time.Duration) {}, testlogger.NewSilentLogger("", true))

	anchorAddr := cryptolib.NewRandomAddress()
	r0, err := isc.OnLedgerFromMoveRequest(isctest.RandomRequestWithRef(), anchorAddr)
	require.NoError(t, err)
	r1, err := isc.OnLedgerFromMoveRequest(isctest.RandomRequestWithRef(), anchorAddr)
	require.NoError(t, err)
	r2, err := isc.OnLedgerFromMoveRequest(isctest.RandomRequestWithRef(), anchorAddr)
	require.NoError(t, err)
	r3, err := isc.OnLedgerFromMoveRequest(isctest.RandomRequestWithRef(), anchorAddr)
	require.NoError(t, err)

	pool.Add(r0)
	pool.Add(r1)
	pool.Add(r2)
	pool.Add(r3)

	require.Equal(t, 3, size)
}
