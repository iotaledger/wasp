package testchain

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
)

type MockedLedgers struct {
	ledgers      map[string]*MockedLedger
	stateAddress iotago.Address
	milestones   *events.Event
	log          *logger.Logger
	mutex        sync.Mutex
}

func NewMockedLedgers(stateAddress iotago.Address, log *logger.Logger) *MockedLedgers {
	result := &MockedLedgers{
		ledgers:      make(map[string]*MockedLedger),
		stateAddress: stateAddress,
		milestones: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(chain.NodeConnectionMilestonesHandlerFun)(params[0].(*nodeclient.MilestonePointer))
		}),
		log: log.Named("mls"),
	}
	go result.pushMilestonesLoop()
	result.log.Debugf("Mocked ledgers created")
	return result
}

func (mlT *MockedLedgers) GetLedger(chainID *iscp.ChainID) *MockedLedger {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	result, ok := mlT.ledgers[chainID.Key()]
	if !ok {
		mlT.log.Debugf("New ledger for chain address %s created", chainID)
		result = NewMockedLedger(chainID, mlT.stateAddress, mlT.log)
		mlT.ledgers[chainID.Key()] = result
	}
	return result
}

func (mlT *MockedLedgers) AttachMilestones(handler chain.NodeConnectionMilestonesHandlerFun) *events.Closure {
	closure := events.NewClosure(handler)
	mlT.milestones.Attach(closure)
	return closure
}

func (mlT *MockedLedgers) DetachMilestones(attachID *events.Closure) {
	mlT.milestones.Detach(attachID)
}

func (mlT *MockedLedgers) pushMilestonesLoop() {
	milestone := uint32(0)
	for {
		if milestone%10 == 0 {
			mlT.log.Debugf("Milestone %v reached", milestone)
		}
		time.Sleep(100 * time.Millisecond)
		mlT.milestones.Trigger(&nodeclient.MilestonePointer{
			Index:     milestone,
			Timestamp: uint64(time.Now().UnixNano()),
		})
		milestone++
	}
}

func (mlT *MockedLedgers) SetPushOutputToNodesNeeded(flag bool) {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	for _, ledger := range mlT.ledgers {
		ledger.SetPushOutputToNodesNeeded(flag)
	}
}
