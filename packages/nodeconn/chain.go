// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconn

import (
	"context"
	"fmt"
	"sync"

	"github.com/fardream/go-bcs/bcs"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
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
	tx  chain.SignedTx
	cb  chain.TxPostHandler
}

func newNCChain(
	ctx context.Context,
	nodeConn *nodeConnection,
	chainID isc.ChainID,
	requestHandler chain.RequestHandler,
	anchorHandler chain.AnchorHandler,
) *ncChain {
	anchorAddress := chainID.AsAddress().AsSuiAddress()
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
		res, err := ncc.nodeConn.wsClient.ExecuteTransactionBlock(task.ctx, suiclient.ExecuteTransactionBlockRequest{
			TxDataBytes: txBytes,
			Signatures:  task.tx.Signatures,
			Options: &suijsonrpc.SuiTransactionBlockResponseOptions{
				ShowEffects: true,
			},
			RequestType: suijsonrpc.TxnRequestTypeWaitForLocalExecution,
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
	anchorRef, reqs, err := ncc.feed.FetchCurrentState(ctx)
	if err != nil {
		return err
	}
	ncc.anchorHandler(anchorRef)
	for _, req := range reqs {
		ncc.requestHandler(req)
	}
	ncc.LogInfof("Synchronizing chain state for %s... done", ncc.chainID)
	return nil
}

func (ncc *ncChain) subscribeToUpdates(ctx context.Context) {
	anchorUpdates := make(chan *iscmove.RefWithObject[iscmove.Anchor])
	newRequests := make(chan *iscmove.Request)
	ncc.feed.SubscribeToUpdates(ctx, anchorUpdates, newRequests)

	ncc.shutdownWaitGroup.Add(1)
	go func() {
		defer ncc.shutdownWaitGroup.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case anchorRef := <-anchorUpdates:
				ncc.anchorHandler(anchorRef)
			case req := <-newRequests:
				ncc.requestHandler(req)
			}
		}
	}()
}
