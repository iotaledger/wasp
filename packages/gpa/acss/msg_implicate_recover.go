// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type msgImplicateKind byte

const (
	msgImplicateRecoverKindIMPLICATE msgImplicateKind = iota
	msgImplicateRecoverKindRECOVER
)

// The <IMPLICATE, i, skᵢ> and <RECOVER, i, skᵢ> messages.
type msgImplicateRecover struct {
	sender    gpa.NodeID
	recipient gpa.NodeID
	kind      msgImplicateKind
	i         int
	data      []byte // Either implication or the recovered secret.
}

var _ gpa.Message = &msgImplicateRecover{}

func (m *msgImplicateRecover) Recipient() gpa.NodeID {
	return m.recipient
}

func (m *msgImplicateRecover) SetSender(sender gpa.NodeID) {
	m.sender = sender
}

func (m *msgImplicateRecover) MarshalBinary() ([]byte, error) {
	w := &bytes.Buffer{}
	if err := rwutil.WriteByte(w, msgTypeImplicateRecover); err != nil {
		return nil, err
	}
	if err := rwutil.WriteByte(w, byte(m.kind)); err != nil {
		return nil, err
	}
	if err := rwutil.WriteUint16(w, uint16(m.i)); err != nil {
		return nil, err
	}
	if err := rwutil.WriteBytes(w, m.data); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (m *msgImplicateRecover) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	t, err := rwutil.ReadByte(r)
	if err != nil {
		return err
	}
	if t != msgTypeImplicateRecover {
		return fmt.Errorf("unexpected msgType: %v in acss.msgImplicateRecover", t)
	}
	k, err := rwutil.ReadByte(r)
	if err != nil {
		return err
	}
	var i uint16
	if i, err = rwutil.ReadUint16(r); err != nil { // TODO: Resolve I from the context, trusting it might be unsafe.
		return err
	}
	d, err := rwutil.ReadBytes(r)
	if err != nil {
		return err
	}
	m.kind = msgImplicateKind(k)
	m.i = int(i)
	m.data = d
	return nil
}
