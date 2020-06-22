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
}

// creates new sc transaction. It is immutable, i.e. tx hash is stable
func NewTransaction(vtx *valuetransaction.Transaction, stateBlock *StateBlock, requestBlocks []*RequestBlock) (*Transaction, error) {
	ret := &Transaction{
		Transaction:   vtx,
		stateBlock:    stateBlock,
		requestBlocks: requestBlocks,
	}
	scpayload, err := ret.DataScPayloadBytes()
	if err != nil {
		return nil, err
	}
	if err := vtx.SetDataPayload(scpayload); err != nil {
		return nil, err
	}
	return ret, nil
}

// 1 + balance.ColorLength // metabyte + color

func NewFromBytes(data []byte) (*Transaction, error) {
	vtx, _, err := valuetransaction.FromBytes(data)
	if err != nil {
		return nil, err
	}
	tx, err := ParseValueTransaction(vtx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// parses dataPayload. Error is returned only if pre-parsing succeeded and parsing failed
// usually this can happen only due to targeted attack or
func ParseValueTransaction(vtx *valuetransaction.Transaction) (*Transaction, error) {
	rdr := bytes.NewReader(vtx.GetDataPayload())
	ret := &Transaction{Transaction: vtx}
	if err := ret.ReadDataPayload(rdr); err != nil {
		return nil, err
	}
	return ret, nil
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

func (tx *Transaction) MustRequest(index uint16) *RequestBlock {
	return tx.requestBlocks[index]
}

func (tx *Transaction) DataScPayloadBytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := tx.WriteDataPayload(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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
func (tx *Transaction) WriteDataPayload(w io.Writer) error {
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

func (tx *Transaction) ReadDataPayload(r io.Reader) error {
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
		vh := stateBlock.VariableStateHash()
		ret += fmt.Sprintf("State: color: %s statehash: %s, ts: %d\n",
			stateBlock.Color().String(),
			vh.String(), stateBlock.Timestamp(),
		)
	} else {
		ret += "State: none\n"
	}
	for i, reqBlk := range tx.Requests() {
		addr := reqBlk.Address()
		ret += fmt.Sprintf("Req #%d: addr: %s code: %d\n", i, util.Short(addr.String()), reqBlk.RequestCode())
	}
	return ret
}

var errorWrongTokens = errors.New("rong number of request tokens")

// StateAddress returns address of the smart contract the state block is targeted to
func (tx *Transaction) StateAddress() (address.Address, bool, error) {
	stateBlock, ok := tx.State()
	if !ok {
		return address.Address{}, false, errors.New("not a state transaction")
	}

	var ret address.Address
	var totalTokens int64
	if stateBlock.Color() != balance.ColorNew {
		tx.Outputs().ForEach(func(addr address.Address, bals []*balance.Balance) bool {
			for _, bal := range bals {
				if bal.Color() == stateBlock.Color() {
					ret = addr
					totalTokens += bal.Value()
				}
			}
			return true
		})
		if totalTokens != 1 {
			return address.Address{}, false, errorWrongTokens
		}
		return ret, true, nil
	}
	// origin case
	newByAddress := make(map[address.Address]int64)
	tx.Outputs().ForEach(func(addr address.Address, bals []*balance.Balance) bool {
		s := util.BalanceOfColor(bals, balance.ColorNew)
		if s != 0 {
			newByAddress[addr] = s
		}
		return true
	})
	for _, reqBlock := range tx.Requests() {
		s, ok := newByAddress[reqBlock.Address()]
		if !ok {
			return address.Address{}, false, errorWrongTokens
		}
		newByAddress[reqBlock.Address()] = s - 1
	}
	// must be left only one token with 1 and the rest with 0
	sum := int64(0)
	for addr, s := range newByAddress {
		if s == 1 {
			ret = addr
		}
		if s != 0 && s != 1 {
			return address.Address{}, false, errorWrongTokens
		}
		sum += s
	}
	if sum != 1 {
		return address.Address{}, false, errorWrongTokens
	}
	return ret, true, nil
}
