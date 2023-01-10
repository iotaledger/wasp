// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util"
)

const (
	msgTypeNextLogIndex byte = iota
)

func (cl *cmtLogImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return UnmarshalMessage(data)
}

func UnmarshalMessage(data []byte) (gpa.Message, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("cmtLogImpl::UnmarshalMessage: data too short")
	}
	if data[0] == msgTypeNextLogIndex {
		m := &msgNextLogIndex{}
		if err := m.UnmarshalBinary(data); err != nil {
			return nil, fmt.Errorf("cannot unmarshal msgNextLogIndex: %w", err)
		}
		return m, nil
	}
	return nil, fmt.Errorf("cmtLogImpl::UnmarshalMessage: cannot parse message starting with: %v", util.PrefixHex(data, 20))
}
