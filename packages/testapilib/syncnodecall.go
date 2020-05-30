package testapilib

import (
	"errors"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/plugins/dispatcher"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"sync"
	"time"
)

const nodeRequestTimeout = 2 * time.Second

// synchronously requests balances from node
// this is non-standard communication with the node, too complicated
// will be deleted after testing interfaces become unnecessary
func GetBalancesFromNodeSync(addr address.Address) (map[valuetransaction.ID][]*balance.Balance, error) {
	var ret map[valuetransaction.ID][]*balance.Balance
	chCompleted := make(chan struct{})
	var closeMutex sync.Mutex
	var closed bool

	dispatcher.AddBalanceTrigger(addr, func(bals map[valuetransaction.ID][]*balance.Balance) {
		closeMutex.Lock()
		if !closed {
			ret = bals
			close(chCompleted)
			closed = true
		}
		closeMutex.Unlock()
	})

	if err := nodeconn.RequestOutputsFromNode(&addr); err != nil {
		close(chCompleted)
		closed = true
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
		return nil, errors.New("GetBalancesFromNodeSync: request to node timeout")
	}
}
