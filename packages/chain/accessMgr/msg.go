// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package accessMgr

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
)

const (
	msgTypeAccess byte = iota
)

func (ami *accessMgrImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("accessMgrImpl::UnmarshalMessage: data to short")
	}
	if data[0] == msgTypeAccess {
		m := &msgAccess{}
		if err := m.UnmarshalBinary(data); err != nil {
			return nil, fmt.Errorf("cannot unmarshal msgAccess: %w", err)
		}
		return m, nil
	}
	var logData []byte
	if len(data) <= 20 {
		logData = data
	} else {
		logData = data[0:20]
	}
	return nil, fmt.Errorf("accessMgrImpl::UnmarshalMessage: cannot parse message starting with: %v", logData)
}
