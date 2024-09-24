// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/wasp/packages/gpa"
)

// This message is used as a payload of the RBC:
//
// > RBC(C||E)
type msgRBCCEPayload struct {
	gpa.BasicMessage
	suite suites.Suite
	data  []byte `bcs:""`
	err   error  // Transient field, should not be serialized.
}

var _ gpa.Message = new(msgRBCCEPayload)
