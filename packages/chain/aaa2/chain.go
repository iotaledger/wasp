// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package aaa2

import (
	"context"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cmtLog"
	"github.com/iotaledger/wasp/packages/chain/aaa2/node"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/tcrypto"
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

func New(
	ctx context.Context,
	chainID *isc.ChainID,
	nodeConn node.ChainNodeConn,
	nodeIdentity *cryptolib.KeyPair,
	processorsConfig *processors.Config,
	net peering.NetworkProvider,
	log *logger.Logger,
) (Chain, error) {
	var dkRegistry tcrypto.DKShareRegistryProvider // TODO: Get it somehow.
	var cmtLogStore cmtLog.Store                   // TODO: Get it somehow.
	var smBlockWAL smGPAUtils.BlockWAL             // TODO: Get it somehow.
	var chainKVStore kvstore.KVStore               // TODO: Get it somehow.
	return node.New(ctx, chainID, chainKVStore, nodeConn, nodeIdentity, processorsConfig, dkRegistry, cmtLogStore, smBlockWAL, net, log)
}

func ChainList() map[isc.ChainID]Chain {
	return nil // TODO: ...
}
