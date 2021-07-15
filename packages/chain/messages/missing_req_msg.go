// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

// region MissingRequestIDsMsg ///////////////////////////////////////////////////

type MissingRequestIDsMsg struct {
	IDs []iscp.RequestID
}

func NewMissingRequestIDsMsg(missingIDs *[]iscp.RequestID) *MissingRequestIDsMsg {
	return &MissingRequestIDsMsg{
		IDs: *missingIDs,
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
	var buf bytes.Buffer
	_ = msg.write(&buf)
	return buf.Bytes()
}

// TODO check - is this okay? will the msg be received properly with an arbitrary amount of IDS? do we need to specify the length or is that done automatically?
func (msg *MissingRequestIDsMsg) read(r io.Reader) error {
	for {
		var buf [ledgerstate.OutputIDLength]byte
		_, err := r.Read(buf[:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return xerrors.Errorf("failed to read requestIDs: %w", err)
		}
		msg.IDs = append(msg.IDs, buf)
	}
}

func MissingRequestIDsMsgFromBytes(buf []byte) (MissingRequestIDsMsg, error) {
	r := bytes.NewReader(buf)
	msg := MissingRequestIDsMsg{}
	err := msg.read(r)
	return msg, err
}

// endregion ///////////////////////////////////////////////////////////////////

// region MissingRequestMsg ///////////////////////////////////////////////////

type MissingRequestMsg struct {
	IsOffledger bool
	Request     iscp.Request
}

func NewMissingRequestMsg(req iscp.Request) *MissingRequestMsg {
	return &MissingRequestMsg{
		IsOffledger: req.Output() == nil,
		Request:     req,
	}
}

func (msg *MissingRequestMsg) write(w io.Writer) error {
	if err := util.WriteBoolByte(w, msg.IsOffledger); err != nil {
		return xerrors.Errorf("failed to write isOffledger: %w", err)
	}
	if _, err := w.Write(msg.Request.Bytes()); err != nil {
		return xerrors.Errorf("failed to write request: %w", err)
	}
	return nil
}

func (msg *MissingRequestMsg) Bytes() []byte {
	var buf bytes.Buffer
	_ = msg.write(&buf)
	return buf.Bytes()
}

func (msg *MissingRequestMsg) read(r io.Reader) error {
	if err := util.ReadBoolByte(r, &msg.IsOffledger); err != nil {
		return xerrors.Errorf("failed to read isOffledger: %w", err)
	}
	reqBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return xerrors.Errorf("failed to read request: %w", err)
	}
	if msg.Request, err = request.FromBytes(reqBytes); err != nil {
		return xerrors.Errorf("failed to read request: %w", err)
	}
	return nil
}

func MissingRequestMsgFromBytes(buf []byte) (MissingRequestMsg, error) {
	r := bytes.NewReader(buf)
	msg := MissingRequestMsg{}
	err := msg.read(r)
	return msg, err
}

// endregion ///////////////////////////////////////////////////////////////////
