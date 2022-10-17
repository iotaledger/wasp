// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

// That's a local message.
type msgAliasOutputRejected struct {
	gpa.BasicMessage
	aliasOutput *isc.AliasOutputWithID
}

var _ gpa.Message = &msgAliasOutputRejected{}

func NewMsgAliasOutputRejected(recipient gpa.NodeID, aliasOutput *isc.AliasOutputWithID) gpa.Message {
	return &msgAliasOutputRejected{
		BasicMessage: gpa.NewBasicMessage(recipient),
		aliasOutput:  aliasOutput,
	}
}

func (m *msgAliasOutputRejected) MarshalBinary() ([]byte, error) {
	panic("trying to marshal local message")
}
