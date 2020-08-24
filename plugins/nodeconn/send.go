package nodeconn

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/waspconn"
	"github.com/iotaledger/wasp/plugins/peering"
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

func RequestConfirmedTransactionFromNode(txid *valuetransaction.ID) error {
	data, err := waspconn.EncodeMsg(&waspconn.WaspToNodeGetConfirmedTransactionMsg{
		TxId: *txid,
	})
	if err != nil {
		return err
	}
	if err := SendDataToNode(data); err != nil {
		return err
	}
	return nil
}

func PostTransactionToNode(tx *valuetransaction.Transaction, fromSc *address.Address, fromLeader uint16) error {
	data, err := waspconn.EncodeMsg(&waspconn.WaspToNodeTransactionMsg{
		Tx:        tx,
		SCAddress: *fromSc, // just for tracing
		Leader:    fromLeader,
	})
	if err != nil {
		return err
	}
	if err = SendDataToNode(data); err != nil {
		return err
	}
	return nil
}
