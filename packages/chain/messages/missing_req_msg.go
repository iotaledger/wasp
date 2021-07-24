// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"io"

	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"golang.org/x/xerrors"
)

// region MissingRequestIDsMsg ///////////////////////////////////////////////////

type MissingRequestIDsMsg struct {
	IDs []iscp.RequestID
}

func NewMissingRequestIDsMsg(missingIDs []iscp.RequestID) *MissingRequestIDsMsg {
	return &MissingRequestIDsMsg{
		IDs: missingIDs,
	}
}

func (msg *MissingRequestIDsMsg) write(w io.Writer) error {
	for _, ID := range msg.IDs {
		if _, err := w.Write(ID.Bytes()); err != nil {
			return xerrors.Errorf("failed to write requestIDs: %w", err)
		}
	}
	return nil
}

func (msg *MissingRequestIDsMsg) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteUint16(uint16(len(msg.IDs)))
	for i := range msg.IDs {
		mu.WriteBytes(msg.IDs[i].Bytes())
	}
	return mu.Bytes()
}

func MissingRequestIDsMsgFromBytes(data []byte) (*MissingRequestIDsMsg, error) {
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

// endregion ///////////////////////////////////////////////////////////////////

// region MissingRequestMsg ///////////////////////////////////////////////////

type MissingRequestMsg struct {
	Request iscp.Request
}

func NewMissingRequestMsg(req iscp.Request) *MissingRequestMsg {
	return &MissingRequestMsg{
		Request: req,
	}
}

func (msg *MissingRequestMsg) Bytes() []byte {
	return msg.Request.Bytes()
}

func MissingRequestMsgFromBytes(data []byte) (*MissingRequestMsg, error) {
	msg := &MissingRequestMsg{}
	var err error
	msg.Request, err = request.FromMarshalUtil(marshalutil.New(data))
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// endregion ///////////////////////////////////////////////////////////////////
