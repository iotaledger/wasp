// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bracha

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"golang.org/x/xerrors"
)

type msgBrachaType byte

const (
	msgBrachaTypeInitial msgBrachaType = iota
	msgBrachaTypeEcho
	msgBrachaTypeReady
)

type msgBracha struct {
	t msgBrachaType // Type
	s gpa.NodeID    // Sender
	r gpa.NodeID    // Recipient
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
	// return serializer.NewSerializer(). // TODO: Implement.
	// 	WriteByte(byte(m.t), func(err error) error { return xerrors.Errorf("unable to serialize t: %w", err) }).
	// 	WriteString(string(m.s), serializer.SeriLengthPrefixTypeAsUint16, func(err error) error { return xerrors.Errorf("unable to serialize s: %w", err) }).
	// 	WriteString(string(m.r), serializer.SeriLengthPrefixTypeAsUint16, func(err error) error { return xerrors.Errorf("unable to serialize r: %w", err) }).
	// 	WriteVariableByteSlice(m.v, serializer.SeriLengthPrefixTypeAsUint16, func(err error) error { return xerrors.Errorf("unable to serialize v: %w", err) }).
	// 	Serialize()
	panic(xerrors.Errorf("not implemented"))
}
