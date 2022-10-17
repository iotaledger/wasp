// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainMgr

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type msgConsensusOutput struct {
	gpa.BasicMessage
	committeeID       CommitteeID
	logIndex          journal.LogIndex
	baseAliasOutputID iotago.OutputID
	nextAliasOutput   *isc.AliasOutputWithID
}

var _ gpa.Message = &msgConsensusOutput{}

func NewMsgConsensusOutput(
	recipient gpa.NodeID,
	committeeID CommitteeID,
	logIndex journal.LogIndex,
	baseAliasOutputID iotago.OutputID,
	nextAliasOutput *isc.AliasOutputWithID,
) gpa.Message {
	return &msgConsensusOutput{
		BasicMessage:      gpa.NewBasicMessage(recipient),
		committeeID:       committeeID,
		logIndex:          logIndex,
		baseAliasOutputID: baseAliasOutputID,
		nextAliasOutput:   nextAliasOutput,
	}
}

func (m *msgConsensusOutput) MarshalBinary() ([]byte, error) {
	panic("that's local message")
}
