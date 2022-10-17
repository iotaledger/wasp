// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainMgr

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

// That's a local message.
type msgAliasOutputConfirmed struct {
	gpa.BasicMessage
	aliasOutput *isc.AliasOutputWithID
}

var _ gpa.Message = &msgAliasOutputConfirmed{}

func NewMsgAliasOutputConfirmed(recipient gpa.NodeID, aliasOutput *isc.AliasOutputWithID) gpa.Message {
	return &msgAliasOutputConfirmed{
		BasicMessage: gpa.NewBasicMessage(recipient),
		aliasOutput:  aliasOutput,
	}
}

func (m *msgAliasOutputConfirmed) MarshalBinary() ([]byte, error) {
	panic("trying to marshal local message")
}
