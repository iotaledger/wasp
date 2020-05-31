package sctransaction

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/util"
)

// validates state block and requests and returns if it origin state (if not error)
// address is address of the SC
func (tx *Transaction) ValidateBlocks(addr *address.Address) (bool, error) {
	isOrigin, err := tx.validateStateBlock(addr)
	if err != nil {
		return false, err
	}
	if isOrigin {
		return true, nil
	}
	return false, tx.validateRequests()
}

// check correctness of the SC token
func (tx *Transaction) validateStateBlock(addr *address.Address) (bool, error) {
	stateBlock, hasState := tx.State()
	if !hasState {
		return false, nil
	}

	color := stateBlock.Color()
	mayBeOrigin := false
	if color == balance.ColorNew {
		// may be origin
		color = (balance.Color)(tx.ID())
		mayBeOrigin = true
	}

	if mayBeOrigin && len(tx.Requests()) > 0 {
		return false, fmt.Errorf("origin transaction can't contain requests")
	}

	balances, hasAddress := tx.OutputBalancesByAddress(addr)
	if !hasAddress {
		// invalid state
		return false, fmt.Errorf("invalid state block: SC state output not found")
	}
	outBalance := util.BalanceOfColor(balances, color)
	if outBalance == 0 && mayBeOrigin {
		outBalance = util.BalanceOfColor(balances, balance.ColorNew)
	}
	if outBalance != 1 {
		// supply of the SC token must be exactly 1
		return false, fmt.Errorf("non-existent or wrong output with SC token")
	}
	if mayBeOrigin && stateBlock.StateIndex() != 0 {
		return false, fmt.Errorf("origin transaction must have state index 0")
	}
	return mayBeOrigin, nil
}

// check correctness of the request tokens
func (tx *Transaction) validateRequests() error {
	sumReqTokens := int64(0)
	tx.Outputs().ForEach(func(addr address.Address, bals []*balance.Balance) bool {
		sumReqTokens += util.BalanceOfColor(bals, balance.ColorNew)
		return true
	})
	if sumReqTokens != int64(len(tx.Requests())) {
		return fmt.Errorf("wrong number of request tokens")
	}
	return nil
}
