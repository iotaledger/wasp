// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/gpa"
)

const (
	msgTypeNextLogIndex byte = iota
)

func (cl *cmtLogImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	if len(data) < 1 {
		return nil, xerrors.Errorf("cmtLogImpl::UnmarshalMessage: data to short")
	}
	if data[0] == msgTypeNextLogIndex {
		m := &msgNextLogIndex{}
		if err := m.UnmarshalBinary(data); err != nil {
			return nil, xerrors.Errorf("cannot unmarshal msgNextLogIndex: %w", err)
		}
		return m, nil
	}
	var logData []byte
	if len(data) <= 20 {
		logData = data
	} else {
		logData = data[0:20]
	}
	return nil, xerrors.Errorf("cmtLogImpl::UnmarshalMessage: cannot parse message starting with: %v", logData)
}
