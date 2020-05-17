package commiteeimpl

import (
	"errors"
	"fmt"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/statemgr"
	"github.com/iotaledger/wasp/plugins/peering"
	"go.uber.org/atomic"
	"time"
)

const (
	useTimer        = true
	timerTickPeriod = 20 * time.Millisecond
)

type committeeObj struct {
	isOpenQueue atomic.Bool
	peers       []*peering.Peer
	ownIndex    uint16
	scdata      *registry.SCMetaData
	chMsg       chan interface{}
	stateMgr    committee.StateManager
	operator    committee.Operator
	log         *logger.Logger
}

func newCommitteeObj(scdata *registry.SCMetaData, log *logger.Logger) (committee.Committee, error) {
	dkshare, keyExists, err := registry.GetDKShare(&scdata.Address)
	if err != nil {
		return nil, err
	}
	if !keyExists {
		return nil, fmt.Errorf("unkniwn key. sc addr = %s", scdata.Address.String())
	}
	err = fmt.Errorf("sc data inconsstent with key parameteres for sc addr %s", scdata.Address.String())
	if scdata.Address != *dkshare.Address {
		return nil, err
	}
	if err := checkNetworkLocations(scdata.NodeLocations, dkshare.N, dkshare.Index); err != nil {
		return nil, err
	}

	ret := &committeeObj{
		chMsg:    make(chan interface{}),
		scdata:   scdata,
		peers:    make([]*peering.Peer, len(scdata.NodeLocations)),
		ownIndex: dkshare.Index,
		log:      log.Named("cmt"),
	}
	myLocation := scdata.NodeLocations[dkshare.Index]
	for i, remoteLocation := range scdata.NodeLocations {
		if i != int(dkshare.Index) {
			ret.peers[i] = peering.UsePeer(remoteLocation, myLocation)
		}
	}

	ret.stateMgr = statemgr.New(ret, ret.log)
	//ret.operator = consensus.NewOperator(ret, dkshare)

	ret.OpenQueue() // TODO only for testing

	go func() {
		for msg := range ret.chMsg {
			ret.dispatchMessage(msg)
		}
	}()

	if useTimer {
		go func() {
			tick := 0
			for {
				time.Sleep(timerTickPeriod)
				ret.ReceiveMessage(committee.TimerTick(tick))
			}
		}()
	}

	return ret, nil
}

// checkNetworkLocations checks if netLocations makes sense
func checkNetworkLocations(netLocations []string, n, index uint16) error {
	if len(netLocations) != int(n) {
		return fmt.Errorf("wrong number of network locations")
	}
	// check for duplicates
	for i := range netLocations {
		for j := i + 1; j < len(netLocations); j++ {
			if netLocations[i] == netLocations[j] {
				return errors.New("duplicate network locations in the list")
			}
		}
	}
	return peering.CheckMyNetworkID(netLocations[index])
}
