// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"bytes"
	"io"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
)

// Consensus -> Consensus
type SignedResultAckMsg struct {
	ChainInputID ledgerstate.OutputID
	EssenceHash  hashing.HashValue
}

type SignedResultAckMsgIn struct {
	SignedResultAckMsg
	SenderIndex uint16
}

func NewSignedResultAckMsg(data []byte) (*SignedResultAckMsg, error) {
	msg := &SignedResultAckMsg{}
	r := bytes.NewReader(data)
	var err error
	if err = util.ReadHashValue(r, &msg.EssenceHash); err != nil { // nolint:gocritic // - ignore sloppyReassign
		return nil, err
	}
	if err = util.ReadOutputID(r, &msg.ChainInputID); err != nil { // nolint:gocritic // - ignore sloppyReassign
		return nil, err
	}
	return msg, nil
}

func (msg *SignedResultAckMsg) Write(w io.Writer) error {
	if _, err := w.Write(msg.EssenceHash[:]); err != nil {
		return err
	}
	if _, err := w.Write(msg.ChainInputID[:]); err != nil {
		return err
	}
	return nil
}
