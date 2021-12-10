// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"bytes"
	"io"

	"github.com/iotaledger/wasp/packages/util"
)

// GetBlockMsg StateManager queries specific block data from another peer (access node)
// StateManager -> StateManager
type GetBlockMsg struct {
	BlockIndex uint32
}

type GetBlockMsgIn struct {
	GetBlockMsg
	SenderNetID string
}

func NewGetBlockMsg(data []byte) (*GetBlockMsg, error) {
	msg := &GetBlockMsg{}
	r := bytes.NewReader(data)
	if err := util.ReadUint32(r, &msg.BlockIndex); err != nil {
		return nil, err
	}
	return msg, nil
}

func (msg *GetBlockMsg) Write(w io.Writer) error {
	return util.WriteUint32(w, msg.BlockIndex)
}
