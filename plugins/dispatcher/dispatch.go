// this package defines main entry how value transactions are entering the qnode
package dispatcher

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/plugins/committees"
)

func dispatchState(tx *sctransaction.Transaction) {
	cmt := getCommitteeByState(tx)
	if cmt == nil {
		return
	}
	log.Debugw("dispatchState",
		"txid", tx.ID().String(),
		"addr", cmt.Address().String(),
	)
	_, err := tx.ValidateBlocks(cmt.Address())
	if err != nil {
		log.Errorf("invalid transaction %s ignored: %v", tx.ID().String(), err)
		return
	}

	cmt.ReceiveMessage(committee.StateTransactionMsg{
		Transaction: tx,
	})
}

func dispatchBalances(addr address.Address, bals map[valuetransaction.ID][]*balance.Balance) {
	// pass to the committee by address
	if cmt := committees.CommitteeByAddress(addr); cmt != nil {
		cmt.ReceiveMessage(committee.BalancesMsg{Balances: bals})
	}
}

func dispatchAddressUpdate(addr address.Address, balances map[valuetransaction.ID][]*balance.Balance, tx *sctransaction.Transaction) {
	log.Debugw("dispatchAddressUpdate", "addr", addr.String())

	cmt := committees.CommitteeByAddress(addr)
	if cmt == nil {
		log.Debugw("committee not found", "addr", addr.String())
		// wrong addressee
		return
	}
	if _, ok := balances[tx.ID()]; !ok {
		log.Errorf("transaction %s is not among provided outputs. Ignored", tx.ID().String())
		return
	}
	if _, err := tx.ValidateBlocks(&addr); err != nil {
		log.Warnf("invalid transaction %s ignored: %v", tx.ID().String(), err)
		return
	}

	log.Debugf("received with balances: %s", tx.String())

	var stateTxMsg committee.StateTransactionMsg
	requestMsgs := make([]committee.RequestMsg, 0, len(tx.Requests()))

	if cmtState := getCommitteeByState(tx); cmtState != nil && *cmtState.Address() == addr {
		stateTxMsg = committee.StateTransactionMsg{
			Transaction: tx,
		}
	}

	for i, reqBlk := range tx.Requests() {
		if reqBlk.Address() == addr {
			requestMsgs = append(requestMsgs, committee.RequestMsg{
				Transaction: tx,
				Index:       uint16(i),
			})
		}
	}

	// balances must be refreshed before requests to ensure corresponding outputs with request tokens
	if stateTxMsg.Transaction != nil || len(requestMsgs) > 0 {
		cmt.ReceiveMessage(committee.BalancesMsg{Balances: balances})
	}

	if stateTxMsg.Transaction != nil {
		cmt.ReceiveMessage(stateTxMsg)

		sh := stateTxMsg.Transaction.MustState().VariableStateHash()
		log.Debugw("state tx dispatched",
			"txid", stateTxMsg.Transaction.ID().String(),
			"state index", stateTxMsg.Transaction.MustState().StateIndex(),
			"state hash", sh.String(),
		)
	}
	for _, reqMsg := range requestMsgs {
		cmt.ReceiveMessage(reqMsg)

		reqid := sctransaction.NewRequestId(tx.ID(), reqMsg.Index)
		log.Debugw("request dispatched",
			"addr", addr.String(),
			"req", reqid.String(),
		)
	}
}

func getCommitteeByState(tx *sctransaction.Transaction) committee.Committee {
	//log.Debugw("getCommitteeByState", "txid", tx.ID().String())

	stateAddr, ok, err := tx.StateAddress()
	if err != nil {
		log.Errorf("getCommitteeByState: StateAddress returned for txid = %s: %v", tx.ID().String(), err)
	}
	if !ok || err != nil {
		return nil
	}
	log.Debugw("getCommitteeByState",
		"txid", tx.ID().String(),
		"addr", stateAddr.String(),
	)

	return committees.CommitteeByAddress(stateAddr)
}
