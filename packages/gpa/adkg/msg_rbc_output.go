// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package adkg

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

//
// An event to self.
type msgRBCOutput struct { // TODO: Not used.
	me      gpa.NodeID
	indexes []int
}

func (m *msgRBCOutput) Recipient() gpa.NodeID {
	return m.me
}

func (m *msgRBCOutput) MarshalBinary() ([]byte, error) {
	return nil, nil // TODO: Implement.
}
