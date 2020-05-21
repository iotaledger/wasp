package nodeconn

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
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

	data, err := waspconn.EncodeMsg(&waspconn.WaspToNodeSubscribeMsg{
		Addresses:   subscriptions,
		PullBacklog: false,
	})
	if err != nil {
		log.Errorf("sending subscriptions: %v", err)
		return
	}

	num := len(subscriptions)
	subscriptionsSent = true
	go func() {
		if err := SendDataToNode(data); err != nil {
			log.Errorf("sending subscriptions: %v", err)
			bconnMutex.Lock()
			defer bconnMutex.Unlock()
			subscriptionsSent = false
		} else {
			log.Infof("sent subscriptions to node for %d addresses", num)
		}
	}()
}

func RequestBalancesFromNode(addr *address.Address) error {
	data, err := waspconn.EncodeMsg(&waspconn.WaspToNodeGetBalancesMsg{
		Address: addr,
	})
	if err != nil {
		return err
	}
	if err := SendDataToNode(data); err != nil {
		return err
	}
	return nil
}

func RequestTransactionFromNode(txid *valuetransaction.ID) error {
	data, err := waspconn.EncodeMsg(&waspconn.WaspToNodeGetTransactionMsg{
		TxId: txid,
	})
	if err != nil {
		return err
	}
	if err := SendDataToNode(data); err != nil {
		return err
	}
	return nil
}

func PostTransactionToNode(tx *valuetransaction.Transaction) error {
	data, err := waspconn.EncodeMsg(&waspconn.WaspToNodeTransactionMsg{
		Tx: tx,
	})
	if err != nil {
		return err
	}
	if err = SendDataToNode(data); err != nil {
		return err
	}
	return nil
}
