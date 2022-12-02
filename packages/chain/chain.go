// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain/cmtLog"
	"github.com/iotaledger/wasp/packages/chain/node"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type ChainInfo struct { // TODO: ...
	ChainID *isc.ChainID
}

type Chain interface { // TODO: ...
	node.ChainNode
	// TODO: Info() *ChainInfo
	// TODO: OffLedgerRequest(req ...)
	// TODO: GetCurrentCommittee.
	// TODO: GetCurrentAccessNodes.
}

type NodeConnection interface {
	// TODO: node.ChainNodeConn
	GetMetrics() nodeconnmetrics.NodeConnectionMetrics
}

func New(
	ctx context.Context,
	chainID *isc.ChainID,
	chainStore state.Store,
	nodeConn node.ChainNodeConn,
	nodeIdentity *cryptolib.KeyPair,
	processorsConfig *processors.Config,
	net peering.NetworkProvider,
	log *logger.Logger,
) (Chain, error) {
	var dkRegistry registry.DKShareRegistryProvider // TODO: Get it somehow.
	var cmtLogStore cmtLog.Store                    // TODO: Get it somehow.
	var smBlockWAL smGPAUtils.BlockWAL              // TODO: Get it somehow.
	return node.New(ctx, chainID, chainStore, nodeConn, nodeIdentity, processorsConfig, dkRegistry, cmtLogStore, smBlockWAL, net, log)
}
