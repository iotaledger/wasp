// implements triggers on arrived balances on addresses
// needs for implementation of sync calls
package dispatcher

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"sync"
	"time"
)

const ttlDuration = 5 * time.Second

type BalanceTrigger struct {
	fun func(bals map[valuetransaction.ID][]*balance.Balance)
	ttl time.Time
}

var (
	balanceConsumers      map[address.Address][]BalanceTrigger
	balanceConsumersMutex sync.Mutex
)

func init() {
	balanceConsumers = make(map[address.Address][]BalanceTrigger)
	go cleanupLoop()
}

func cleanupLoop() {
	for {
		time.Sleep(5 * time.Second)

		newMap := make(map[address.Address][]BalanceTrigger)

		balanceConsumersMutex.Lock()

		deleted := false
		nowis := time.Now()
		for addr, lst := range balanceConsumers {
			lstNew := make([]BalanceTrigger, 0)
			for _, c := range lst {
				if c.ttl.Before(nowis) {
					deleted = true
				} else {
					lstNew = append(lstNew, c)
				}
			}
			if len(lstNew) > 0 {
				newMap[addr] = lstNew
			}
		}
		if deleted {
			balanceConsumers = newMap
		}

		balanceConsumersMutex.Unlock()
	}
}

func AddBalanceTrigger(addr address.Address, fun func(map[valuetransaction.ID][]*balance.Balance)) {
	balanceConsumersMutex.Lock()
	defer balanceConsumersMutex.Unlock()

	lst, ok := balanceConsumers[addr]
	if !ok {
		lst = make([]BalanceTrigger, 0)
	}
	lst = append(lst, BalanceTrigger{
		fun: fun,
		ttl: time.Now().Add(ttlDuration),
	})
	balanceConsumers[addr] = lst
}

func triggerBalanceConsumers(addr address.Address, bals map[valuetransaction.ID][]*balance.Balance) {
	balanceConsumersMutex.Lock()
	defer balanceConsumersMutex.Unlock()

	lst, ok := balanceConsumers[addr]
	if !ok {
		return
	}
	for _, trig := range lst {
		trig.fun(bals)
	}
	delete(balanceConsumers, addr)
}
