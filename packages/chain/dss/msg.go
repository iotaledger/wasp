// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dss

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util"
)

const (
	msgTypePartialSig byte = iota
	msgTypeWrapped
)

func (d *dssImpl) msgWrapperFunc(subsystem byte, index int) (gpa.GPA, error) {
	if subsystem == subsystemDKG {
		if index != 0 {
			return nil, fmt.Errorf("unexpected DKG index: %v", index)
		}
		return d.dkg, nil
	}
	return nil, fmt.Errorf("unexpected subsystem: %v", subsystem)
}

func (d *dssImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("dssImpl::UnmarshalMessage: data too short")
	}
	switch data[0] {
	case msgTypePartialSig:
		m := &msgPartialSig{suite: d.suite}
		if err := m.UnmarshalBinary(data); err != nil {
			return nil, fmt.Errorf("cannot unmarshal msgPartialSig: %w", err)
		}
		return m, nil
	case msgTypeWrapped:
		m, err := d.msgWrapper.UnmarshalMessage(data)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal Wrapped msg: %w", err)
		}
		return m, nil
	}
	return nil, fmt.Errorf("dssImpl::UnmarshalMessage: cannot parse message starting with: %v", util.PrefixHex(data, 20))
}
