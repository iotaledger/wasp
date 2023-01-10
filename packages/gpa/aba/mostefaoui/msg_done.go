// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"bytes"
	"encoding"
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util"
)

type msgDone struct {
	sender    gpa.NodeID
	recipient gpa.NodeID
	round     int
}

var (
	_ gpa.Message                = &msgDone{}
	_ encoding.BinaryUnmarshaler = &msgDone{}
)

func multicastMsgDone(nodeIDs []gpa.NodeID, me gpa.NodeID, round int) gpa.OutMessages {
	msgs := gpa.NoMessages()
	for _, n := range nodeIDs {
		if n != me {
			msgs.Add(&msgDone{recipient: n, round: round})
		}
	}
	return msgs
}

func (m *msgDone) Recipient() gpa.NodeID {
	return m.recipient
}

func (m *msgDone) SetSender(sender gpa.NodeID) {
	m.sender = sender
}

func (m *msgDone) MarshalBinary() ([]byte, error) {
	w := bytes.NewBuffer([]byte{})
	if err := util.WriteByte(w, msgTypeDone); err != nil {
		return nil, err
	}
	if err := util.WriteUint16(w, uint16(m.round)); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (m *msgDone) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	msgType, err := util.ReadByte(r)
	if err != nil {
		return err
	}
	if msgType != msgTypeDone {
		return fmt.Errorf("expected msgTypeDone, got %v", msgType)
	}
	var round uint16
	if err := util.ReadUint16(r, &round); err != nil {
		return err
	}
	m.round = int(round)
	return nil
}
