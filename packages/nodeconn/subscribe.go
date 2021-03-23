package nodeconn

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/waspconn"
)

func (n *NodeConn) Subscribe(addr *ledgerstate.AliasAddress) {
	n.chSubscribe <- addr
}

func (n *NodeConn) Unsubscribe(addr *ledgerstate.AliasAddress) {
	n.chUnsubscribe <- addr
}

func (n *NodeConn) subscriptionsLoop() {
	subscriptions := make(map[[ledgerstate.AddressLength]byte]*ledgerstate.AliasAddress)

	ticker1m := time.NewTicker(time.Minute)
	defer ticker1m.Stop()

	for {
		select {
		case <-n.shutdown:
			return
		case addr := <-n.chSubscribe:
			addrBytes := addr.Array()
			if _, ok := subscriptions[addrBytes]; !ok {
				// new address is subscribed
				n.log.Infof("subscribed to address %s", addr.Base58())
				subscriptions[addrBytes] = addr
				n.sendSubscriptions(subscriptions)
			}
		case addr := <-n.chUnsubscribe:
			delete(subscriptions, addr.Array())
		case <-ticker1m.C:
			// send subscriptions once every minute
			n.sendSubscriptions(subscriptions)
		}
	}
}

func (n *NodeConn) sendSubscriptions(subscriptions map[[ledgerstate.AddressLength]byte]*ledgerstate.AliasAddress) {
	if len(subscriptions) == 0 {
		return
	}

	addrs := make([]*ledgerstate.AliasAddress, len(subscriptions))
	{
		i := 0
		for _, addr := range subscriptions {
			addrs[i] = addr
			i++
		}
	}

	if err := n.sendToNode(&waspconn.WaspToNodeUpdateSubscriptionsMsg{
		ChainAddresses: addrs,
	}); err != nil {
		n.log.Errorf("sending subscriptions to node: %v", err)
	} else {
		n.log.Infof("sent subscriptions to node for addresses %+v", addrs)
	}
}
