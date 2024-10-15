// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconn

import (
	"context"
	"fmt"
	"sync"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/cryptolib"
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

	return ncc
}

func (ncc *ncChain) WaitUntilStopped() {
	ncc.shutdownWaitGroup.Wait()
}

func (ncc *ncChain) postTxLoop(ctx context.Context) {
	defer ncc.shutdownWaitGroup.Done()

	postTx := func(task publishTxTask) error {
		txBytes, err := bcs.Marshal(task.tx.Data)
		if err != nil {
			return err
		}
		res, err := ncc.nodeConn.wsClient.ExecuteTransactionBlock(task.ctx, iotaclient.ExecuteTransactionBlockRequest{
			TxDataBytes: txBytes,
			Signatures:  task.tx.Signatures,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects: true,
			},
			RequestType: iotajsonrpc.TxnRequestTypeWaitForLocalExecution,
		})
		if err != nil {
			return err
		}
		if !res.Effects.Data.IsSuccess() {
			return fmt.Errorf("error executing tx: %s", res.Effects.Data.V1.Status.Error)
		}
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return
		case task := <-ncc.publishTxQueue:
			err := postTx(task)
			task.cb(task.tx, err)
		}
	}
}

func (ncc *ncChain) syncChainState(ctx context.Context) error {
	ncc.LogInfof("Synchronizing chain state for %s...", ncc.chainID)
	moveAnchor, reqs, err := ncc.feed.FetchCurrentState(ctx)
	if err != nil {
		return err
	}
	panic("FIXME the owner should not be empty")
	anchor := isc.NewStateAnchor(moveAnchor, nil, ncc.feed.GetISCPackageID())
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
		defer ncc.shutdownWaitGroup.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case moveAnchor := <-anchorUpdates:
				panic("FIXME the owner should not be empty")
				anchor := isc.NewStateAnchor(moveAnchor, nil, ncc.feed.GetISCPackageID())
				ncc.anchorHandler(&anchor)
			case req := <-newRequests:
				onledgerReq, err := isc.OnLedgerFromRequest(req, cryptolib.NewAddressFromIota(&anchorID))
				if err != nil {
					panic(err)
				}
				ncc.requestHandler(onledgerReq)
			}
		}
	}()
}
