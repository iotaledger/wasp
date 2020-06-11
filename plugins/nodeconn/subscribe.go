package nodeconn

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/packages/waspconn"
)

func Subscribe(addrs []address.Address) {
	bconnMutex.Lock()
	defer bconnMutex.Unlock()

	for _, a := range addrs {
		if _, ok := subscriptions[a]; !ok {
			subscriptionsSent = false
		}
		subscriptions[a] = struct{}{}
	}
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
