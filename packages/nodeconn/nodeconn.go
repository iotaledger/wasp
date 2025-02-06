// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// nodeconn package provides an interface to the L1 node (Hornet).
package nodeconn

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/chain/cons/cons_gr"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/transaction"

	"github.com/iotaledger/hive.go/app/shutdown"
	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
)

const (
	l1NodeSyncWaitTimeout = 2 * time.Minute

	chainsCleanupThresholdRatio = 50.0
	chainsCleanupThresholdCount = 10
)

var ErrOperationAborted = errors.New("operation was aborted")

type SingleGasCoinInfo struct {
	gasCoinObject coin.CoinWithRef
	gasPrice      uint64
}

func (g *SingleGasCoinInfo) GetGasCoins() []*coin.CoinWithRef {
	return []*coin.CoinWithRef{&g.gasCoinObject}
}

func (g *SingleGasCoinInfo) GetGasPrice() uint64 {
	return g.gasPrice
}

// nodeConnection implements chain.NodeConnection.
// Single Wasp node is expected to connect to a single L1 node, thus
// we expect to have a single instance of this structure.
type nodeConnection struct {
	*logger.WrappedLogger

	iscPackageID iotago.PackageID
	wsClient     *iscmoveclient.Client

	synced     sync.WaitGroup
	chainsLock sync.RWMutex
	chainsMap  *shrinkingmap.ShrinkingMap[isc.ChainID, *ncChain]

	shutdownHandler *shutdown.ShutdownHandler
}

var _ chain.NodeConnection = &nodeConnection{}

func New(
	ctx context.Context,
	iscPackageID iotago.PackageID,
	wsURL string,
	log *logger.Logger,
	shutdownHandler *shutdown.ShutdownHandler,
) (chain.NodeConnection, error) {
	wsClient, err := iscmoveclient.NewWebsocketClient(ctx, wsURL, "", iotaclient.WaitForEffectsDisabled, log)
	if err != nil {
		return nil, err
	}
	return &nodeConnection{
		WrappedLogger: logger.NewWrappedLogger(log),
		iscPackageID:  iscPackageID,
		wsClient:      wsClient,
		chainsMap: shrinkingmap.New[isc.ChainID, *ncChain](
			shrinkingmap.WithShrinkingThresholdRatio(chainsCleanupThresholdRatio),
			shrinkingmap.WithShrinkingThresholdCount(chainsCleanupThresholdCount),
		),
		shutdownHandler: shutdownHandler,
	}, nil
}

func (nc *nodeConnection) AttachChain(
	ctx context.Context,
	chainID isc.ChainID,
	recvRequest chain.RequestHandler,
	recvAnchor chain.AnchorHandler,
	onChainConnect func(),
	onChainDisconnect func(),
) {
	ncc := func() *ncChain {
		nc.chainsLock.Lock()
		defer nc.chainsLock.Unlock()

		ncc := newNCChain(ctx, nc, chainID, recvRequest, recvAnchor)

		nc.chainsMap.Set(chainID, ncc)
		util.ExecuteIfNotNil(onChainConnect)
		nc.LogDebugf("chain registered: %s = %s", chainID.ShortString(), chainID)

		return ncc
	}()

	if err := ncc.syncChainState(ctx); err != nil {
		nc.LogError(fmt.Sprintf("synchronizing chain state %s failed: %s", chainID, err.Error()))
		nc.shutdownHandler.SelfShutdown(
			fmt.Sprintf("Cannot sync chain %s with L1, %s", ncc.chainID, err.Error()),
			true)
	}

	subscriptionCtx, cancelSubscriptions := context.WithCancel(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ncc.feed.ConnectionRecreated():
				nc.LogInfo("Connection recreated, resubscribing to updates")
				// oldCancel := cancelSubscriptions
				cancelSubscriptions()

				nc.LogInfo("subscriptions cancelled")
				// TODO: REMOVE THIS HACK, probably we can run cancel in parallel
				time.Sleep(2 * time.Second)
				nc.LogInfo("resubscribing to updates")
				subscriptionCtx, cancelSubscriptions = context.WithCancel(ctx)
				ncc.subscribeToUpdates(subscriptionCtx, chainID.AsObjectID())
				nc.LogInfo("resubscribed to updates")
				// oldCancel()
			}
		}
	}()

	ncc.subscribeToUpdates(subscriptionCtx, chainID.AsObjectID())

	// disconnect the chain after the context is done
	go func() {
		<-ctx.Done()
		ncc.WaitUntilStopped()

		nc.chainsLock.Lock()
		defer nc.chainsLock.Unlock()

		nc.chainsMap.Delete(chainID)
		util.ExecuteIfNotNil(onChainDisconnect)
		nc.LogDebugf("chain unregistered: %s = %s, |remaining|=%v", chainID.ShortString(), chainID, nc.chainsMap.Size())
	}()
}

func (nc *nodeConnection) GetGasCoinRef(ctx context.Context, chainID isc.ChainID) (*coin.CoinWithRef, error) {
	ncChain, ok := nc.chainsMap.Get(chainID)
	if !ok {
		panic("unexpected chainID")
	}
	gasCoinRef, gasCoinBal, err := ncChain.feed.GetChainGasCoin(ctx)
	if err != nil {
		return nil, err
	}
	return &coin.CoinWithRef{
		Type:  coin.BaseTokenType,
		Value: coin.Value(gasCoinBal),
		Ref:   gasCoinRef,
	}, nil
}

func (nc *nodeConnection) ConsensusGasPriceProposal(
	ctx context.Context,
	anchor *isc.StateAnchor,
) <-chan cons_gr.NodeConnGasInfo {
	t := make(chan cons_gr.NodeConnGasInfo)

	// TODO: Refactor this separate goroutine and place it somewhere connection related instead
	go func() {
		stateMetadata, err := transaction.StateMetadataFromBytes(anchor.GetStateMetadata())
		if err != nil {
			panic(err)
		}

		gasCoinGetObjectRes, err := nc.wsClient.GetObject(ctx, iotaclient.GetObjectRequest{
			ObjectID: stateMetadata.GasCoinObjectID,
			Options:  &iotajsonrpc.IotaObjectDataOptions{ShowBcs: true},
		})
		if err != nil {
			panic(err)
		}

		var gasCoin iscmoveclient.MoveCoin
		err = iotaclient.UnmarshalBCS(gasCoinGetObjectRes.Data.Bcs.Data.MoveObject.BcsBytes, &gasCoin)
		if err != nil {
			panic(err)
		}

		gasCoinRef := gasCoinGetObjectRes.Data.Ref()
		var coinInfo cons_gr.NodeConnGasInfo = &SingleGasCoinInfo{
			coin.CoinWithRef{
				Type:  coin.BaseTokenType,
				Value: coin.Value(gasCoin.Balance),
				Ref:   &gasCoinRef,
			},
			// FIXME we will pass l1param from consensus
			parameters.L1().Protocol.ReferenceGasPrice.Uint64(),
		}

		t <- coinInfo
	}()

	return t
}

func (nc *nodeConnection) RefreshOnLedgerRequests(ctx context.Context, chainID isc.ChainID) {
	ncChain, ok := nc.chainsMap.Get(chainID)
	if !ok {
		panic("unexpected chainID")
	}
	if err := ncChain.syncChainState(ctx); err != nil {
		nc.LogError(fmt.Sprintf("error refreshing outputs: %s", err.Error()))
	}
}

// TODO is this still needed?
func (nc *nodeConnection) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (nc *nodeConnection) WaitUntilInitiallySynced(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			_, err := nc.wsClient.GetLatestIotaSystemState(ctx)
			if err != nil {
				nc.LogWarnf("WaitUntilInitiallySynced: %s", err)
				continue
			}
			return nil
		}
	}
}

func (nc *nodeConnection) GetL1Params() *parameters.L1Params {
	panic("TODO")
	// return nc.l1Params
}

// GetChain returns the chain if it was registered, otherwise it returns an error.
func (nc *nodeConnection) getChain(chainID isc.ChainID) (*ncChain, error) {
	nc.chainsLock.RLock()
	defer nc.chainsLock.RUnlock()

	ncc, exists := nc.chainsMap.Get(chainID)
	if !exists {
		return nil, fmt.Errorf("chain %v is not connected", chainID.String())
	}
	return ncc, nil
}

func (nc *nodeConnection) PublishTX(
	ctx context.Context,
	chainID isc.ChainID,
	tx iotasigner.SignedTransaction,
	callback chain.TxPostHandler,
) error {
	// check if the chain exists
	ncc, err := nc.getChain(chainID)
	if err != nil {
		return err
	}
	ncc.publishTxQueue <- publishTxTask{
		ctx: ctx,
		tx:  tx,
		cb:  callback,
	}
	return nil
}
