// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"bytes"
	"io"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"go.dedis.ch/kyber/v3/sign/tbls"
)

// Consensus -> Consensus
type SignedResultMsg struct {
	ChainInputID ledgerstate.OutputID
	EssenceHash  hashing.HashValue
	SigShare     tbls.SigShare
}

type SignedResultMsgIn struct {
	SignedResultMsg
	SenderIndex uint16
}

func NewSignedResultMsg(data []byte) (*SignedResultMsg, error) {
	msg := &SignedResultMsg{}
	r := bytes.NewReader(data)
	var err error
	if err = util.ReadHashValue(r, &msg.EssenceHash); err != nil { // nolint:gocritic // - ignore sloppyReassign
		return nil, err
	}
	if msg.SigShare, err = util.ReadBytes16(r); err != nil {
		return nil, err
	}
	if err = util.ReadOutputID(r, &msg.ChainInputID); err != nil { // nolint:gocritic // - ignore sloppyReassign
		return nil, err
	}
	return msg, nil
}

func (msg *SignedResultMsg) Write(w io.Writer) error {
	if _, err := w.Write(msg.EssenceHash[:]); err != nil {
		return err
	}
	if err := util.WriteBytes16(w, msg.SigShare); err != nil {
		return err
	}
	if _, err := w.Write(msg.ChainInputID[:]); err != nil {
		return err
	}
	return nil
}
