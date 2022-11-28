// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distSync

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type msgMissingRequest struct {
	gpa.BasicMessage
	requestRef *isc.RequestRef
}

var _ gpa.Message = &msgMissingRequest{}

func newMsgMissingRequest(requestRef *isc.RequestRef, recipient gpa.NodeID) gpa.Message {
	return &msgMissingRequest{
		BasicMessage: gpa.NewBasicMessage(recipient),
		requestRef:   requestRef,
	}
}

func (msg *msgMissingRequest) MarshalBinary() (data []byte, err error) {
	ret := []byte{msgTypeMissingRequest}
	ret = append(ret, msg.requestRef.Bytes()...)
	return ret, nil
}

func (msg *msgMissingRequest) UnmarshalBinary(data []byte) (err error) {
	msg.requestRef, err = isc.RequestRefFromBytes(data[1:])
	return err
}
