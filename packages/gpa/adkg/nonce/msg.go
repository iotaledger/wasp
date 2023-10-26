// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nonce

import (
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/wasp/packages/gpa"
)

func (n *nonceDKGImpl) subsystemFunc(subsystem byte, index int) (gpa.GPA, error) {
	if subsystem == msgWrapperACSS {
		if index < 0 || index >= len(n.acss) {
			return nil, ierrors.Errorf("unexpected acss index: %v", index)
		}
		return n.acss[index], nil
	}
	return nil, ierrors.Errorf("unexpected subsystem: %v", subsystem)
}

func (n *nonceDKGImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	// All non-node-local messages are from the ACSS, so just pass it there.
	return n.wrapper.UnmarshalMessage(data)
}
