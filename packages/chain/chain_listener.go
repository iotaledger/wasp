// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
)

// Implementation of this interface will receive events in the chain.
// Initial intention was to provide info to the published/WebSocket endpoint.
// All the function MUST NOT BLOCK.
type ChainListener interface {
	mempool.ChainListener
}

////////////////////////////////////////////////////////////////////////////////
// emptyChainListener

type emptyChainListener struct{}

var _ ChainListener = &emptyChainListener{}

func NewEmptyChainListener() ChainListener {
	return &emptyChainListener{}
}

func (ecl *emptyChainListener) BlockApplied(chainID isc.ChainID, block state.Block) {}
