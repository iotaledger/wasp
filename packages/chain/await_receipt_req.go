// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

type awaitReceiptReq struct {
	ctx        context.Context
	requestID  isc.RequestID
	responseCh chan<- *blocklog.RequestReceipt
}

func newAwaitReceiptReq(ctx context.Context, requestID isc.RequestID) (*awaitReceiptReq, <-chan *blocklog.RequestReceipt) {
	responseCh := make(chan *blocklog.RequestReceipt, 1)
	req := &awaitReceiptReq{
		ctx:        ctx,
		requestID:  requestID,
		responseCh: responseCh,
	}
	return req, responseCh
}

func (r *awaitReceiptReq) Respond(receipt *blocklog.RequestReceipt) {
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
