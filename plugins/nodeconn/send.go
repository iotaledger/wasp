package nodeconn

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/packages/waspconn"
)

func SetSubscriptions(addrs []address.Address) {
	bconnMutex.Lock()
	defer bconnMutex.Unlock()

	subscriptions = addrs
	subscriptionsSent = false
}

func sendSubscriptionsIfNeeded() {
	bconnMutex.RLock()

	if subscriptionsSent || subscriptions == nil || bconn == nil {
		bconnMutex.RUnlock()
		return
	}
	bconnMutex.RUnlock()

	// switch to write lock
	bconnMutex.Lock()
	defer bconnMutex.Unlock()

	msg := &waspconn.WaspSendSubscribeMsg{
		Addresses:   subscriptions,
		PullBacklog: false,
	}
	var buf bytes.Buffer
	buf.WriteByte(waspconn.WaspSendSubscribeCode)
	buf.Write(msg.Encode())

	num := len(subscriptions)
	subscriptionsSent = true
	go func() {
		if err := SendDataToNode(buf.Bytes()); err != nil {
			log.Errorf("error while sending subscriptions: %v", err)
			bconnMutex.Lock()
			defer bconnMutex.Unlock()
			subscriptionsSent = false
		} else {
			log.Infof("sent subscriptions to node for %d addresses", num)
		}
	}()
}

func GetBalancesFromNode(addr *address.Address) {
	msg := waspconn.WaspSendGetBalancesMsg{
		Address: addr,
	}

	var buf bytes.Buffer
	buf.WriteByte(waspconn.WaspSendGetBalancesCode)
	buf.Write(msg.Encode())

	go func() {
		_ = SendDataToNode(buf.Bytes())
	}()
}
