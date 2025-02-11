// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bp

import (
	"time"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

type BatchProposal struct {
	nodeIndex               uint16              `bcs:"export"`          // Just for a double-check.
	baseAliasOutput         *isc.StateAnchor    `bcs:"export"`          // Proposed Base AliasOutput to use.
	dssIndexProposal        util.BitVector      `bcs:"export"`          // DSS Index proposal.
	rotateTo                *iotago.Address     `bcs:"export,optional"` // Suggestion to rotate the committee, optional.
	timeData                time.Time           `bcs:"export"`          // Our view of time.
	validatorFeeDestination isc.AgentID         `bcs:"export"`          // Proposed destination for fees.
	requestRefs             []*isc.RequestRef   `bcs:"export"`          // Requests we propose to include into the execution.
	gasCoins                []*coin.CoinWithRef `bcs:"export"`          // Coins to use for gas payment.
	gasPrice                uint64              `bcs:"export"`          // The gas price to use.
}

func NewBatchProposal(
	nodeIndex uint16,
	baseAliasOutput *isc.StateAnchor,
	dssIndexProposal util.BitVector,
	rotateTo *iotago.Address,
	timeData time.Time,
	validatorFeeDestination isc.AgentID,
	requestRefs []*isc.RequestRef,
	gasCoins []*coin.CoinWithRef,
	gasPrice uint64,
) *BatchProposal {
	return &BatchProposal{
		nodeIndex:               nodeIndex,
		baseAliasOutput:         baseAliasOutput,
		dssIndexProposal:        dssIndexProposal,
		rotateTo:                rotateTo,
		timeData:                timeData,
		validatorFeeDestination: validatorFeeDestination,
		requestRefs:             requestRefs,
		gasCoins:                gasCoins,
		gasPrice:                gasPrice,
	}
}

func (b *BatchProposal) Bytes() []byte {
	return bcs.MustMarshal(b)
}
