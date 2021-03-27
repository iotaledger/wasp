// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
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
	isReadyStateManager          bool
	isReadyConsensus             bool
	isConnectPeriodOver          bool
	isQuorumOfConnectionsReached bool
	mutexIsReady                 sync.Mutex
	isOpenQueue                  atomic.Bool
	dismissed                    atomic.Bool
	dismissOnce                  sync.Once
	onActivation                 func()
	//
	chainID         coretypes.ChainID
	procset         *processors.ProcessorCache
	peers           peering.GroupProvider
	size            uint16
	quorum          uint16
	ownIndex        uint16
	chMsg           chan interface{}
	stateMgr        chain.StateManager
	operator        chain.Operator
	isCommitteeNode atomic.Bool
	//
	eventRequestProcessed *events.Event
	log                   *logger.Logger
	netProvider           peering.NetworkProvider
	peersAttachRef        interface{}
	dksProvider           tcrypto.RegistryProvider
	blobProvider          coretypes.BlobCache
}

func newChainObj(
	chr *registry.ChainRecord,
	log *logger.Logger,
	netProvider peering.NetworkProvider,
	dksProvider tcrypto.RegistryProvider,
	blobProvider coretypes.BlobCache,
	onActivation func(),
) chain.Chain {
	var err error
	log.Debugf("creating chain: %s", chr)

	// TODO temporary. Committee address must come from chain transaction
	cmtRec, err := registry.CommitteeRecordFromRegistry(chr.StateAddressTmp)
	if err != nil || cmtRec == nil {
		log.Errorf("failed to lead committee record for address %s. err = %v", chr.StateAddressTmp, err)
		return nil
	}
	if util.ContainsDuplicates(cmtRec.CommitteeNodes) {
		log.Errorf("can't create chain object for %s: committee record contains duplicate node addresses: %+v",
			chr.StateAddressTmp, cmtRec.CommitteeNodes)
		return nil
	}
	dkshare, err := dksProvider.LoadDKShare(cmtRec.Address)
	if err != nil {
		log.Error(err)
		return nil
	}
	if dkshare.Index == nil || !iAmInTheCommittee(cmtRec.CommitteeNodes, dkshare.N, *dkshare.Index, netProvider) {
		log.Errorf(
			"chain record inconsistency: the own node %s is not in the committee for %s: %+v",
			netProvider.Self().NetID(), cmtRec.Address, cmtRec.CommitteeNodes,
		)
		return nil
	}
	var peers peering.GroupProvider
	if peers, err = netProvider.Group(cmtRec.CommitteeNodes); err != nil {
		log.Errorf(
			"node %s failed to setup committee communication with %+v, reason=%+v",
			netProvider.Self().NetID(), cmtRec.CommitteeNodes, err,
		)
		return nil
	}
	chainLog := log.Named(util.Short(chr.ChainID.String()))
	ret := &chainObj{
		procset:      processors.MustNew(),
		chMsg:        make(chan interface{}, 100),
		chainID:      chr.ChainID,
		peers:        peers,
		onActivation: onActivation,
		eventRequestProcessed: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ coretypes.RequestID))(params[0].(coretypes.RequestID))
		}),
		log:          chainLog,
		netProvider:  netProvider,
		dksProvider:  dksProvider,
		blobProvider: blobProvider,
	}
	ret.peersAttachRef = peers.Attach(&ret.chainID, func(recv *peering.RecvEvent) {
		ret.ReceiveMessage(recv.Msg)
	})

	ret.ownIndex = *dkshare.Index
	ret.size = dkshare.N
	ret.quorum = dkshare.T

	ret.stateMgr = statemgr.New(ret, ret.log)
	ret.operator = consensus.NewOperator(ret, dkshare, ret.log)
	ret.isCommitteeNode.Store(true)
	go func() {
		for msg := range ret.chMsg {
			ret.dispatchMessage(msg)
		}
	}()
	go func() {
		ret.log.Infof("wait for at least quorum of peers (%d) connected before activating the committee", ret.quorum)
		for !ret.HasQuorum() && !ret.IsDismissed() {
			time.Sleep(500 * time.Millisecond)
		}
		ret.log.Infof("peer status: %s", ret.PeerStatus())
		ret.SetQuorumOfConnectionsReached()

		go func() {
			ret.log.Infof("wait for %s more before activating the committee", chain.AdditionalConnectPeriod)
			time.Sleep(chain.AdditionalConnectPeriod)
			ret.log.Infof("connection period is over. Peer status: %s", ret.PeerStatus())

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
