package dispatcher

import (
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"time"
)

func processNodeMsgData(data []byte) {
	//log.Debugf("processNodeMsgData")

	msg, err := waspconn.DecodeMsg(data, true)
	if err != nil {
		log.Errorf("wrong message from node: %v", err)
		return
	}
	switch msgt := msg.(type) {
	case *waspconn.WaspPingMsg:
		roundtrip := time.Since(time.Unix(0, msgt.Timestamp))
		log.Infof("PING %d response from node. Roundtrip %v", msgt.Id, roundtrip)

	case *waspconn.WaspFromNodeTransactionMsg:
		//log.Debugf("*waspconn.WaspFromNodeTransactionMsg")

		tx, err := sctransaction.ParseValueTransaction(msgt.Tx)
		if err != nil {
			log.Debugw("!!!! after parsing", "txid", msgt.Tx.ID().String(), "err", err)
			// not a SC transaction. Ignore
			return
		}
		dispatchState(tx)
		dispatchRequests(tx)

	case *waspconn.WaspFromNodeBalancesMsg:
		dispatchBalances(msgt.Address, msgt.Balances)
	}
}
