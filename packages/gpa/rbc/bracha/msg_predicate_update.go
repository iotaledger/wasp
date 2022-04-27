// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bracha

import "github.com/iotaledger/wasp/packages/gpa"

type msgPredicateUpdate struct {
	me        gpa.NodeID
	predicate func([]byte) bool
}

var _ gpa.Message = &msgPredicateUpdate{}

func (m *msgPredicateUpdate) Recipient() gpa.NodeID {
	return m.me
}

func (m *msgPredicateUpdate) SetSender(sender gpa.NodeID) {
	// Don't care the sender.
}

func (m *msgPredicateUpdate) MarshalBinary() ([]byte, error) {
	panic("not used")
}
