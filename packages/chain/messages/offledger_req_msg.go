// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"golang.org/x/xerrors"
)

type OffLedgerRequestPeerMsg struct {
	ChainID *iscp.ChainID
	Req     *request.OffLedger
}

func NewOffLedgerRequestPeerMsg(chainID *iscp.ChainID, req *request.OffLedger) *OffLedgerRequestPeerMsg {
	return &OffLedgerRequestPeerMsg{
		ChainID: chainID,
		Req:     req,
	}
}

func (msg *OffLedgerRequestPeerMsg) Bytes() []byte {
	return marshalutil.New().
		Write(msg.ChainID).
		Write(msg.Req).
		Bytes()
}

func OffLedgerRequestPeerMsgFromBytes(data []byte) (*OffLedgerRequestPeerMsg, error) {
	mu := marshalutil.New(data)
	chainID, err := iscp.ChainIDFromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	req, err := request.FromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	reqCasted, ok := req.(*request.OffLedger)
	if !ok {
		return nil, xerrors.New("OffLedgerRequestMsgFromBytes: wrong type of request data")
	}
	return &OffLedgerRequestPeerMsg{
		ChainID: chainID,
		Req:     reqCasted,
	}, nil
}
