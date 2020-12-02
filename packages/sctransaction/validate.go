package sctransaction

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
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

	balances, hasAddress := tx.OutputBalancesByAddress(*addr)
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
	if mayBeOrigin && stateBlock.BlockIndex() != 0 {
		return false, fmt.Errorf("origin transaction must have state index 0")
	}
	return mayBeOrigin, nil
}

// check correctness of the request tokens
func (tx *Transaction) validateRequests(isOrigin bool) error {
	newByTargetChain := make(map[coretypes.ChainID]int64)
	tx.Outputs().ForEach(func(addr address.Address, bals []*balance.Balance) bool {
		s := txutil.BalanceOfColor(bals, balance.ColorNew)
		if s != 0 {
			newByTargetChain[(coretypes.ChainID)(addr)] = s
		}
		return true
	})
	for _, reqBlock := range tx.Requests() {
		s, ok := newByTargetChain[reqBlock.Target().ChainID()]
		if !ok {
			return errors.New("invalid request tokens")
		}
		newByTargetChain[reqBlock.Target().ChainID()] = s - 1
	}
	sum := int64(0)
	for _, s := range newByTargetChain {
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
