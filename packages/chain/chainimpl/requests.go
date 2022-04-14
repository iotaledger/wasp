// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Provides implementations for chain.ChainRequests methods
package chainimpl

import (
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

func (c *chainObj) GetRequestReceipt(reqID iscp.RequestID) (*blocklog.RequestReceipt, error) {
	res, err := blocklog.GetRequestRecordDataByRequestID(
		c.stateReader.KVStoreReader(),
		reqID,
	)
	if err != nil {
		return nil, err
	}
	receipt, err := blocklog.RequestReceiptFromBytes(res.ReceiptBin)
	if err != nil {
		c.log.Errorf("error parsing receipt from bin: %s", err)
		return nil, err
	}
	receipt.BlockIndex = res.BlockIndex
	receipt.RequestIndex = res.RequestIndex
	return receipt, nil
}

func (c *chainObj) AttachToRequestProcessed(handler func(iscp.RequestID)) *events.Closure {
	closure := events.NewClosure(handler)
	c.eventRequestProcessed.Attach(closure)
	return closure
}

func (c *chainObj) DetachFromRequestProcessed(attachID *events.Closure) {
	c.eventRequestProcessed.Detach(attachID)
}
