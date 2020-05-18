package dispatcher

import (
	"bytes"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

func processMsgData(data []byte) {
	if len(data) == 0 {
		return
	}
	switch data[0] {
	case waspconn.WaspRecvTransactionCode:
		msg := &waspconn.WaspRecvTransactionMsg{}
		if err := msg.Read(bytes.NewReader(data[1:])); err != nil {
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

	case waspconn.WaspRecvBalancesCode:
		bals := &waspconn.WaspRecvBalancesMsg{}
		if err := bals.Read(bytes.NewReader(data[1:])); err != nil {
			log.Errorf("error parsing 'WaspRecvBalancesMsg' message: %v", err)
			return
		}
		dispatchBalances(bals.Address, bals.Balances)
	}
}
