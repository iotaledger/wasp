package testchain

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/core/events"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/inx-app/nodebridge"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

type MockedLedgers struct {
	ledgers              map[string]*MockedLedger
	milestones           *events.Event
	pushMilestonesNeeded bool
	log                  *logger.Logger
	mutex                sync.Mutex
}

func NewMockedLedgers(log *logger.Logger) *MockedLedgers {
	result := &MockedLedgers{
		ledgers: make(map[string]*MockedLedger),
		milestones: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(*nodebridge.Milestone))(params[0].(*nodebridge.Milestone))
		}),
		log: log.Named("mls"),
	}
	result.SetPushMilestonesToNodesNeeded(true)
	go result.pushMilestonesLoop()
	result.log.Debugf("Mocked ledgers created")
	return result
}

func (mlT *MockedLedgers) InitLedger(stateAddress iotago.Address) *isc.ChainID {
	ledger, chainID := NewMockedLedger(stateAddress, mlT.log)
	mlT.ledgers[chainID.Key()] = ledger
	mlT.log.Debugf("New ledger for chain address %s ID %s created", stateAddress, chainID)
	return chainID
}

func (mlT *MockedLedgers) GetLedger(chainID *isc.ChainID) *MockedLedger {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	result, ok := mlT.ledgers[chainID.Key()]
	if !ok {
		mlT.log.Errorf("Ledger for chain ID %s not found", chainID)
	}
	return result
}

func (mlT *MockedLedgers) AttachMilestones(handler func(*nodebridge.Milestone)) *events.Closure {
	closure := events.NewClosure(handler)
	mlT.milestones.Hook(closure)
	return closure
}

func (mlT *MockedLedgers) DetachMilestones(attachID *events.Closure) {
	mlT.milestones.Detach(attachID)
}

func (mlT *MockedLedgers) pushMilestonesLoop() {
	milestone := uint32(0)
	for {
		if milestone%10 == 0 {
			mlT.log.Debugf("Milestone %v reached, will push to nodes: %v", milestone, mlT.pushMilestonesNeeded)
		}
		if mlT.pushMilestonesNeeded {
			mlT.milestones.Trigger(&nodebridge.Milestone{
				MilestoneID: [32]byte{},
				Milestone: &iotago.Milestone{
					Index:     milestone,
					Timestamp: uint32(time.Now().Unix()),
				},
			})
		}
		time.Sleep(100 * time.Millisecond)
		milestone++
	}
}

func (mlT *MockedLedgers) SetPushMilestonesToNodesNeeded(flag bool) {
	mlT.pushMilestonesNeeded = flag
}

func (mlT *MockedLedgers) SetPushOutputToNodesNeeded(flag bool) {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	for _, ledger := range mlT.ledgers {
		ledger.SetPushOutputToNodesNeeded(flag)
	}
}
