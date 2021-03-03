// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// implements smart contract transaction.
// smart contract transaction is value transaction with special payload
package sctransaction

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/util"
	"io"
	"sync"
)

// Smart contract transaction wraps value transaction
// the stateSection and requestSection are parsed from the dataPayload of the value transaction
type Transaction struct {
	*valuetransaction.Transaction
	stateSection     *StateSection
	requestSection   []*RequestSection
	cachedProperties coretypes.SCTransactionProperties
}

// function which analyzes the transaction and calculates properties of it
type constructorNew func(transaction *Transaction) (coretypes.SCTransactionProperties, error)

var newProperties constructorNew
var newPropertiesMutex sync.Mutex

func RegisterSemanticAnalyzerConstructor(constr constructorNew) {
	newPropertiesMutex.Lock()
	defer newPropertiesMutex.Unlock()
	if newProperties != nil {
		panic("RegisterSemanticAnalyzerConstructor: already registered")
	}
	newProperties = constr
}

// creates new sc transaction. It is immutable, i.e. tx hash is stable
func NewTransaction(vtx *valuetransaction.Transaction, stateBlock *StateSection, requestBlocks []*RequestSection) (*Transaction, error) {
	ret := &Transaction{
		Transaction:    vtx,
		stateSection:   stateBlock,
		requestSection: requestBlocks,
	}
	var buf bytes.Buffer
	if err := ret.writeDataPayload(&buf); err != nil {
		return nil, err
	}
	if err := vtx.SetDataPayload(buf.Bytes()); err != nil {
		return nil, err
	}
	return ret, nil
}

// parses dataPayload. Error is returned only if pre-parsing succeeded and parsing failed
// usually this can happen only due to targeted attack or
func ParseValueTransaction(vtx *valuetransaction.Transaction) (*Transaction, error) {
	// parse data payload as smart contract metadata
	rdr := bytes.NewReader(vtx.GetDataPayload())
	ret := &Transaction{Transaction: vtx}
	if err := ret.readDataPayload(rdr); err != nil {
		return nil, err
	}
	// semantic validation
	if _, err := ret.Properties(); err != nil {
		return nil, err
	}
	return ret, nil
}

// Properties returns valid properties if sc transaction is semantically correct
func (tx *Transaction) Properties() (coretypes.SCTransactionProperties, error) {
	if tx.cachedProperties != nil {
		return tx.cachedProperties, nil
	}
	var err error
	tx.cachedProperties, err = newProperties(tx)
	return tx.cachedProperties, err
}

func (tx *Transaction) MustProperties() coretypes.SCTransactionProperties {
	ret, err := tx.Properties()
	if err != nil {
		panic(err)
	}
	return ret
}

func (tx *Transaction) State() (*StateSection, bool) {
	return tx.stateSection, tx.stateSection != nil
}

func (tx *Transaction) MustState() *StateSection {
	if tx.stateSection == nil {
		panic("MustState: state block expected")
	}
	return tx.stateSection
}

func (tx *Transaction) Requests() []*RequestSection {
	return tx.requestSection
}

// Sender returns first input address. It is the unique address, because
// ParseValueTransaction doesn't allow other options
func (tx *Transaction) Sender() *address.Address {
	var ret address.Address
	tx.Inputs().ForEachAddress(func(currentAddress address.Address) bool {
		ret = currentAddress
		return false
	})
	return &ret
}

func (tx *Transaction) OutputBalancesByAddress(addr address.Address) ([]*balance.Balance, bool) {
	untyped, ok := tx.Outputs().Get(addr)
	if !ok {
		return nil, false
	}

	ret, ok := untyped.([]*balance.Balance)
	if !ok {
		panic("OutputBalancesByAddress: balances expected")
	}
	return ret, true
}

// function writes bytes of the SC transaction-specific part
func (tx *Transaction) writeDataPayload(w io.Writer) error {
	if tx.stateSection == nil && len(tx.requestSection) == 0 {
		return errors.New("can't encode empty chain transaction")
	}
	if len(tx.requestSection) > 127 {
		return errors.New("max number of request sections 127 exceeded")
	}
	numRequests := byte(len(tx.requestSection))
	b, err := encodeMetaByte(tx.stateSection != nil, numRequests)
	if err != nil {
		return err
	}
	if err = util.WriteByte(w, b); err != nil {
		return err
	}
	if tx.stateSection != nil {
		if err := tx.stateSection.Write(w); err != nil {
			return err
		}
	}
	for _, reqBlk := range tx.requestSection {
		if err := reqBlk.Write(w); err != nil {
			return err
		}
	}
	return nil
}

// readDataPayload parses data stream of data payload to value transaction as smart contract meta data
func (tx *Transaction) readDataPayload(r io.Reader) error {
	var hasState bool
	var numRequests byte
	if b, err := util.ReadByte(r); err != nil {
		return err
	} else {
		hasState, numRequests = decodeMetaByte(b)
	}
	var stateBlock *StateSection
	if hasState {
		stateBlock = &StateSection{}
		if err := stateBlock.Read(r); err != nil {
			return err
		}
	}
	reqBlks := make([]*RequestSection, numRequests)
	for i := range reqBlks {
		reqBlks[i] = &RequestSection{}
		if err := reqBlks[i].Read(r); err != nil {
			return err
		}
	}
	tx.stateSection = stateBlock
	tx.requestSection = reqBlks
	return nil
}

func (tx *Transaction) String() string {
	ret := fmt.Sprintf("TX: %s\n", tx.Transaction.ID().String())
	stateBlock, ok := tx.State()
	if ok {
		vh := stateBlock.StateHash()
		ret += fmt.Sprintf("State: color: %s statehash: %s, ts: %d\n",
			stateBlock.Color().String(),
			vh.String(), stateBlock.Timestamp(),
		)
	} else {
		ret += "State: none\n"
	}
	for i, reqBlk := range tx.Requests() {
		addr := reqBlk.Target()
		ret += fmt.Sprintf("Req #%d: addr: %s code: %s\n", i,
			util.Short(addr.String()), reqBlk.EntryPointCode().String())
	}
	return ret
}
