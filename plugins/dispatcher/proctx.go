// this package defines main entry how value transactions are entering the qnode
package dispatcher

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

func dispatchState(tx *sctransaction.Transaction) {
	cmt, ok := validateState(tx)
	if !ok {
		return
	}
	cmt.ReceiveMessage(committee.StateTransactionMsg{Transaction: tx})
}

func dispatchRequests(tx *sctransaction.Transaction) {
	for i, reqBlk := range tx.Requests() {
		if cmt := committeeByAddress(reqBlk.Address()); cmt != nil {
			cmt.ReceiveMessage(&committee.RequestMsg{
				Transaction: tx,
				Index:       uint16(i),
			})
		}
	}
}

func dispatchBalances(address *address.Address, bals map[valuetransaction.ID][]*balance.Balance) {
	// pass to the committee by address
	if cmt := committeeByAddress(address); cmt != nil {
		cmt.ReceiveMessage(committee.BalancesMsg{Balances: bals})
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
	if *color == balance.ColorNew {
		// may be origin
		*color = (balance.Color)(tx.ID())
		mayBeOrigin = true
	}
	cmt := committeeByColor(color)
	if cmt == nil {
		return nil, false
	}
	// get address of the SC and check if the transaction contains and output to that address with
	// the respective color and value 1
	addr := cmt.Address()

	balances, hasAddress := tx.OutputBalancesByAddress(addr)
	if !hasAddress {
		// invalid state
		log.Errorw("invalid state block: SC state output not found",
			"addr", addr.String(),
			"tx", tx.ID().String(),
		)
		return nil, false
	}
	outBalance := sctransaction.SumBalancesOfColor(balances, color)
	if outBalance == 0 && mayBeOrigin {
		outBalance = sctransaction.SumBalancesOfColor(balances, (*balance.Color)(&balance.ColorNew))
	}
	if outBalance != 1 {
		// supply of the SC token must be exactly 1
		log.Errorf("non-existent or wrong output with SC token in sc tx %s", tx.ID().String())
		return nil, false
	}
	return cmt, true
}
