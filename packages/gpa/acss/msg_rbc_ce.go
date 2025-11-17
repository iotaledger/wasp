// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

// This message is used as a payload of the RBC:
//
// > RBC(C||E)
type MsgRBCCEPayload struct {
	data []byte `bcs:"export"`
}
