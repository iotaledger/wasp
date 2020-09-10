package nodeconn

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/waspconn"
	"github.com/iotaledger/wasp/packages/util"
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

	addrs := make([]address.Address, 0, len(subscriptions))
	for a := range subscriptions {
		addrs = append(addrs, a)
	}
	subscriptionsSent = true
	go func() {
		astr := util.AddressesToStringsShort(addrs)
		data, err := waspconn.EncodeMsg(&waspconn.WaspToNodeSubscribeMsg{
			Addresses: addrs,
		})
		if err != nil {
			log.Errorf("sending subscriptions: %v. Addrs: %+v", err, astr)
			return
		}
		if err := SendDataToNode(data); err != nil {
			log.Errorf("sending subscriptions: %v. Addrs: %+v", err, astr)
			bconnMutex.Lock()
			defer bconnMutex.Unlock()
			subscriptionsSent = false
		} else {
			log.Infof("sent subscriptions to node for addresses %+v", astr)
		}
	}()
}
