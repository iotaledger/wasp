// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util"
)

const (
	msgTypeVote byte = iota
	msgTypeDone
	msgTypeWrapped
)

// Implements the gpa.GPA interface.
func (a *abaImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	if len(data) < 1 {
		return nil, errors.New("data to short")
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
			return nil, fmt.Errorf("cannot unmarshal Wrapped msg: %w", err)
		}
		return m, nil
	}
	return nil, fmt.Errorf("abaImpl::UnmarshalMessage: unexpected msgType: %v, message starts with: %s", msgType, util.PrefixHex(data, 20))
}
