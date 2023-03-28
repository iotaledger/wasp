// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainMgr

import (
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util"
)

const (
	msgTypeCmtLog byte = iota
	msgTypeBlockProduced
)

func (cl *chainMgrImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	if len(data) < 1 {
		return nil, errors.New("chainMgr::UnmarshalMessage: data too short")
	}
	var msg gpa.Message
	switch data[0] {
	case msgTypeCmtLog:
		msg = &msgCmtLog{}
	case msgTypeBlockProduced:
		msg = &msgBlockProduced{}
	default:
		return nil, fmt.Errorf("chainMgr::UnmarshalMessage: cannot parse message starting with: %v", util.PrefixHex(data, 20))
	}
	if err := msg.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return msg, nil
}
