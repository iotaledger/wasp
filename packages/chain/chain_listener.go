// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"github.com/iotaledger/wasp/v2/packages/chain/mempool"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/state"
)

// ChainListener is an interface whose implementation will receive events in the chain.
// Initial intention was to provide info to the published/WebSocket endpoint.
// All the function MUST NOT BLOCK.
type ChainListener interface {
	mempool.ChainListener
	AccessNodesUpdated(accessNodes []*cryptolib.PublicKey)
	ServerNodesUpdated(serverNodes []*cryptolib.PublicKey)
}

////////////////////////////////////////////////////////////////////////////////
// emptyChainListener

type emptyChainListener struct{}

var _ ChainListener = &emptyChainListener{}

func NewEmptyChainListener() ChainListener {
	return &emptyChainListener{}
}

func (ecl *emptyChainListener) BlockApplied(block state.Block, latestState kv.KVStoreReader) {
}
func (ecl *emptyChainListener) AccessNodesUpdated([]*cryptolib.PublicKey) {}
func (ecl *emptyChainListener) ServerNodesUpdated([]*cryptolib.PublicKey) {}
