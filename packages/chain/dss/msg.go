// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dss

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"golang.org/x/xerrors"
)

const (
	msgTypePartialSig byte = iota
	msgTypeWrapped
)

func (d *dssImpl) msgWrapperFunc(subsystem byte, index int) (gpa.GPA, error) {
	if subsystem == subsystemDKG {
		if index != 0 {
			return nil, xerrors.Errorf("unexpected DKG index: %v", index)
		}
		return d.dkg, nil
	}
	return nil, xerrors.Errorf("unexpected subsystem: %v", subsystem)
}

func (d *dssImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	if len(data) < 1 {
		return nil, xerrors.Errorf("dssImpl::UnmarshalMessage: data to short")
	}
	switch data[0] {
	case msgTypePartialSig:
		m := &msgPartialSig{suite: d.suite}
		if err := m.UnmarshalBinary(data); err != nil {
			return nil, xerrors.Errorf("cannot unmarshal msgPartialSig: %w", err)
		}
		return m, nil
	case msgTypeWrapped:
		m, err := d.msgWrapper.UnmarshalMessage(data)
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
	return nil, xerrors.Errorf("dssImpl::UnmarshalMessage: cannot parse message starting with: %v", logData)
}
