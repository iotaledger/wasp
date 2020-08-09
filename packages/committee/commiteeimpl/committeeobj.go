package commiteeimpl

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/committee/consensus"
	"github.com/iotaledger/wasp/packages/committee/statemgr"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/peering"
	"go.uber.org/atomic"
	"sync"
)

type committeeObj struct {
	isReadyStateManager bool
	isReadyConsensus    bool
	mutexIsReady        sync.Mutex
	isOpenQueue         atomic.Bool
	dismissed           atomic.Bool
	dismissOnce         sync.Once
	//
	params       *committee.Parameters
	address      address.Address
	ownerAddress address.Address
	color        balance.Color
	peers        []*peering.Peer
	size         uint16
	ownIndex     uint16
	chMsg        chan interface{}
	stateMgr     committee.StateManager
	operator     committee.Operator
	log          *logger.Logger
}

func newCommitteeObj(bootupData *registry.BootupData, log *logger.Logger, params *committee.Parameters) committee.Committee {
	log.Debugw("creating committee", "addr", bootupData.Address.String())

	addr := bootupData.Address
	if util.ContainsDuplicates(bootupData.CommitteeNodes) ||
		util.ContainsDuplicates(bootupData.AccessNodes) ||
		util.IntersectsLists(bootupData.CommitteeNodes, bootupData.AccessNodes) ||
		util.ContainsInList(peering.MyNetworkId(), bootupData.AccessNodes) {

		log.Errorf("can't create committee object for %s: bootup data contains duplicate node addresses", addr.String())
		return nil
	}
	dkshare, keyExists, err := registry.GetDKShare(&bootupData.Address)
	if err != nil {
		log.Error(err)
		return nil
	}

	if !keyExists {
		// if key doesn't exists, the node still can provide access to the smart contract state as an "access node"
		// for access nodes, committee nodes are ignored
		if len(bootupData.AccessNodes) > 0 {
			log.Info("can't find private key. Node will run as an access node for the address %s", addr.String())
		} else {
			log.Errorf("private key wasn't found and no access peers specified. Node can't run for the address %s", addr.String())
			return nil
		}
	} else {
		if !iAmInTheCommittee(bootupData.CommitteeNodes, dkshare.N, dkshare.Index) {
			log.Errorf("bootup data inconsistency: the own node %s is not in the committee for %s: %+v",
				peering.MyNetworkId(), addr.String(), bootupData.CommitteeNodes)
			return nil
		}
		// check for owner address. It is mandatory for the committee node
		var niladdr address.Address
		if bootupData.OwnerAddress == niladdr {
			log.Errorf("undefined owner address for the committee node. Dismiss. Addr = %s", addr.String())
			return nil
		}
	}

	ret := &committeeObj{
		chMsg:        make(chan interface{}, 100),
		params:       params,
		address:      bootupData.Address,
		ownerAddress: bootupData.OwnerAddress,
		color:        bootupData.Color,
		peers:        make([]*peering.Peer, 0),
		log:          log.Named(util.Short(bootupData.Address.String())),
	}
	if keyExists {
		ret.ownIndex = dkshare.Index
		ret.size = dkshare.N

		for _, remoteLocation := range bootupData.CommitteeNodes {
			ret.peers = append(ret.peers, peering.UsePeer(remoteLocation))
		}
	}
	for _, remoteLocation := range bootupData.AccessNodes {
		p := peering.UsePeer(remoteLocation)
		if p != nil {
			ret.peers = append(ret.peers, p)
		}
	}

	ret.stateMgr = statemgr.New(ret, ret.log)
	if keyExists {
		ret.operator = consensus.NewOperator(ret, dkshare, ret.log)
	}
	go func() {
		for msg := range ret.chMsg {
			ret.dispatchMessage(msg)
		}
	}()

	return ret
}

// iAmInTheCommittee checks if netLocations makes sense
func iAmInTheCommittee(committeeNodes []string, n, index uint16) bool {
	if len(committeeNodes) != int(n) {
		return false
	}
	// check for duplicates
	return committeeNodes[index] == peering.MyNetworkId()
}
