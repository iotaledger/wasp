// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/gpa"
)

const (
	msgTypeVote byte = iota
	msgTypeDone
	msgTypeWrapped
)

// Implements the gpa.GPA interface.
func (a *abaImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	if len(data) < 1 {
		return nil, xerrors.Errorf("data to short")
	}
	msgType := data[0]
	switch msgType {
	case msgTypeVote:
		m := &msgVote{}
		if err := m.UnmarshalBinary(data); err != nil {
			return nil, err
		}
		return m, nil
	case msgTypeDone:
		m := &msgDone{}
		if err := m.UnmarshalBinary(data); err != nil {
			return nil, err
		}
		return m, nil
	case msgTypeWrapped:
		m, err := a.msgWrapper.UnmarshalMessage(data)
		if err != nil {
			return nil, xerrors.Errorf("cannot unmarshal Wrapped msg: %w", err)
		}
		return m, nil
	}
	var logData []byte
	if len(data) <= 20 {
		logData = data
	} else {
		logData = data[0:20]
	}
	return nil, xerrors.Errorf("abaImpl::UnmarshalMessage: unexpected msgType: %v, message starts with: %w", msgType, logData)
}
