package nodeconn

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/wasp/packages/parameters"
)

func (n *NodeConn) Subscribe(addr ledgerstate.Address, color ledgerstate.Color) {
	n.bconn.Lock()
	defer n.bconn.Unlock()

	if _, ok := n.subscriptions[addr]; !ok {
		n.subscriptionsSent = false
	}
	n.subscriptions[addr] = color
}

func (n *NodeConn) Unsubscribe(addr ledgerstate.Address) {
	n.bconn.Lock()
	defer n.bconn.Unlock()

	delete(n.subscriptions, addr)
}

func (n *NodeConn) sendSubscriptions(forceSend bool) {
	n.bconn.Lock()
	defer n.bconn.Unlock()

	if n.bconn.BufferedConnection == nil {
		return
	}
	if len(n.subscriptions) == 0 {
		return
	}
	if n.subscriptionsSent && !forceSend {
		return
	}

	addrsWithColors := make([]waspconn.AddressColor, 0, len(n.subscriptions))
	addrs := make([]string, 0)
	for a, c := range n.subscriptions {
		addrsWithColors = append(addrsWithColors, waspconn.AddressColor{
			Address: a,
			Color:   c,
		})
		addrs = append(addrs, a.String()[:6]+"..")
	}
	n.subscriptionsSent = true
	go func() {
		data := waspconn.EncodeMsg(&waspconn.WaspToNodeSubscribeMsg{
			AddressesWithColors: addrsWithColors,
		})
		if err := n.SendDataToNode(data); err != nil {
			n.log.Errorf("sending subscriptions to %s: %v. Addrs: %+v",
				parameters.GetString(parameters.NodeAddress), err, addrs)
			n.bconn.Lock()
			defer n.bconn.Unlock()
			n.subscriptionsSent = false
		} else {
			n.log.Infof("sent subscriptions to node %s for addresses %+v",
				parameters.GetString(parameters.NodeAddress), addrs)
		}
	}()
}
