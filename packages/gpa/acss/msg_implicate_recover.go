// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"bytes"

	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util"
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
	if err := util.WriteByte(w, msgTypeImplicateRecover); err != nil {
		return nil, err
	}
	if err := util.WriteByte(w, byte(m.kind)); err != nil {
		return nil, err
	}
	if err := util.WriteUint16(w, uint16(m.i)); err != nil {
		return nil, err
	}
	if err := util.WriteBytes32(w, m.data); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (m *msgImplicateRecover) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	t, err := util.ReadByte(r)
	if err != nil {
		return err
	}
	if t != msgTypeImplicateRecover {
		return xerrors.Errorf("unexpected msgType: %v", t)
	}
	k, err := util.ReadByte(r)
	if err != nil {
		return err
	}
	var i uint16
	if err := util.ReadUint16(r, &i); err != nil { // TODO: Resolve I from the context, trusting it might be unsafe.
		return err
	}
	d, err := util.ReadBytes32(r)
	if err != nil {
		return err
	}
	m.kind = msgImplicateKind(k)
	m.i = int(i)
	m.data = d
	return nil
}
