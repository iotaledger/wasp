package nodeconn

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/waspconn"
)

func Subscribe(addr address.Address) {
	bconnMutex.Lock()
	defer bconnMutex.Unlock()

	if _, ok := subscriptions[addr]; !ok {
		subscriptionsSent = false
	}
	subscriptions[addr] = struct{}{}
}

func Unsubscribe(addr address.Address) {
	bconnMutex.Lock()
	defer bconnMutex.Unlock()

	delete(subscriptions, addr)
}

func sendSubscriptionsIfNeeded() {
	bconnMutex.RLock()
	if subscriptionsSent || bconn == nil {
		bconnMutex.RUnlock()
		return
	}
	// switch to write lock
	bconnMutex.RUnlock()
	bconnMutex.Lock()
	defer bconnMutex.Unlock()

	if len(subscriptions) == 0 {
		return
	}

	addrs := make([]address.Address, 0, len(subscriptions))
	for a := range subscriptions {
		addrs = append(addrs, a)
	}
	subscriptionsSent = true

	go func() {
		data, err := waspconn.EncodeMsg(&waspconn.WaspToNodeSubscribeMsg{
			Addresses: addrs,
		})
		if err != nil {
			log.Errorf("sending subscriptions: %v", err)
			return
		}
		if err := SendDataToNode(data); err != nil {
			log.Errorf("sending subscriptions: %v", err)
			bconnMutex.Lock()
			defer bconnMutex.Unlock()
			subscriptionsSent = false
		} else {
			log.Infof("sent subscriptions to node for %d addresses", len(addrs))
		}
	}()
}
