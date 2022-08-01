// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

type MissingRequestIDsMsg struct {
	IDs []isc.RequestID
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
		IDs: make([]isc.RequestID, num),
	}
	for i := range ret.IDs {
		if ret.IDs[i], err = isc.RequestIDFromMarshalUtil(mu); err != nil {
			return nil, err
		}
	}
	return ret, nil
}
