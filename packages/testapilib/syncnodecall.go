package testapilib

import (
	"errors"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/plugins/dispatcher"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"sync"
	"time"
)

const nodeRequestTimeout = 2 * time.Second

// synchronously requests balances from node
func GetBalancesFromNodeSync(addr address.Address) (map[valuetransaction.ID][]*balance.Balance, error) {
	var ret map[valuetransaction.ID][]*balance.Balance
	chCompleted := make(chan struct{})
	var closeMutex sync.Mutex
	var closed bool

	closure := events.NewClosure(func(addrIntern address.Address, bals map[valuetransaction.ID][]*balance.Balance) {
		if addr == addrIntern {
			closeMutex.Lock()
			if !closed {
				ret = bals
				close(chCompleted)
				closed = true
			}
			closeMutex.Unlock()
		}
	})
	dispatcher.Events.BalancesArrivedFromNode.Attach(closure)
	defer dispatcher.Events.BalancesArrivedFromNode.Detach(closure)

	if err := nodeconn.RequestBalancesFromNode(&addr); err != nil {
		close(chCompleted)
		return nil, err
	}
	select {
	case <-chCompleted:
		return ret, nil

	case <-time.After(nodeRequestTimeout):
		closeMutex.Lock()
		if !closed {
			close(chCompleted)
			closed = true
		}
		closeMutex.Unlock()
		return nil, errors.New("request to node timeout")
	}
}
