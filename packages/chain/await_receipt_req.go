// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"
	"time"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/blocklog"
)

type awaitReceiptReq struct {
	ctx        context.Context
	requestID  isc.RequestID
	responseCh chan<- *blocklog.RequestReceipt
	startTime  time.Time
	log        log.Logger
}

func newAwaitReceiptReq(ctx context.Context, requestID isc.RequestID, log log.Logger) (*awaitReceiptReq, <-chan *blocklog.RequestReceipt) {
	log.LogDebugf("AwaitRequestProcessed(%v) started", requestID)
	responseCh := make(chan *blocklog.RequestReceipt, 1)
	r := &awaitReceiptReq{
		ctx:        ctx,
		requestID:  requestID,
		responseCh: responseCh,
		startTime:  time.Now(),
		log:        log,
	}
	return r, responseCh
}

func (r *awaitReceiptReq) Respond(receipt *blocklog.RequestReceipt) {
	r.log.LogDebugf("AwaitRequestProcessed(%v) responding in %v", r.requestID, time.Since(r.startTime))
	if r.ctx.Err() != nil {
		close(r.responseCh)
		return
	}
	r.responseCh <- receipt
	close(r.responseCh)
}

func (r *awaitReceiptReq) CloseIfCanceled() bool {
	if r.ctx.Err() != nil {
		close(r.responseCh)
		return true
	}
	return false
}
