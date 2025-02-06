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

const MaxRetriesGetAnchorAfterPostTX = 5

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

func (ncc *ncChain) retryGetTransactionBlock(
	ctx context.Context,
	digest *iotago.TransactionDigest,
	maxAttempts int,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		res, err := ncc.nodeConn.wsClient.GetTransactionBlock(ctx, iotaclient.GetTransactionBlockRequest{
			Digest: digest,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowObjectChanges:  true,
				ShowBalanceChanges: true,
				ShowEffects:        true,
			},
		})

		if err == nil && (res.Effects != nil && res.Effects.Data.IsSuccess()) {
			return res, nil
		}

		// Log the error
		ncc.LogInfof("Anchor GetTransactionBlock attempt %d/%d failed, err=%v", attempt, maxAttempts, err)

		// If this was our last attempt, return the error
		if attempt == maxAttempts {
			return nil, fmt.Errorf("failed after %d attempts: %w", maxAttempts, err)
		}

		// Wait before the next retry
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
			continue
		}
	}

	// This should never be reached due to the return in the loop
	return nil, fmt.Errorf("unexpected error in retry logic")
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
		if err != nil {
			return nil, err
		}
		if !res.Effects.Data.IsSuccess() {
			return nil, fmt.Errorf("error executing tx: %s Digest: %s", res.Effects.Data.V1.Status.Error, res.Digest)
		}

		res, err = ncc.retryGetTransactionBlock(ctx, &res.Digest, MaxRetriesGetAnchorAfterPostTX)
		if err != nil {
			return nil, err
		}

		if err != nil {
			ncc.LogInfof("GetTransactionBlock, err=%v", err)
			return nil, err
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
	moveAnchor, reqs, err := ncc.feed.FetchCurrentState(ctx)
	if err != nil {
		return err
	}
	anchor := isc.NewStateAnchor(moveAnchor, ncc.feed.GetISCPackageID())
	ncc.anchorHandler(&anchor)

	for _, req := range reqs {
		onledgerReq, err := isc.OnLedgerFromRequest(req, cryptolib.NewAddressFromIota(moveAnchor.ObjectID))
		if err != nil {
			return err
		}
		ncc.requestHandler(onledgerReq)
	}

	ncc.LogInfof("Synchronizing chain state for %s... done", ncc.chainID)
	return nil
}

func (ncc *ncChain) subscribeToUpdates(ctx context.Context, anchorID iotago.ObjectID) {
	anchorUpdates := make(chan *iscmove.AnchorWithRef)
	newRequests := make(chan *iscmove.RefWithObject[iscmove.Request])

	ncc.feed.SubscribeToUpdates(ctx, anchorID, anchorUpdates, newRequests)

	ncc.shutdownWaitGroup.Add(1)
	go func() {
		ncc.LogInfo("subscribeToUpdates: loop started")
		defer ncc.LogInfo("subscribeToUpdates: loop exited")
		defer ncc.shutdownWaitGroup.Done()

		for anchorUpdates != nil || newRequests != nil {
			select {
			case moveAnchor, ok := <-anchorUpdates:
				if !ok {
					anchorUpdates = nil
					continue
				}
				anchor := isc.NewStateAnchor(moveAnchor, ncc.feed.GetISCPackageID())
				ncc.anchorHandler(&anchor)
			case req, ok := <-newRequests:
				if !ok {
					newRequests = nil
					continue
				}
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
