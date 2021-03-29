// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	txstream "github.com/iotaledger/goshimmer/packages/txstream/client"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/vm/processors"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/consensus"
	"github.com/iotaledger/wasp/packages/chain/statemgr"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/util"
	"go.uber.org/atomic"
)

type chainObj struct {
	committee                    chain.Committee
	isReadyStateManager          bool
	isReadyConsensus             bool
	isConnectPeriodOver          bool
	isQuorumOfConnectionsReached bool
	mutexIsReady                 sync.Mutex
	isOpenQueue                  atomic.Bool
	dismissed                    atomic.Bool
	dismissOnce                  sync.Once
	onActivation                 func()
	chainID                      coretypes.ChainID
	procset                      *processors.ProcessorCache
	chMsg                        chan interface{}
	stateMgr                     chain.StateManager
	operator                     chain.Consensus
	isCommitteeNode              atomic.Bool
	eventRequestProcessed        *events.Event
	log                          *logger.Logger
	nodeConn                     *txstream.Client
	netProvider                  peering.NetworkProvider
	dksProvider                  tcrypto.RegistryProvider
	blobProvider                 coretypes.BlobCache
}

func newChainObj(
	chr *registry.ChainRecord,
	log *logger.Logger,
	nodeConn *txstream.Client,
	netProvider peering.NetworkProvider,
	dksProvider tcrypto.RegistryProvider,
	blobProvider coretypes.BlobCache,
	onActivation func(),
) chain.Chain {
	var err error
	log.Debugf("creating chain: %s", chr)

	chainLog := log.Named(util.Short(chr.ChainID.String()))
	ret := &chainObj{
		procset:      processors.MustNew(),
		chMsg:        make(chan interface{}, 100),
		chainID:      chr.ChainID,
		onActivation: onActivation,
		eventRequestProcessed: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ coretypes.RequestID))(params[0].(coretypes.RequestID))
		}),
		log:          chainLog,
		nodeConn:     nodeConn,
		netProvider:  netProvider,
		dksProvider:  dksProvider,
		blobProvider: blobProvider,
	}
	ret.committee, err = NewCommittee(ret, chr.StateAddressTmp, netProvider, dksProvider)
	if err != nil {
		log.Errorf("failed to create chain. ChainID: %s: %v", chr.ChainID, err)
		return nil
	}
	ret.committee.OnPeerMessage(func(recv *peering.RecvEvent) {
		ret.ReceiveMessage(recv.Msg)
	})

	ret.stateMgr = statemgr.New(ret, ret.log)
	ret.operator = consensus.New(ret.committee, nodeConn, ret.log)
	ret.isCommitteeNode.Store(true)
	go func() {
		for msg := range ret.chMsg {
			ret.dispatchMessage(msg)
		}
	}()
	go func() {
		ret.log.Infof("wait for at least quorum of peers (%d) connected before activating the committee", ret.committee.Quorum())
		for !ret.committee.QuorumIsAlive() && !ret.IsDismissed() {
			time.Sleep(500 * time.Millisecond)
		}
		ret.log.Infof("peer status: %s", ret.committee.PeerStatus())
		ret.SetQuorumOfConnectionsReached()

		go func() {
			ret.log.Infof("wait for %s more before activating the committee", chain.AdditionalConnectPeriod)
			time.Sleep(chain.AdditionalConnectPeriod)
			ret.log.Infof("connection period is over. Peer status: %s", ret.committee.PeerStatus())

			ret.SetConnectPeriodOver()
		}()
	}()
	return ret
}

// iAmInTheCommittee checks if NetIDs makes sense
func iAmInTheCommittee(committeeNodes []string, n, index uint16, netProvider peering.NetworkProvider) bool {
	if len(committeeNodes) != int(n) {
		return false
	}
	return committeeNodes[index] == netProvider.Self().NetID()
}
