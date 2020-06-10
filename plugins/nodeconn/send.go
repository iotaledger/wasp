package nodeconn

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/plugins/peering"
	"time"
)

func SendWaspIdToNode() error {
	data, err := waspconn.EncodeMsg(&waspconn.WaspToNodeSetIdMsg{
		Waspid: peering.MyNetworkId(),
	})
	if err != nil {
		return err
	}
	if err := SendDataToNode(data); err != nil {
		return err
	}
	return nil
}

func RequestOutputsFromNode(addr *address.Address) error {
	data, err := waspconn.EncodeMsg(&waspconn.WaspToNodeGetOutputsMsg{
		Address: *addr,
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

func PostTransactionToNodeAsyncWithRetry(tx *valuetransaction.Transaction, retryEach, maxDuration time.Duration, log *logger.Logger) {
	deadline := time.Now().Add(maxDuration)
	go func() {
		for {
			if time.Now().After(deadline) {
				log.Warn("PostTransactionToNodeAsyncWithRetry: cancelled sending transaction to node txid = %s", tx.ID().String())
				return
			}
			err := PostTransactionToNode(tx)
			if err == nil {
				return
			}
			log.Warn("PostTransactionToNodeAsyncWithRetry: txid %s err = %v", tx.ID().String(), err)
			time.Sleep(retryEach)
		}
	}()
}
