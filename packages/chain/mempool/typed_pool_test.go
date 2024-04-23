// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestTypedMempoolPoolLimit(t *testing.T) {
	waitReq := NewWaitReq(waitRequestCleanupEvery)
	poolSizeLimit := 3
	size := 0
	pool := NewTypedPool[isc.OnLedgerRequest](poolSizeLimit, waitReq, func(newSize int) { size = newSize }, func(time.Duration) {}, testlogger.NewSilentLogger("", true))

	r0, err := isc.OnLedgerFromUTXO(&iotago.BasicOutput{}, tpkg.RandOutputID(0))
	require.NoError(t, err)
	r1, err := isc.OnLedgerFromUTXO(&iotago.BasicOutput{}, tpkg.RandOutputID(1))
	require.NoError(t, err)
	r2, err := isc.OnLedgerFromUTXO(&iotago.BasicOutput{}, tpkg.RandOutputID(2))
	require.NoError(t, err)
	r3, err := isc.OnLedgerFromUTXO(&iotago.BasicOutput{}, tpkg.RandOutputID(3))
	require.NoError(t, err)

	pool.Add(r0)
	pool.Add(r1)
	pool.Add(r2)
	pool.Add(r3)

	require.Equal(t, 3, size)
}
