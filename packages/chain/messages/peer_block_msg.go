// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"bytes"
	"io"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util"
)

// BlockMsg StateManager in response to GetBlockMsg sends block data to the querying node's StateManager
// StateManager -> StateManager
type BlockMsg struct {
	BlockBytes []byte
}

type BlockMsgIn struct {
	BlockMsg
	SenderPubKey *cryptolib.PublicKey
}

func NewBlockMsg(data []byte) (*BlockMsg, error) {
	msg := &BlockMsg{}
	r := bytes.NewReader(data)
	var err error
	if msg.BlockBytes, err = util.ReadBytes32(r); err != nil {
		return nil, err
	}
	return msg, nil
}

func (msg *BlockMsg) Write(w io.Writer) error {
	return util.WriteBytes32(w, msg.BlockBytes)
}
