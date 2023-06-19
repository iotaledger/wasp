// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

const (
	msgTypeCmtLog gpa.MessageType = iota
	msgTypeBlockProduced
)

func (cmi *chainMgrImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgTypeCmtLog:        func() gpa.Message { return new(msgCmtLog) },
		msgTypeBlockProduced: func() gpa.Message { return new(msgBlockProduced) },
	})
}
