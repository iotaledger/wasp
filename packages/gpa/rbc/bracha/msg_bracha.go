// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bracha

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util"
)

type msgBrachaType byte

const (
	msgBrachaTypeInitial msgBrachaType = iota
	msgBrachaTypeEcho
	msgBrachaTypeReady
)

type msgBracha struct {
	t msgBrachaType // Type
	s gpa.NodeID    // Transient: Sender
	r gpa.NodeID    // Transient: Recipient
	v []byte        // Value
}

var _ gpa.Message = &msgBracha{}

func (m *msgBracha) Recipient() gpa.NodeID {
	return m.r
}

func (m *msgBracha) SetSender(sender gpa.NodeID) {
	m.s = sender
}

func (m *msgBracha) MarshalBinary() ([]byte, error) {
	w := &bytes.Buffer{}
	if err := util.WriteByte(w, byte(m.t)); err != nil {
		return nil, err
	}
	if err := util.WriteBytes32(w, m.v); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (m *msgBracha) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	t, err := util.ReadByte(r)
	if err != nil {
		return err
	}
	v, err := util.ReadBytes32(r)
	if err != nil {
		return err
	}
	m.t = msgBrachaType(t)
	m.v = v
	return nil
}
