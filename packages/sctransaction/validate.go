package sctransaction

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/txutil"
)

// validates state block and requests and returns if it origin state (if not error)
// address is address of the SC
func (tx *Transaction) ValidateBlocks(addr *address.Address) (bool, error) {
	isOrigin, err := tx.validateStateBlock(addr)
	if err != nil {
		return false, err
	}
	return isOrigin, tx.validateRequests(isOrigin)
}

// check correctness of the SC token
func (tx *Transaction) validateStateBlock(addr *address.Address) (bool, error) {
	stateBlock, hasState := tx.State()
	if !hasState {
		return false, nil
	}

	color := stateBlock.Color()
	mayBeOrigin := color == balance.ColorNew

	balances, hasAddress := tx.OutputBalancesByAddress(addr)
	if !hasAddress {
		// invalid state
		return false, fmt.Errorf("invalid state block: SC state output not found")
	}
	outBalance := txutil.BalanceOfColor(balances, color)
	expectedOutputBalance := int64(1)
	if mayBeOrigin {
		expectedOutputBalance += int64(len(tx.Requests()))
	}
	// expected 1 SC token if tx is not origin or 1 + number of request if color is MintColor (origin)
	if outBalance != expectedOutputBalance {
		return false, fmt.Errorf("non-existent or wrong output with SC token")
	}
	if mayBeOrigin && stateBlock.StateIndex() != 0 {
		return false, fmt.Errorf("origin transaction must have state index 0")
	}
	return mayBeOrigin, nil
}

// check correctness of the request tokens
func (tx *Transaction) validateRequests(isOrigin bool) error {
	newByAddress := make(map[address.Address]int64)
	tx.Outputs().ForEach(func(addr address.Address, bals []*balance.Balance) bool {
		s := txutil.BalanceOfColor(bals, balance.ColorNew)
		if s != 0 {
			newByAddress[addr] = s
		}
		return true
	})
	for _, reqBlock := range tx.Requests() {
		s, ok := newByAddress[reqBlock.Address()]
		if !ok {
			return errors.New("invalid request tokens")
		}
		newByAddress[reqBlock.Address()] = s - 1
	}
	sum := int64(0)
	for _, s := range newByAddress {
		sum += s
		if s != 0 && s != 1 {
			return errors.New("invalid tokens")
		}
	}
	if isOrigin && sum == 1 {
		return nil
	}
	if !isOrigin && sum == 0 {
		return nil
	}
	return errors.New("invalid tokens")
}

// checks if transaction value part:
// - contains only inputs from the address
// - contains all inputs
// - correctness of colored balances
//func (tx *Transaction) ValidateConsumptionOfInputs(addr *address.Addresses, inputBalances map[valuetransaction.ID][]*balance.Balance) error {
//	if err := waspconn.ValidateBalances(inputBalances); err != nil {
//		return err
//	}
//	var err error
//	var first bool
//	var addrTmp address.Addresses
//	var totalInputs2 int64
//
//	inputBalancesByColor, totalInputs1 := txutil.BalancesByColor(inputBalances)
//
//	tx.Inputs().ForEach(func(outputId valuetransaction.OutputID) bool {
//		if !first {
//			addrTmp = outputId.Addresses()
//			first = true
//		}
//		if addrTmp != outputId.Addresses() {
//			err = errors.New("only 1 address is allowed in inputs")
//			return false
//		}
//		bals, ok := inputBalances[outputId.TransactionID()]
//		if !ok {
//			err = errors.New("unexpected txid in inputs")
//			return false
//		}
//		totalInputs2 += txutil.BalancesSumTotal(bals)
//		return true
//	})
//	if err != nil {
//		return err
//	}
//
//	if totalInputs1 != totalInputs2 {
//		return errors.New("not all provided inputs are consumed")
//	}
//
//	outputBalancesByColor := make(map[balance.Color]int64)
//	tx.Outputs().ForEach(func(_ address.Addresses, bals []*balance.Balance) bool {
//		for _, b := range bals {
//			if s, ok := outputBalancesByColor[b.Color]; !ok {
//				outputBalancesByColor[b.Color] = b.Value
//			} else {
//				outputBalancesByColor[b.Color] = s + b.Value
//			}
//		}
//		return true
//	})
//
//	for col, inb := range inputBalancesByColor {
//		if !(col != balance.ColorNew) {
//			return errors.New("assertion failed: col != balance.ColorNew")
//		}
//		if col == balance.ColorIOTA {
//			continue
//		}
//		outb, ok := outputBalancesByColor[col]
//		if !ok {
//			continue
//		}
//		if outb > inb {
//			// colored supply can't be inflated
//			return errors.New("colored supply can't be inflated")
//		}
//	}
//	return nil
//}
