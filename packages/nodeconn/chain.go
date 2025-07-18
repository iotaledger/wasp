// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconn

import (
	"context"
	"fmt"
	"sync"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

// ncChain is responsible for maintaining the information related to a single chain.
type ncChain struct {
	log.Logger

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
	socketURL string,
	httpURL string,
) (*ncChain, error) {
	anchorAddress := chainID.AsAddress().AsIotaAddress()

	feed, err := iscmoveclient.NewChainFeed(
		ctx,
		nodeConn.iscPackageID,
		*anchorAddress,
		nodeConn.Logger,
		socketURL,
		httpURL,
	)
	if err != nil {
		return nil, err
	}

	ncc := &ncChain{
		Logger:         nodeConn.Logger,
		nodeConn:       nodeConn,
		chainID:        chainID,
		feed:           feed,
		requestHandler: requestHandler,
		anchorHandler:  anchorHandler,
		publishTxQueue: make(chan publishTxTask),
	}

	ncc.shutdownWaitGroup.Add(1)
	go ncc.postTxLoop(ctx)

	return ncc, nil
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

		// Executing the transaction via DryRun before posting to make sure the transaction is valid, as failed transactions cost gas!
		// Repeatedly failing transactions == sad gas coin
		dryRes, err := ncc.nodeConn.httpClient.DryRunTransaction(task.ctx, txBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to dry-run Anchor transaction: %w", err)
		}

		if dryRes == nil {
			return nil, fmt.Errorf("failed to dry-run Anchor transaction: response == nil")
		}

		if dryRes.Effects.Data.IsFailed() {
			return nil, fmt.Errorf("failed to dry-run Anchor transaction: response.Effects.Failed")
		}

		if dryRes.Effects.Data.IsSuccess() {
			ncc.LogDebug("successfully dry-run Anchor transaction")
		}

		res, err := ncc.nodeConn.httpClient.ExecuteTransactionBlock(task.ctx, iotaclient.ExecuteTransactionBlockRequest{
			TxDataBytes: txBytes,
			Signatures:  task.tx.Signatures,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowObjectChanges: true,
				ShowEffects:       true,
			},
			RequestType: iotajsonrpc.TxnRequestTypeWaitForLocalExecution,
		})

		if err != nil {
			ncc.LogErrorf("POSTING TX error: %v\n", err)
		} else {
			ncc.LogDebugf("POSTING TX response: %v\n", res)
		}

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

		anchor, err := ncc.nodeConn.httpClient.GetAnchorFromObjectID(ctx, anchorInfo.ObjectID)
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
		onLedgerReq, err := isc.OnLedgerFromMoveRequest(req, cryptolib.NewAddressFromIota(req.Owner))
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
	l1Params, err := ncc.nodeConn.L1ParamsFetcher().GetOrFetchLatest(ctx)
	if err != nil {
		return err
	}
	ncc.anchorHandler(&anchor, l1Params)

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
				l1Params, err := ncc.nodeConn.L1ParamsFetcher().GetOrFetchLatest(ctx)
				if err != nil {
					panic(err)
				}
				ncc.anchorHandler(&anchor, l1Params)
			case req := <-newRequests:
				onledgerReq, err := isc.OnLedgerFromMoveRequest(req, cryptolib.NewAddressFromIota(&anchorID))
				if err != nil {
					panic(err)
				}
				ncc.LogInfo("Incoming request ", req.ObjectID.String(), " ", onledgerReq.String(), " ", onledgerReq.ID().String())
				ncc.requestHandler(onledgerReq)
			}
		}
	}()
}
