// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/iotaledger/wasp/packages/coretypes"
	"golang.org/x/xerrors"
)

// region MissingRequestsMsg ///////////////////////////////////////////////////

type RequestAcKMsg struct {
	ReqID *coretypes.RequestID
}

func NewRequestAckMsg(reqID coretypes.RequestID) *RequestAcKMsg {
	return &RequestAcKMsg{
		ReqID: &reqID,
	}
}

func (msg *RequestAcKMsg) write(w io.Writer) error {
	if _, err := w.Write(msg.ReqID.Bytes()); err != nil {
		return xerrors.Errorf("failed to write requestIDs: %w", err)
	}
	return nil
}

func (msg *RequestAcKMsg) Bytes() []byte {
	var buf bytes.Buffer
	_ = msg.write(&buf)
	return buf.Bytes()
}

func (msg *RequestAcKMsg) read(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return xerrors.Errorf("failed to read requestIDs: %w", err)
	}
	reqID, err := coretypes.RequestIDFromBytes(b)
	if err != nil {
		return xerrors.Errorf("failed to read requestIDs: %w", err)
	}
	msg.ReqID = &reqID
	return nil
}

func RequestAckMsgFromBytes(buf []byte) (RequestAcKMsg, error) {
	r := bytes.NewReader(buf)
	msg := RequestAcKMsg{}
	err := msg.read(r)
	return msg, err
}
