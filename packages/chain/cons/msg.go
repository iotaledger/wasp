// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/gpa"
)

const (
	msgTypeBLSShare byte = iota
	msgTypeWrapped
)

func (c *consImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	if len(data) < 1 {
		return nil, xerrors.Errorf("consImpl::UnmarshalMessage: data to short")
	}
	switch data[0] {
	case msgTypeBLSShare:
		m := &msgBLSPartialSig{blsSuite: c.blsSuite}
		if err := m.UnmarshalBinary(data); err != nil {
			return nil, xerrors.Errorf("cannot unmarshal msgBLSPartialSig: %w", err)
		}
		return m, nil
	case msgTypeWrapped:
		m, err := c.msgWrapper.UnmarshalMessage(data)
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
	return nil, xerrors.Errorf("consImpl::UnmarshalMessage: cannot parse message starting with: %v", logData)
}
