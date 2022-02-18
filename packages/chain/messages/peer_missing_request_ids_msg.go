// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
)

type MissingRequestIDsMsg struct {
	IDs []iscp.RequestID
}

type MissingRequestIDsMsgIn struct {
	MissingRequestIDsMsg
	SenderPubKey *cryptolib.PublicKey
}

func (msg *MissingRequestIDsMsg) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteUint16(uint16(len(msg.IDs)))
	for i := range msg.IDs {
		mu.WriteBytes(msg.IDs[i].Bytes())
	}
	return mu.Bytes()
}

func NewMissingRequestIDsMsg(data []byte) (*MissingRequestIDsMsg, error) {
	mu := marshalutil.New(data)
	num, err := mu.ReadUint16()
	if err != nil {
		return nil, err
	}
	ret := &MissingRequestIDsMsg{
		IDs: make([]iscp.RequestID, num),
	}
	for i := range ret.IDs {
		if ret.IDs[i], err = iscp.RequestIDFromMarshalUtil(mu); err != nil {
			return nil, err
		}
	}
	return ret, nil
}
