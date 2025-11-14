// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"go.dedis.ch/kyber/v3/suites"
)

// This message is used as a payload of the RBC:
//
// > RBC(C||E)
type MsgRBCCEPayload struct {
	suite suites.Suite
	data  []byte `bcs:"export"`
	err   error  // Transient field, should not be serialized.
}
