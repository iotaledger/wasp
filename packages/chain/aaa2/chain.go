// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package aaa2

import (
	"context"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain/aaa2/node"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/peering"
)

type ChainInfo struct { // TODO: ...
	ChainID *isc.ChainID
}

type Chain interface { // TODO: ...
	// Info() *ChainInfo
}

func New(
	ctx context.Context,
	chainID *isc.ChainID,
	nodeConn node.ChainNodeConn,
	nodeIdentity *cryptolib.KeyPair,
	net peering.NetworkProvider,
	log *logger.Logger,
) Chain {
	return node.New(ctx, chainID, nodeConn, nodeIdentity, net, log)
}

func ChainList() map[isc.ChainID]Chain {
	return nil // TODO: ...
}
