// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"bytes"
	"io"

	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/util"
)

// Consensus -> Consensus
//
// Used to resync between consensus nodes on the last log index.
// This is needed to recover from misc failures, downtimes, etc.
type PeerLogIndexMsg struct {
	LogIndex journal.LogIndex
}

type PeerLogIndexMsgIn struct {
	PeerLogIndexMsg
	SenderIndex uint16
}

func NewPeerLogIndexMsg(data []byte) (*PeerLogIndexMsg, error) {
	r := bytes.NewReader(data)
	var li uint32
	if err := util.ReadUint32(r, &li); err != nil {
		return nil, xerrors.Errorf("cannot deserialize PeerLogIndexMsg.LogIndex: %w", err)
	}
	return &PeerLogIndexMsg{LogIndex: journal.LogIndex(li)}, nil
}

func (msg *PeerLogIndexMsg) Write(w io.Writer) error {
	if err := util.WriteUint32(w, msg.LogIndex.AsUint32()); err != nil {
		return xerrors.Errorf("cannot serialize PeerLogIndexMsg.LogIndex: %w", err)
	}
	return nil
}

func (msg *PeerLogIndexMsg) Bytes() ([]byte, error) {
	w := bytes.NewBuffer([]byte{})
	if err := msg.Write(w); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}
