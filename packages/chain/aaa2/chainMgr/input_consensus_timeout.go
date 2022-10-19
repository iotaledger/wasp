// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainMgr

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cmtLog"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputConsensusTimeout struct {
	committeeAddr iotago.Ed25519Address
	logIndex      cmtLog.LogIndex
}

func NewInputConsensusTimeout(committeeAddr iotago.Ed25519Address, logIndex cmtLog.LogIndex) gpa.Input {
	return &inputConsensusTimeout{
		committeeAddr: committeeAddr,
		logIndex:      logIndex,
	}
}
