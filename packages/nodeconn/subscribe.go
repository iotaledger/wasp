package nodeconn

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/waspconn"
)

func (n *NodeConn) Subscribe(addr *ledgerstate.AliasAddress) {
	n.mu.Lock()
	defer n.mu.Unlock()

	addrBytes := addr.Array()
	if _, ok := n.subscriptions[addrBytes]; !ok {
		n.subscriptionsSent = false
	}
	n.subscriptions[addrBytes] = addr
}

func (n *NodeConn) Unsubscribe(addr *ledgerstate.AliasAddress) {
	n.mu.Lock()
	defer n.mu.Unlock()

	delete(n.subscriptions, addr.Array())
}

func (n *NodeConn) sendSubscriptions(forceSend bool, goshimerNodeAddress string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.bconn == nil {
		return
	}
	if len(n.subscriptions) == 0 {
		return
	}
	if n.subscriptionsSent && !forceSend {
		return
	}

	addrs := make([]*ledgerstate.AliasAddress, len(n.subscriptions))
	{
		i := 0
		for _, addr := range n.subscriptions {
			addrs[i] = addr
			i++
		}
	}
	n.subscriptionsSent = true
	go func() {
		if err := n.sendToNode(&waspconn.WaspToNodeSubscribeMsg{
			ChainAddresses: addrs,
		}); err != nil {
			n.log.Errorf("sending subscriptions to %s: %v. Addrs: %+v", goshimerNodeAddress, err, addrs)
			n.mu.Lock()
			defer n.mu.Unlock()
			n.subscriptionsSent = false
		} else {
			n.log.Infof("sent subscriptions to node %s for addresses %+v", goshimerNodeAddress, addrs)
		}
	}()
}
