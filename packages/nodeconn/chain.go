// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconn

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util/bcs"

	"github.com/iotaledger/hive.go/logger"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
)

// ncChain is responsible for maintaining the information related to a single chain.
type ncChain struct {
	*logger.WrappedLogger

	nodeConn       *nodeConnection
	chainID        isc.ChainID
	feed           *iscmoveclient.ChainFeed
	requestHandler chain.RequestHandler
	anchorHandler  chain.AnchorHandler

	publishTxQueue chan publishTxTask

	shutdownWaitGroup sync.WaitGroup
}

type publishTxTask struct {
	ctx context.Context
	tx  iotasigner.SignedTransaction
	cb  chain.TxPostHandler
}

func newNCChain(
	ctx context.Context,
	nodeConn *nodeConnection,
	chainID isc.ChainID,
	requestHandler chain.RequestHandler,
	anchorHandler chain.AnchorHandler,
) *ncChain {
	anchorAddress := chainID.AsAddress().AsIotaAddress()
	feed := iscmoveclient.NewChainFeed(
		ctx,
		nodeConn.wsClient,
		nodeConn.iscPackageID,
		*anchorAddress,
		nodeConn.Logger(),
	)
	ncc := &ncChain{
		WrappedLogger:  logger.NewWrappedLogger(nodeConn.Logger()),
		nodeConn:       nodeConn,
		chainID:        chainID,
		feed:           feed,
		requestHandler: requestHandler,
		anchorHandler:  anchorHandler,
		publishTxQueue: make(chan publishTxTask),
	}

	ncc.shutdownWaitGroup.Add(1)
	go ncc.postTxLoop(ctx)

	// FIXME make timeout configurable
	// FIXME this will be replaced by passing l1param from consensus
	l1syncer := parameters.NewL1Syncer(nodeConn.wsClient.Client, 600*time.Second, nodeConn.Logger())
	go l1syncer.Start()

	return ncc
}

func (ncc *ncChain) WaitUntilStopped() {
	ncc.shutdownWaitGroup.Wait()
}

func (ncc *ncChain) postTxLoop(ctx context.Context) {
	defer ncc.shutdownWaitGroup.Done()

	postTx := func(task publishTxTask) (*isc.StateAnchor, error) {
		txBytes, err := bcs.Marshal(task.tx.Data)
		if err != nil {
			return nil, err
		}
		res, err := ncc.nodeConn.wsClient.ExecuteTransactionBlock(task.ctx, iotaclient.ExecuteTransactionBlockRequest{
			TxDataBytes: txBytes,
			Signatures:  task.tx.Signatures,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowObjectChanges:  true,
				ShowBalanceChanges: true,
				ShowEffects:        true,
			},
			RequestType: iotajsonrpc.TxnRequestTypeWaitForLocalExecution,
		})

		ncc.LogDebug("POSTING TX")
		ncc.LogDebugf("%v %v\n", res, err)

		if err != nil {
			return nil, err
		}
		if !res.Effects.Data.IsSuccess() {
			return nil, fmt.Errorf("error executing tx: %s Digest: %s", res.Effects.Data.V1.Status.Error, res.Digest)
		}

		anchorInfo, err := res.GetMutatedObjectInfo(iscmove.AnchorModuleName, iscmove.AnchorObjectName)
		if err != nil {
			return nil, err
		}

		anchor, err := ncc.nodeConn.wsClient.GetAnchorFromObjectID(ctx, anchorInfo.ObjectID)
		if err != nil {
			return nil, err
		}

		stateAnchor := isc.NewStateAnchor(anchor, ncc.nodeConn.iscPackageID)

		return &stateAnchor, nil
	}

	for {
		select {
		case <-ctx.Done():
			return
		case task := <-ncc.publishTxQueue:
			stateAnchor, err := postTx(task)
			task.cb(task.tx, stateAnchor, err)
		}
	}
}

func (ncc *ncChain) syncChainState(ctx context.Context) error {
	ncc.LogInfof("Synchronizing chain state for %s...", ncc.chainID)

	moveAnchor, err := ncc.feed.FetchCurrentState(ctx, ncc.nodeConn.maxNumberOfRequests, func(err error, req *iscmove.RefWithObject[iscmove.Request]) {
		if err != nil {
			return
		}

		// The owner will always be the Anchor, so instead of pulling the Anchor and using its ID
		// the owner address will be used.
		onLedgerReq, err := isc.OnLedgerFromRequest(req, cryptolib.NewAddressFromIota(req.Owner))
		if err != nil {
			return
		}

		ncc.LogDebugf("Sending %s to request handler", req.ObjectID)
		ncc.requestHandler(onLedgerReq)
	})
	if err != nil {
		return err
	}

	anchor := isc.NewStateAnchor(moveAnchor, ncc.feed.GetISCPackageID())
	ncc.anchorHandler(&anchor)

	ncc.LogInfof("Synchronizing chain state for %s... done", ncc.chainID)
	return nil
}

func (ncc *ncChain) subscribeToUpdates(ctx context.Context, anchorID iotago.ObjectID) {
	anchorUpdates := make(chan *iscmove.AnchorWithRef)
	newRequests := make(chan *iscmove.RefWithObject[iscmove.Request])
	ncc.feed.SubscribeToUpdates(ctx, anchorID, anchorUpdates, newRequests)

	ncc.shutdownWaitGroup.Add(1)
	go func() {
		defer ncc.shutdownWaitGroup.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case moveAnchor := <-anchorUpdates:
				anchor := isc.NewStateAnchor(moveAnchor, ncc.feed.GetISCPackageID())
				ncc.anchorHandler(&anchor)
			case req := <-newRequests:
				onledgerReq, err := isc.OnLedgerFromRequest(req, cryptolib.NewAddressFromIota(&anchorID))
				if err != nil {
					panic(err)
				}
				ncc.LogInfo("Incoming request ", req.ObjectID.String(), " ", onledgerReq.String(), " ", onledgerReq.ID().String())
				ncc.requestHandler(onledgerReq)
			}
		}
	}()
}
