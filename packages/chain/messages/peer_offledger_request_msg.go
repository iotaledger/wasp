// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

type OffLedgerRequestMsg struct {
	ChainID *iscp.ChainID
	Req     *iscp.OffLedgerRequestData
}

type OffLedgerRequestMsgIn struct {
	OffLedgerRequestMsg
	SenderPubKey *cryptolib.PublicKey
}

func NewOffLedgerRequestMsg(chainID *iscp.ChainID, req *iscp.OffLedgerRequestData) *OffLedgerRequestMsg {
	return &OffLedgerRequestMsg{
		ChainID: chainID,
		Req:     req,
	}
}

func (msg *OffLedgerRequestMsg) Bytes() []byte {
	return marshalutil.New().
		Write(msg.ChainID).
		Write(msg.Req).
		Bytes()
}

func OffLedgerRequestMsgFromBytes(data []byte) (*OffLedgerRequestMsg, error) {
	mu := marshalutil.New(data)
	chainID, err := iscp.ChainIDFromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	req, err := iscp.RequestDataFromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	reqCasted, ok := req.(*iscp.OffLedgerRequestData)
	if !ok {
		return nil, xerrors.New("OffLedgerRequestMsgFromBytes: wrong type of request data")
	}
	return &OffLedgerRequestMsg{
		ChainID: chainID,
		Req:     reqCasted,
	}, nil
}
