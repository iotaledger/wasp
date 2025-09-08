// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainrunner

import (
	"github.com/iotaledger/wasp/v2/packages/chain"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/state"
)

type chainsListener struct {
	parent        chain.ChainListener
	accessNodesCB func(accessNodes []*cryptolib.PublicKey)
}

func NewChainsListener(parent chain.ChainListener, accessNodesCB func(accessNodes []*cryptolib.PublicKey)) chain.ChainListener {
	return &chainsListener{parent: parent, accessNodesCB: accessNodesCB}
}

func (cl *chainsListener) BlockApplied(block state.Block, latestState kv.KVStoreReader) {
	cl.parent.BlockApplied(block, latestState)
}

func (cl *chainsListener) AccessNodesUpdated(accessNodes []*cryptolib.PublicKey) {
	cl.accessNodesCB(accessNodes)
	cl.parent.AccessNodesUpdated(accessNodes)
}

func (cl *chainsListener) ServerNodesUpdated(serverNodes []*cryptolib.PublicKey) {
	cl.parent.ServerNodesUpdated(serverNodes)
}
