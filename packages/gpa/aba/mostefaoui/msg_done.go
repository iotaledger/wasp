// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"io"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type msgDone struct {
	gpa.BasicMessage
	round int
}

var _ gpa.Message = new(msgDone)

func multicastMsgDone(recipients []gpa.NodeID, me gpa.NodeID, round int) gpa.OutMessages {
	msgs := gpa.NoMessages()
	for _, recipient := range recipients {
		if recipient != me {
			msgs.Add(&msgDone{
				BasicMessage: gpa.NewBasicMessage(recipient),
				round:        round,
			})
		}
	}
	return msgs
}

func (msg *msgDone) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msgTypeDone.ReadAndVerify(rr)
	msg.round = int(rr.ReadUint16())
	return rr.Err
}

func (msg *msgDone) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	msgTypeDone.Write(ww)
	ww.WriteUint16(uint16(msg.round))
	return ww.Err
}
