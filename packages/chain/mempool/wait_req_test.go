// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

func TestWaitReq(t *testing.T) {
	kp := cryptolib.NewKeyPair()

	ctxA := context.Background()
	ctxM := context.Background()
	req0 := isc.NewOffLedgerRequest(isc.RandomChainID(), isc.Hn("foo"), isc.Hn("bar"), nil, 0).Sign(kp)
	req1 := isc.NewOffLedgerRequest(isc.RandomChainID(), isc.Hn("foo"), isc.Hn("bar"), nil, 1).Sign(kp)
	req2 := isc.NewOffLedgerRequest(isc.RandomChainID(), isc.Hn("foo"), isc.Hn("bar"), nil, 2).Sign(kp)
	ref0 := isc.RequestRefFromRequest(req0)
	ref1 := isc.RequestRefFromRequest(req1)
	ref2 := isc.RequestRefFromRequest(req2)

	wr := mempool.NewWaitReq(3)
	var recvAny isc.Request
	recvMany := []isc.Request{}
	wr.WaitAny(ctxA, func(req isc.Request) {
		require.Nil(t, recvAny)
		recvAny = req
	})
	wr.WaitMany(ctxM, []*isc.RequestRef{ref0, ref1, ref2}, func(req isc.Request) {
		recvMany = append(recvMany, req)
	})
	wr.Have(req0)
	wr.Have(req1)
	wr.Have(req2)
	require.NotNil(t, recvAny)
	require.Len(t, recvMany, 3)
	require.Contains(t, recvMany, req0)
	require.Contains(t, recvMany, req1)
	require.Contains(t, recvMany, req2)
}
