package chainimpl

import (
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/consensus"
	"github.com/iotaledger/wasp/packages/chain/statemgr"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/peering"
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
	ownerAddress    address.Address
	color           balance.Color
	peers           []*peering.Peer
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
}

func requestIDCaller(handler interface{}, params ...interface{}) {
	handler.(func(interface{}))(params[0])
}

func newCommitteeObj(chr *registry.ChainRecord, log *logger.Logger, onActivation func()) chain.Chain {
	log.Debugw("creating committee", "addr", chr.ChainID.String())

	addr := chr.ChainID
	if util.ContainsDuplicates(chr.CommitteeNodes) ||
		util.ContainsDuplicates(chr.AccessNodes) ||
		util.IntersectsLists(chr.CommitteeNodes, chr.AccessNodes) ||
		util.ContainsInList(peering.MyNetworkId(), chr.AccessNodes) {

		log.Errorf("can't create committee object for %s: chain record contains duplicate node addresses. Chain nodes: %+v",
			addr.String(), chr.CommitteeNodes)
		return nil
	}
	a := (address.Address)(chr.ChainID)
	dkshare, keyExists, err := registry.GetDKShare(&a)
	if err != nil {
		log.Error(err)
		return nil
	}

	if !keyExists {
		// if key doesn't exists, the node still can provide access to the smart contract state as an "access node"
		// for access nodes, committee nodes are ignored
		if len(chr.AccessNodes) > 0 {
			log.Info("can't find private key. Node will run as an access node for the address %s", addr.String())
		} else {
			log.Errorf("private key wasn't found and no access peers specified. Node can't run for the address %s", addr.String())
			return nil
		}
	} else {
		if !iAmInTheCommittee(chr.CommitteeNodes, dkshare.N, dkshare.Index) {
			log.Errorf("chain record inconsistency: the own node %s is not in the committee for %s: %+v",
				peering.MyNetworkId(), addr.String(), chr.CommitteeNodes)
			return nil
		}
		// check for owner address. It is mandatory for the committee node
		var niladdr address.Address
		if chr.OwnerAddress == niladdr {
			log.Errorf("undefined owner address for the committee node. Dismiss. Addr = %s", addr.String())
			return nil
		}
	}
	ret := &chainObj{
		procset:      processors.MustNew(),
		chMsg:        make(chan interface{}, 100),
		chainID:      chr.ChainID,
		ownerAddress: chr.OwnerAddress,
		color:        chr.Color,
		peers:        make([]*peering.Peer, 0),
		onActivation: onActivation,
		eventRequestProcessed: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ coretypes.RequestID))(params[0].(coretypes.RequestID))
		}),
		log: log.Named(util.Short(chr.ChainID.String())),
	}
	if keyExists {
		ret.ownIndex = dkshare.Index
		ret.size = dkshare.N
		ret.quorum = dkshare.T

		numNil := 0
		for _, remoteLocation := range chr.CommitteeNodes {
			peer := peering.UsePeer(remoteLocation)
			if peer == nil {
				numNil++
			}
			ret.peers = append(ret.peers, peer)
		}
		if numNil != 1 || ret.peers[dkshare.Index] != nil {
			// at this point must be exactly 1 element in ret.peers == to nil,
			// the one with the index in the committee
			ret.log.Panicf("failed to initialize peers of the committee. committeePeers: %+v. myId: %s", chr.CommitteeNodes, peering.MyNetworkId())
		}
	}
	for _, remoteLocation := range chr.AccessNodes {
		p := peering.UsePeer(remoteLocation)
		if p != nil {
			ret.peers = append(ret.peers, p)
		}
	}

	ret.stateMgr = statemgr.New(ret, ret.log)
	if keyExists {
		ret.operator = consensus.NewOperator(ret, dkshare, ret.log)
		ret.isCommitteeNode.Store(true)
	} else {
		ret.isCommitteeNode.Store(false)
	}
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

// iAmInTheCommittee checks if netLocations makes sense
func iAmInTheCommittee(committeeNodes []string, n, index uint16) bool {
	if len(committeeNodes) != int(n) {
		return false
	}
	return committeeNodes[index] == peering.MyNetworkId()
}
