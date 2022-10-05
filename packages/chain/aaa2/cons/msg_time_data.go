// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"time"

	"github.com/iotaledger/wasp/packages/gpa"
)

type msgTimeData struct {
	node     gpa.NodeID
	timeData time.Time
}

var _ gpa.Message = &msgTimeData{}

func NewMsgTimeData(node gpa.NodeID, timeData time.Time) gpa.Message {
	return &msgTimeData{node: node, timeData: timeData}
}

func (m *msgTimeData) Recipient() gpa.NodeID {
	return m.node
}

func (m *msgTimeData) SetSender(sender gpa.NodeID) {
	if sender != m.node {
		panic("strange node for local message")
	}
}

func (m *msgTimeData) MarshalBinary() ([]byte, error) {
	panic("no marshaling for local message")
}
