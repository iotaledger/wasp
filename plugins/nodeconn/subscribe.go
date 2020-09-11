package nodeconn

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/waspconn"
)

func Subscribe(addr address.Address, color balance.Color) {
	bconnMutex.Lock()
	defer bconnMutex.Unlock()

	if _, ok := subscriptions[addr]; !ok {
		subscriptionsSent = false
	}
	subscriptions[addr] = color
}

func Unsubscribe(addr address.Address) {
	bconnMutex.Lock()
	defer bconnMutex.Unlock()

	delete(subscriptions, addr)
}

func sendSubscriptions(forceSend bool) {
	bconnMutex.Lock()
	defer bconnMutex.Unlock()

	if bconn == nil {
		return
	}
	if len(subscriptions) == 0 {
		return
	}
	if subscriptionsSent && !forceSend {
		return
	}

	addrsWithColors := make([]waspconn.AddressColor, 0, len(subscriptions))
	addrs := make([]string, 0)
	for a, c := range subscriptions {
		addrsWithColors = append(addrsWithColors, waspconn.AddressColor{
			Address: a,
			Color:   c,
		})
		addrs = append(addrs, a.String()[:6]+"..")
	}
	subscriptionsSent = true
	go func() {
		data, err := waspconn.EncodeMsg(&waspconn.WaspToNodeSubscribeMsg{
			AddressesWithColors: addrsWithColors,
		})
		if err != nil {
			log.Errorf("sending subscriptions: %v. Addrs: %+v", err, addrs)
			return
		}
		if err := SendDataToNode(data); err != nil {
			log.Errorf("sending subscriptions: %v. Addrs: %+v", err, addrs)
			bconnMutex.Lock()
			defer bconnMutex.Unlock()
			subscriptionsSent = false
		} else {
			log.Infof("sent subscriptions to node for addresses %+v", addrs)
		}
	}()
}
