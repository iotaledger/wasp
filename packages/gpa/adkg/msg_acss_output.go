// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package adkg

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"go.dedis.ch/kyber/v3/share"
)

//
// An event to self.
type msgACSSOutput struct {
	me       gpa.NodeID
	index    int
	priShare *share.PriShare
}

func (m *msgACSSOutput) Recipient() gpa.NodeID {
	return m.me
}

func (m *msgACSSOutput) MarshalBinary() ([]byte, error) {
	return nil, nil // TODO: Implement.
}
