// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package amDist

import (
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util"
)

const (
	msgTypeAccess byte = iota
)

func (ami *accessMgrDist) UnmarshalMessage(data []byte) (gpa.Message, error) {
	if len(data) < 1 {
		return nil, errors.New("accessMgrImpl::UnmarshalMessage: data too short")
	}
	if data[0] == msgTypeAccess {
		m := &msgAccess{}
		if err := m.UnmarshalBinary(data); err != nil {
			return nil, fmt.Errorf("cannot unmarshal msgAccess: %w", err)
		}
		return m, nil
	}
	return nil, fmt.Errorf("accessMgrImpl::UnmarshalMessage: cannot parse message starting with: %v", util.PrefixHex(data, 20))
}
