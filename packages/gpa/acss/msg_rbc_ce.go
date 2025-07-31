// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/wasp/v2/packages/gpa"
)

// This message is used as a payload of the RBC:
//
// > RBC(C||E)
type msgRBCCEPayload struct {
	gpa.BasicMessage
	suite suites.Suite
	data  []byte `bcs:"export"`
	err   error  // Transient field, should not be serialized.
}

var _ gpa.Message = new(msgRBCCEPayload)

func (m *msgRBCCEPayload) MsgType() gpa.MessageType {
	return msgTypeRBCCEPayload
}
