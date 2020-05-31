// this package defines main entry how value transactions are entering the qnode
package dispatcher

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/committees"
)

func dispatchState(tx *sctransaction.Transaction) {
	cmt, ok := validateState(tx)
	if !ok {
		return
	}
	log.Debugw("dispatchState", "txid", tx.ID().String())

	cmt.ReceiveMessage(committee.StateTransactionMsg{Transaction: tx})
}

func dispatchBalances(addr address.Address, bals map[valuetransaction.ID][]*balance.Balance) {
	// pass to the committee by address
	if cmt := committees.CommitteeByAddress(addr); cmt != nil {
		cmt.ReceiveMessage(committee.BalancesMsg{Balances: bals})
	}
	triggerBalanceConsumers(addr, bals)
}

func dispatchAddressUpdate(addr address.Address, balances map[valuetransaction.ID][]*balance.Balance, tx *sctransaction.Transaction) {
	cmt := committees.CommitteeByAddress(addr)
	if cmt == nil {
		// wrong addressee
		return
	}
	if err := validateTransactionWithBalances(tx, balances); err != nil {
		log.Warnf("transaction %s ignored: %v", tx.ID().String(), err)
		return
	}

	var stateTxMsg committee.StateTransactionMsg
	requestMsgs := make([]committee.RequestMsg, 0, len(tx.Requests()))

	stateBlock, ok := tx.State()
	if ok && stateBlock.Color() == cmt.Color() {
		stateTxMsg = committee.StateTransactionMsg{tx}
	}

	for i, reqBlk := range tx.Requests() {
		if reqBlk.Address() == addr {
			requestMsgs = append(requestMsgs, committee.RequestMsg{
				Transaction: tx,
				Index:       uint16(i),
			})
		}
	}

	if stateTxMsg.Transaction != nil || len(requestMsgs) > 0 {
		cmt.ReceiveMessage(committee.BalancesMsg{Balances: balances})
	}
	// send messages
	if stateTxMsg.Transaction != nil {
		cmt.ReceiveMessage(stateTxMsg)
	}
	for _, reqMsg := range requestMsgs {
		cmt.ReceiveMessage(reqMsg)
	}
}

// validates and returns if it has state block, is it origin state or error
func validateState(tx *sctransaction.Transaction) (committee.Committee, bool) {
	stateBlock, hasState := tx.State()
	if !hasState {
		return nil, false
	}
	color := stateBlock.Color()
	mayBeOrigin := false
	if color == balance.ColorNew {
		// may be origin
		color = (balance.Color)(tx.ID())
		mayBeOrigin = true
	}
	cmt := committees.CommitteeByColor(color)
	if cmt == nil {
		return nil, false
	}
	// get address of the SC and check if the transaction contains and output to that address with
	// the respective color and value 1
	addr := cmt.Address()

	balances, hasAddress := tx.OutputBalancesByAddress(&addr)
	if !hasAddress {
		// invalid state
		log.Errorw("invalid state block: SC state output not found",
			"addr", addr.String(),
			"tx", tx.ID().String(),
		)
		return nil, false
	}
	outBalance := util.BalanceOfColor(balances, color)
	if outBalance == 0 && mayBeOrigin {
		outBalance = util.BalanceOfColor(balances, balance.ColorNew)
	}
	if outBalance != 1 {
		// supply of the SC token must be exactly 1
		log.Errorf("non-existent or wrong output with SC token in sc tx %s", tx.ID().String())
		return nil, false
	}
	return cmt, true
}

func validateTransactionWithBalances(tx *sctransaction.Transaction, balances map[valuetransaction.ID][]*balance.Balance) error {
	if _, ok := validateState(tx); !ok {
		return fmt.Errorf("invalid state block")
	}
	sumReqTokens := int64(0)
	tx.Outputs().ForEach(func(addr address.Address, bals []*balance.Balance) bool {
		sumReqTokens += util.BalanceOfColor(bals, balance.ColorNew)
		return true
	})
	if sumReqTokens != int64(len(tx.Requests())) {
		return fmt.Errorf("wrong number of request tokens in transaction")
	}
	// check if balances contains all outputs wrt to requests
	bals, ok := balances[tx.ID()]
	if !ok {
		return fmt.Errorf("can't find request tokens")
	}
	if util.BalanceOfColor(bals, (balance.Color)(tx.ID())) != sumReqTokens {
		return fmt.Errorf("wrong number of request tokens in balances")
	}
	return nil
}
