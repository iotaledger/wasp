// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"bytes"
	"io"

	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

type RequestAckPeerMsg struct {
	ReqID *iscp.RequestID
}

func NewRequestAckMsg(reqID iscp.RequestID) *RequestAckPeerMsg {
	return &RequestAckPeerMsg{
		ReqID: &reqID,
	}
}

func (msg *RequestAckPeerMsg) write(w io.Writer) error {
	if _, err := w.Write(msg.ReqID.Bytes()); err != nil {
		return xerrors.Errorf("failed to write requestIDs: %w", err)
	}
	return nil
}

func (msg *RequestAckPeerMsg) Bytes() []byte {
	var buf bytes.Buffer
	_ = msg.write(&buf)
	return buf.Bytes()
}

func (msg *RequestAckPeerMsg) read(r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return xerrors.Errorf("failed to read requestIDs: %w", err)
	}
	reqID, err := iscp.RequestIDFromBytes(b)
	if err != nil {
		return xerrors.Errorf("failed to read requestIDs: %w", err)
	}
	msg.ReqID = &reqID
	return nil
}

func RequestAckPeerMsgFromBytes(buf []byte) (RequestAckPeerMsg, error) {
	r := bytes.NewReader(buf)
	msg := RequestAckPeerMsg{}
	err := msg.read(r)
	return msg, err
}
