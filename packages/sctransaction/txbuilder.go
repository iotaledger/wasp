package sctransaction

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"sort"
)

// object with interface to build SC transaction and value transaction within it
// object panics if attempted to modify structure after finalization
type TransactionBuilder struct {
	inputs        *valuetransaction.Inputs
	outputs       map[address.Address]map[balance.Color]int64
	stateBlock    *StateBlock
	requestBlocks []*RequestBlock
	finalized     bool
}

func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{
		inputs:        valuetransaction.NewInputs(),
		outputs:       make(map[address.Address]map[balance.Color]int64),
		requestBlocks: make([]*RequestBlock, 0),
	}
}

func (txb *TransactionBuilder) Finalize() (*Transaction, error) {
	if txb.finalized {
		return nil, errors.New("attempt to modify already finalized transaction builder")
	}

	outputs := make(map[address.Address][]*balance.Balance)
	for addr, bmap := range txb.outputs {
		// sort to have deterministic result
		blst := make([]balance.Color, 0, len(bmap))
		for c := range bmap {
			blst = append(blst, c)
		}

		sort.Slice(blst, func(i, j int) bool {
			return bytes.Compare(blst[i][:], blst[j][:]) < 0
		})

		outputs[addr] = make([]*balance.Balance, 0)
		for _, c := range blst {
			outputs[addr] = append(outputs[addr], balance.New(c, bmap[c]))
		}
	}
	txv := valuetransaction.New(txb.inputs, valuetransaction.NewOutputs(outputs))
	ret, err := NewTransaction(txv, txb.stateBlock, txb.requestBlocks)
	if err != nil {
		return nil, err
	}
	txb.finalized = true
	return ret, nil
}

func (txb *TransactionBuilder) MustAddInputs(oid ...valuetransaction.OutputID) {
	if err := txb.AddInputs(oid...); err != nil {
		panic(err)
	}
}

func (txb *TransactionBuilder) AddInputs(oid ...valuetransaction.OutputID) error {
	//sort inputs to have deterministic order
	oidClone := make([]valuetransaction.OutputID, len(oid))
	copy(oidClone, oid)

	sort.Slice(oidClone, func(i, j int) bool {
		return bytes.Compare(oidClone[i][:], oidClone[j][:]) < 0
	})

	// check if all different

	for i := 0; i < len(oidClone)-1; i++ {
		if bytes.Equal(oidClone[i][:], oidClone[i+1][:]) {
			return fmt.Errorf("some inputs are identical")
		}
	}

	for _, o := range oidClone {
		txb.inputs.Add(o)
	}
	return nil
}

func (txb *TransactionBuilder) AddBalanceToOutput(addr address.Address, bal *balance.Balance) {
	if _, ok := txb.outputs[addr]; !ok {
		txb.outputs[addr] = make(map[balance.Color]int64)
	}
	balances := txb.outputs[addr]
	if val, ok := balances[bal.Color()]; ok {
		balances[bal.Color()] = val + bal.Value()
	} else {
		balances[bal.Color()] = bal.Value()
	}
}

func (txb *TransactionBuilder) AddStateBlock(par NewStateBlockParams) {
	txb.stateBlock = NewStateBlock(par)
}

func (txb *TransactionBuilder) AddRequestBlock(reqBlk *RequestBlock) {
	txb.requestBlocks = append(txb.requestBlocks, reqBlk)
}
