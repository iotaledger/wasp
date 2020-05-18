package nodeconn

import (
	"bytes"
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

	msg := &waspconn.WaspSendSubscribeMsg{
		Addresses:   subscriptions,
		PullBacklog: false,
	}
	var buf bytes.Buffer
	if err := buf.WriteByte(waspconn.WaspSendSubscribeCode); err != nil {
		log.Error(err)
		return
	}
	if err := msg.Write(&buf); err != nil {
		log.Error(err)
		return
	}

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

func RequestBalancesFromNode(addr *address.Address) {
	msg := waspconn.WaspSendGetBalancesMsg{
		Address: addr,
	}

	var buf bytes.Buffer
	if err := buf.WriteByte(waspconn.WaspSendGetBalancesCode); err != nil {
		log.Error(err)
		return
	}
	if err := msg.Write(&buf); err != nil {
		log.Error(err)
		return
	}

	go func() {
		if err := SendDataToNode(buf.Bytes()); err != nil {
			log.Warnf("failed to send 'WaspSendGetBalancesMsg' to the node: %v", err)
		}
	}()
}

func RequestTransactionFromNode(txid *valuetransaction.ID) {
	msg := waspconn.WaspSendGetTransactionMsg{
		TxId: txid,
	}
	var buf bytes.Buffer
	if err := buf.WriteByte(waspconn.WaspSendGetTransactionCode); err != nil {
		log.Error(err)
		return
	}
	if err := msg.Write(&buf); err != nil {
		log.Error(err)
		return
	}

	go func() {
		if err := SendDataToNode(buf.Bytes()); err != nil {
			log.Warnf("failed to send 'WaspSendGetTransactionMsg' to the node: %v", err)
		}
	}()
}
