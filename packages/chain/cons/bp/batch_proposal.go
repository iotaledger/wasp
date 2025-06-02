// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bp

import (
	"time"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
)

type BatchProposal struct {
	nodeIndex               uint16               `bcs:"export"`          // Just for a double-check.
	baseAliasOutput         *isc.StateAnchor     `bcs:"export,optional"` // Proposed Base AliasOutput to use.
	dssIndexProposal        util.BitVector       `bcs:"export"`          // DSS Index proposal.
	rotateTo                *iotago.Address      `bcs:"export,optional"` // Suggestion to rotate the committee, optional.
	timeData                time.Time            `bcs:"export"`          // Our view of time.
	validatorFeeDestination isc.AgentID          `bcs:"export"`          // Proposed destination for fees.
	requestRefs             []*isc.RequestRef    `bcs:"export"`          // Requests we propose to include into the execution.
	gasCoins                []*coin.CoinWithRef  `bcs:"export,optional"` // Coins to use for gas payment.
	l1params                *parameters.L1Params `bcs:"export,optional"` // The L1Params for current state
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
	l1params *parameters.L1Params,
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
		l1params:                l1params,
	}
}

func (b *BatchProposal) Bytes() []byte {
	return bcs.MustMarshal(b)
}

// IsVoid returns true if a proposal is ⊥, in which case it will not contain request refs nor base AO.
// Other fields are required to help other participants to sign a TX, if such is produced from other node's inputs.
func (b *BatchProposal) IsVoid() bool {
	return b.baseAliasOutput == nil
}
