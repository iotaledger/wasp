// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func TestTypedMempoolPoolLimit(t *testing.T) {
	waitReq := NewWaitReq(waitRequestCleanupEvery)
	poolSizeLimit := 3
	size := 0
	pool := NewTypedPool[isc.OnLedgerRequest](poolSizeLimit, waitReq, func(newSize int) { size = newSize }, func(time.Duration) {}, testlogger.NewSilentLogger("", true))

	r0, err := isc.OnLedgerFromRequest(randomRequest(), cryptolib.NewRandomAddress())
	require.NoError(t, err)
	r1, err := isc.OnLedgerFromRequest(randomRequest(), cryptolib.NewRandomAddress())
	require.NoError(t, err)
	r2, err := isc.OnLedgerFromRequest(randomRequest(), cryptolib.NewRandomAddress())
	require.NoError(t, err)
	r3, err := isc.OnLedgerFromRequest(randomRequest(), cryptolib.NewRandomAddress())
	require.NoError(t, err)

	pool.Add(r0)
	pool.Add(r1)
	pool.Add(r2)
	pool.Add(r3)

	require.Equal(t, 3, size)
}

func randomRequest() *iscmove.RefWithObject[iscmove.Request] {
	ref := sui.RandomObjectRef()
	assetsBagID := sui.RandomAddress()
	return &iscmove.RefWithObject[iscmove.Request]{
		ObjectRef: *ref,
		Object: &iscmove.Request{
			ID:     *ref.ObjectID,
			Sender: cryptolib.NewRandomAddress(),
			AssetsBag: iscmove.Referent[iscmove.AssetsBagWithBalances]{
				ID: *assetsBagID,
				Value: &iscmove.AssetsBagWithBalances{
					AssetsBag: iscmove.AssetsBag{
						ID:   *assetsBagID,
						Size: 5,
					},
					Balances: iscmove.AssetsBagBalances{},
				},
			},
		},
	}
}
