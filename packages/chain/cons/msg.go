// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util"
)

const (
	msgTypeBLSShare byte = iota
	msgTypeWrapped
)

func (c *consImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	if len(data) < 1 {
		return nil, errors.New("consImpl::UnmarshalMessage: data too short")
	}
	switch data[0] {
	case msgTypeBLSShare:
		m := &msgBLSPartialSig{blsSuite: c.blsSuite}
		if err := m.UnmarshalBinary(data); err != nil {
			return nil, fmt.Errorf("cannot unmarshal msgBLSPartialSig: %w", err)
		}
		return m, nil
	case msgTypeWrapped:
		m, err := c.msgWrapper.UnmarshalMessage(data)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal Wrapped msg: %w", err)
		}
		return m, nil
	}
	return nil, fmt.Errorf("consImpl::UnmarshalMessage: cannot parse message starting with: %v", util.PrefixHex(data, 20))
}
