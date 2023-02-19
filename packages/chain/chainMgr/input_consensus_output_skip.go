// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainMgr

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cmtLog"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputConsensusOutputSkip struct {
	committeeAddr     iotago.Ed25519Address
	logIndex          cmtLog.LogIndex
	baseAliasOutputID iotago.OutputID
}

func NewInputConsensusOutputSkip(
	committeeAddr iotago.Ed25519Address,
	logIndex cmtLog.LogIndex,
	baseAliasOutputID iotago.OutputID,
) gpa.Input {
	return &inputConsensusOutputSkip{
		committeeAddr:     committeeAddr,
		logIndex:          logIndex,
		baseAliasOutputID: baseAliasOutputID,
	}
}

func (inp *inputConsensusOutputSkip) String() string {
	return fmt.Sprintf(
		"{chainMgr.inputConsensusOutputSkip, committeeAddr=%v, logIndex=%v, baseAliasOutputID=%v}",
		inp.committeeAddr.String(),
		inp.logIndex,
		inp.baseAliasOutputID.ToHex(),
	)
}
