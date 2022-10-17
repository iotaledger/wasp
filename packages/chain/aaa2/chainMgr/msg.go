// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainMgr

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

func (cl *chainMgrImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	msg := &msgCmtLog{} // Only the CmtLog messages can be exchanged here, hence no msgType is needed.
	if err := msg.UnmarshalBinary(data, cl.cmtLogs); err != nil {
		return nil, err
	}
	return msg, nil
}
