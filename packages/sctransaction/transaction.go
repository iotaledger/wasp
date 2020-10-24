// implement smart contract transaction.
// smart contract transaction is value transaction with special payload
package sctransaction

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

// Smart contract transaction wraps value transaction
// the stateBlock and requestBlocks are parsed from the dataPayload of the value transaction
type Transaction struct {
	*valuetransaction.Transaction
	stateBlock    *StateBlock
	requestBlocks []*RequestBlock
	properties    *Properties // cached properties. If nil, transaction is semantically validated and properties are calculated
}

// creates new sc transaction. It is immutable, i.e. tx hash is stable
func NewTransaction(vtx *valuetransaction.Transaction, stateBlock *StateBlock, requestBlocks []*RequestBlock) (*Transaction, error) {
	ret := &Transaction{
		Transaction:   vtx,
		stateBlock:    stateBlock,
		requestBlocks: requestBlocks,
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

// return valid properties if sc transaction is semantically correct
func (tx *Transaction) Properties() (*Properties, error) {
	if tx.properties != nil {
		return tx.properties, nil
	}
	var err error
	tx.properties, err = tx.calcProperties()
	return tx.properties, err
}

func (tx *Transaction) MustProperties() *Properties {
	ret, err := tx.Properties()
	if err != nil {
		panic(err)
	}
	return ret
}

func (tx *Transaction) State() (*StateBlock, bool) {
	return tx.stateBlock, tx.stateBlock != nil
}

func (tx *Transaction) MustState() *StateBlock {
	if tx.stateBlock == nil {
		panic("MustState: state block expected")
	}
	return tx.stateBlock
}

func (tx *Transaction) Requests() []*RequestBlock {
	return tx.requestBlocks
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

func (tx *Transaction) OutputBalancesByAddress(addr *address.Address) ([]*balance.Balance, bool) {
	untyped, ok := tx.Outputs().Get(*addr)
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
	if tx.stateBlock == nil && len(tx.requestBlocks) == 0 {
		return errors.New("can't encode empty sc transaction")
	}
	if len(tx.requestBlocks) > 127 {
		return errors.New("max number of request blocks 127 exceeded")
	}
	numRequests := byte(len(tx.requestBlocks))
	b, err := encodeMetaByte(tx.stateBlock != nil, numRequests)
	if err != nil {
		return err
	}
	if err = util.WriteByte(w, b); err != nil {
		return err
	}
	if tx.stateBlock != nil {
		if err := tx.stateBlock.Write(w); err != nil {
			return err
		}
	}
	for _, reqBlk := range tx.requestBlocks {
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
	var stateBlock *StateBlock
	if hasState {
		stateBlock = &StateBlock{}
		if err := stateBlock.Read(r); err != nil {
			return err
		}
	}
	reqBlks := make([]*RequestBlock, numRequests)
	for i := range reqBlks {
		reqBlks[i] = &RequestBlock{}
		if err := reqBlks[i].Read(r); err != nil {
			return err
		}
	}
	tx.stateBlock = stateBlock
	tx.requestBlocks = reqBlks
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
