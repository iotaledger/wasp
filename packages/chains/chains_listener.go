// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chains

import (
	"github.com/iotaledger/wasp/v2/packages/chain"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/state"
)

type chainsListener struct {
	parent        chain.ChainListener
	accessNodesCB func(chainID isc.ChainID, accessNodes []*cryptolib.PublicKey)
}

func NewChainsListener(parent chain.ChainListener, accessNodesCB func(chainID isc.ChainID, accessNodes []*cryptolib.PublicKey)) chain.ChainListener {
	return &chainsListener{parent: parent, accessNodesCB: accessNodesCB}
}

func (cl *chainsListener) BlockApplied(chainID isc.ChainID, block state.Block, latestState kv.KVStoreReader) {
	cl.parent.BlockApplied(chainID, block, latestState)
}

func (cl *chainsListener) AccessNodesUpdated(chainID isc.ChainID, accessNodes []*cryptolib.PublicKey) {
	cl.accessNodesCB(chainID, accessNodes)
	cl.parent.AccessNodesUpdated(chainID, accessNodes)
}

func (cl *chainsListener) ServerNodesUpdated(chainID isc.ChainID, serverNodes []*cryptolib.PublicKey) {
	cl.parent.ServerNodesUpdated(chainID, serverNodes)
}
