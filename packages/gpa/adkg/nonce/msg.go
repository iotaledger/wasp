// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nonce

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
)

func (n *nonceDKGImpl) subsystemFunc(subsystem byte, index int) (gpa.GPA, error) {
	if subsystem == msgWrapperACSS {
		if index < 0 || index >= len(n.acss) {
			return nil, fmt.Errorf("unexpected acss index: %v", index)
		}
		return n.acss[index], nil
	}
	return nil, fmt.Errorf("unexpected subsystem: %v", subsystem)
}

func (n *nonceDKGImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	// TODO: This should be improved pronbably considering the new serialization framework.
	if len(data) == 0 {
		return nil, fmt.Errorf("empty msg")
	}
	if data[0] != msgTypeWrapped {
		return nil, fmt.Errorf("unexpected msg type: %v", data[0])
	}
	return n.wrapper.UnmarshalMessage(data[1:])
}
