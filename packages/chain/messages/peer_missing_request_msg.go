// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/iscp"
)

type MissingRequestMsg struct {
	Request iscp.Request
}

func (msg *MissingRequestMsg) Bytes() []byte {
	return msg.Request.Bytes()
}

func NewMissingRequestMsg(data []byte) (*MissingRequestMsg, error) {
	msg := &MissingRequestMsg{}
	var err error
	msg.Request, err = iscp.RequestDataFromMarshalUtil(marshalutil.New(data))
	if err != nil {
		return nil, err
	}
	return msg, nil
}
