package dispatcher

import (
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/plugins/nodeconn"
)

func processMsgData(data []byte) {
	if len(data) == 0 {
		return
	}
	switch data[0] {
	case nodeconn.WaspRecvTransactionCode:
		msg := &nodeconn.WaspRecvTransactionMsg{}
		if err := msg.Decode(data[1:]); err != nil {
			log.Errorf("error parsing 'WaspRecvTransactionMsg' message: %v", err)
			return
		}
		tx, err := sctransaction.ParseValueTransaction(msg.Tx)
		if err != nil {
			// not a SC transaction. Ignore
			return
		}
		dispatchState(tx)
		dispatchRequests(tx)

	case nodeconn.WaspRecvBalancesCode:
		bals := &nodeconn.WaspRecvBalancesMsg{}
		if err := bals.Decode(data[1:]); err != nil {
			log.Errorf("error parsing 'WaspRecvBalancesMsg' message: %v", err)
			return
		}
		dispatchBalances(bals.Address, bals.Balances)
	}
}
