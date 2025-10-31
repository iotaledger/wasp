// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nonce

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/gpa"
)

func (n *nonceDistributedKeyGenerationImpl) subsystemFunc(subsystem byte, index int) (gpa.GPA, error) {
	if subsystem == msgWrapperACSS {
		if index < 0 || index >= len(n.acss) {
			return nil, fmt.Errorf("unexpected acss index: %v", index)
		}
		return n.acss[index], nil
	}
	return nil, fmt.Errorf("unexpected subsystem: %v", subsystem)
}

func (n *nonceDistributedKeyGenerationImpl) UnmarshalPayload(data []byte) (gpa.MessagePayload, error) {
	return gpa.UnmarshalPayload(data, gpa.PayloadAllocator{}, gpa.PayloadFallback{
		msgTypeWrapped: n.wrapper.UnmarshalPayload,
	})
}
