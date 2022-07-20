// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package das

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"golang.org/x/xerrors"
)

//
// An event to self.
type msgACSSOutput struct {
	me       gpa.NodeID
	index    int
	priShare *share.PriShare
	commits  []kyber.Point
}

var _ gpa.Message = &msgACSSOutput{}

func (m *msgACSSOutput) Recipient() gpa.NodeID {
	return m.me
}

func (m *msgACSSOutput) SetSender(sender gpa.NodeID) {
	// Don't care the sender.
}

func (m *msgACSSOutput) MarshalBinary() ([]byte, error) {
	return nil, xerrors.Errorf("msgACSSOutput::MarshalBinary not implemented") // TODO: Implement.
}
