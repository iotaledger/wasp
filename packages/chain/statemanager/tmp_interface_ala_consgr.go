// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// The purpose of this package is to run the consensus protocol
// as a goroutine and communicate with all the related components.
package statemanager

import (
	"context"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
)

// State manager has to implement this interface.
type ConsGrStateMgr interface {
	// State manager has to return a signal via the return channel when it
	// ensures all the needed blocks for the specified AliasOutput is present
	// in the database. Context is used to cancel a request.
	ConsensusStateProposal(
		ctx context.Context,
		aliasOutput *isc.AliasOutputWithID,
	) <-chan interface{}
	// State manager has to ensure all the data needed for the specified alias
	// output (presented as aliasOutputID+stateCommitment) is present in the DB.
	ConsensusDecidedState(
		ctx context.Context,
		aliasOutput *isc.AliasOutputWithID,
	) <-chan state.State
	// State manager has to persistently store the block and respond only after
	// the block was flushed to the disk. A WAL can be used for that as well.
	ConsensusProducedBlock(
		ctx context.Context,
		block state.StateDraft,
	) <-chan error
}
