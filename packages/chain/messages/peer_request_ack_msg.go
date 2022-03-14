// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"bytes"
	"io"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

type RequestAckMsg struct {
	ReqID *iscp.RequestID
}

type RequestAckMsgIn struct {
	RequestAckMsg
	SenderPubKey *ed25519.PublicKey
}

func NewRequestAckMsg(buf []byte) (*RequestAckMsg, error) {
	r := bytes.NewReader(buf)
	msg := &RequestAckMsg{}
	err := msg.read(r)
	return msg, err
}

func (msg *RequestAckMsg) write(w io.Writer) error {
	if _, err := w.Write(msg.ReqID.Bytes()); err != nil {
		return xerrors.Errorf("failed to write requestIDs: %w", err)
	}
	return nil
}

func (msg *RequestAckMsg) Bytes() []byte {
	var buf bytes.Buffer
	_ = msg.write(&buf)
	return buf.Bytes()
}

func (msg *RequestAckMsg) read(r io.Reader) error {
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
