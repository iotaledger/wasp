// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"github.com/iotaledger/hive.go/core/marshalutil"
	"github.com/iotaledger/wasp/packages/isc"
)

type MissingRequestMsg struct {
	Request isc.Request
}

func (msg *MissingRequestMsg) Bytes() []byte {
	return msg.Request.Bytes()
}

func NewMissingRequestMsg(data []byte) (*MissingRequestMsg, error) {
	msg := &MissingRequestMsg{}
	var err error
	msg.Request, err = isc.NewRequestFromMarshalUtil(marshalutil.New(data))
	if err != nil {
		return nil, err
	}
	return msg, nil
}
