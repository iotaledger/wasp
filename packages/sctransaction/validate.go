package sctransaction

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/packages/waspconn"
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

// checks if transaction value part:
// - contains only inputs from the address
// - contains all inputs
// - correctness of colored balances
func (tx *Transaction) ValidateConsumptionOfInputs(addr *address.Address, inputBalances map[valuetransaction.ID][]*balance.Balance) error {
	if err := waspconn.ValidateBalances(inputBalances); err != nil {
		return err
	}
	var err error
	var first bool
	var addrTmp address.Address
	var totalInputs2 int64

	inputBalancesByColor, totalInputs1 := util.BalancesByColor(inputBalances)

	tx.Inputs().ForEach(func(outputId valuetransaction.OutputID) bool {
		if !first {
			addrTmp = outputId.Address()
			first = true
		}
		if addrTmp != outputId.Address() {
			err = errors.New("only 1 address is allowed in inputs")
			return false
		}
		bals, ok := inputBalances[outputId.TransactionID()]
		if !ok {
			err = errors.New("unexpected txid in inputs")
			return false
		}
		totalInputs2 += util.BalancesSumTotal(bals)
		return true
	})

	if totalInputs1 != totalInputs2 {
		return errors.New("not all provided inputs are consumed")
	}

	outputBalancesByColor := make(map[balance.Color]int64)
	tx.Outputs().ForEach(func(_ address.Address, bals []*balance.Balance) bool {
		for _, b := range bals {
			if s, ok := outputBalancesByColor[b.Color()]; !ok {
				outputBalancesByColor[b.Color()] = b.Value()
			} else {
				outputBalancesByColor[b.Color()] = s + b.Value()
			}
		}
		return true
	})

	for col, inb := range inputBalancesByColor {
		if !(col != balance.ColorNew) {
			return errors.New("assertion failed: col != balance.ColorNew")
		}
		if col == balance.ColorIOTA {
			continue
		}
		outb, ok := outputBalancesByColor[col]
		if !ok {
			continue
		}
		if outb > inb {
			// colored supply can't be inflated
			return errors.New("colored supply can't be inflated")
		}
	}
	return nil
}
