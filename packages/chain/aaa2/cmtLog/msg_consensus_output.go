// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type msgConsensusOutput struct {
	gpa.BasicMessage
	logIndex          journal.LogIndex
	baseAliasOutputID iotago.OutputID
	nextAliasOutput   *isc.AliasOutputWithID
}

var _ gpa.Message = &msgConsensusOutput{}

// This message is internal one, but should be sent by other components (e.g. consensus or the chain).
func NewMsgConsensusOutput(
	recipient gpa.NodeID,
	logIndex journal.LogIndex,
	baseAliasOutputID iotago.OutputID,
	nextAliasOutput *isc.AliasOutputWithID,
) gpa.Message {
	return &msgConsensusOutput{
		BasicMessage:      gpa.NewBasicMessage(recipient),
		logIndex:          logIndex,
		baseAliasOutputID: baseAliasOutputID,
		nextAliasOutput:   nextAliasOutput,
	}
}

func (m *msgConsensusOutput) MarshalBinary() ([]byte, error) {
	panic("trying to marshal local message")
}
