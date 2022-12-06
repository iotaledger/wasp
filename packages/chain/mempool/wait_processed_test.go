// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

func TestWaitProcessed(t *testing.T) {
	client := cryptolib.NewKeyPair()
	req := isc.NewOffLedgerRequest(isc.RandomChainID(), isc.EntryPointInit, isc.EntryPointInit, dict.New(), 0).Sign(client)
	reqID := req.ID()
	reqCh := make(chan *blocklog.RequestReceipt, 1)
	reqRcp := &blocklog.RequestReceipt{Request: req}
	ctx := context.Background()

	wp := NewWaitProcessed(0)
	wp.Await(&reqAwaitRequestProcessed{ctx: ctx, requestID: reqID, responseCh: reqCh})
	select {
	case <-reqCh:
		t.Error("responded to fast")
	case <-time.After(100 * time.Millisecond):
		wp.Processed(reqRcp)
		require.Equal(t, reqRcp, <-reqCh)
	}
}
