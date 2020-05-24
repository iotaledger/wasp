package nodeconn

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/packages/waspconn"
)

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
