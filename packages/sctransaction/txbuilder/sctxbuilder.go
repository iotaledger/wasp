package txbuilder

import (
	"errors"
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/txutil"
	"github.com/iotaledger/wasp/packages/txutil/vtxbuilder"
)

type Builder struct {
	*vtxbuilder.Builder
	stateBlock    *sctransaction.StateBlock
	requestBlocks []*sctransaction.RequestBlock
}

var (
	errorWrongScToken = errors.New("wrong or nonexistent smart contract token in inputs")
)

func NewFromAddressBalances(scAddress *address.Address, addressBalances map[valuetransaction.ID][]*balance.Balance) (*Builder, error) {
	vtxb, err := vtxbuilder.NewFromAddressBalances(scAddress, addressBalances)
	if err != nil {
		return nil, err
	}
	return &Builder{
		Builder:       vtxb,
		requestBlocks: make([]*sctransaction.RequestBlock, 0),
	}, nil
}

func NewFromOutputBalances(outputBalances map[valuetransaction.OutputID][]*balance.Balance) (*Builder, error) {
	vtxb, err := vtxbuilder.NewFromOutputBalances(outputBalances)
	if err != nil {
		return nil, err
	}
	return &Builder{
		Builder:       vtxb,
		requestBlocks: make([]*sctransaction.RequestBlock, 0),
	}, nil
}

func (txb *Builder) Clone() *Builder {
	ret := &Builder{
		Builder:       txb.Builder.Clone(),
		stateBlock:    txb.stateBlock.Clone(),
		requestBlocks: make([]*sctransaction.RequestBlock, len(txb.requestBlocks)),
	}
	for i := range ret.requestBlocks {
		ret.requestBlocks[i] = txb.requestBlocks[i].Clone()
	}
	return ret
}

// CreateStateBlock assumes txb contain balances of the smart contract with 'color'.
// It adds state block and moves smart contract token to the same address.
// State block will have 0 state index, 0 timestamp, nil stateHash
// The function is used by VM wrapper to create new state transaction
func (txb *Builder) CreateStateBlock(color balance.Color) error {
	if txb.stateBlock != nil {
		return errors.New("can't set state block twice")
	}
	if color == balance.ColorNew {
		return errors.New("can't use 'ColorNew'")
	}

	if txb.GetInputBalance(color) == 0 {
		return fmt.Errorf("non existent smart contract token with color %s", color.String())
	}
	foundAddress := false
	var scAddress address.Address
	txb.ForEachInputBalance(func(oid *valuetransaction.OutputID, bals []*balance.Balance) bool {
		if txutil.BalanceOfColor(bals, color) > 0 {
			scAddress = oid.Address()
			foundAddress = true
			return false
		}
		return true
	})
	if !foundAddress {
		return errorWrongScToken
	}
	if err := txb.MoveToAddress(scAddress, color, 1); err != nil {
		return err
	}
	txb.stateBlock = sctransaction.NewStateBlock(sctransaction.NewStateBlockParams{
		Color: color,
	})
	return nil
}

// CreateOriginStateBlock initalizes origin state transaction of the smart contract
// with address scAddress in the builder. It mints smart contract token, sets origin state hash
// It sets state index and timestamp to 0
func (txb *Builder) CreateOriginStateBlock(stateHash *hashing.HashValue, scAddress *address.Address) error {
	if txb.stateBlock != nil {
		return errors.New("can't set state block twice")
	}
	if err := txb.MintColor(*scAddress, balance.ColorIOTA, 1); err != nil {
		return err
	}
	txb.stateBlock = sctransaction.NewStateBlock(sctransaction.NewStateBlockParams{
		Color:      balance.ColorNew,
		StateIndex: 0,
		StateHash:  *stateHash,
		Timestamp:  0,
	})
	return nil
}

func (txb *Builder) SetStateParams(stateIndex uint32, stateHash *hashing.HashValue, timestamp int64) error {
	if txb.stateBlock == nil {
		return fmt.Errorf("state block not set")
	}
	txb.stateBlock.WithStateParams(stateIndex, stateHash, timestamp)
	return nil
}

// AddRequestBlock adds new request block to the builder. It automatically handles request token
func (txb *Builder) AddRequestBlock(reqBlk *sctransaction.RequestBlock) error {
	return txb.AddRequestBlockWithTransfer(reqBlk, nil, nil)
}

// AddRequestBlockWithTransfer adds request block with the request token and adds respective
// outputs for the colored transfers
func (txb *Builder) AddRequestBlockWithTransfer(reqBlk *sctransaction.RequestBlock, targetAddr *address.Address, bals map[balance.Color]int64) error {
	if err := txb.MintColor(reqBlk.Address(), balance.ColorIOTA, 1); err != nil {
		return err
	}
	for col, b := range bals {
		if err := txb.MoveToAddress(*targetAddr, col, b); err != nil {
			return err
		}
	}
	txb.requestBlocks = append(txb.requestBlocks, reqBlk)
	return nil
}

func (txb *Builder) Build(useAllInputs bool) (*sctransaction.Transaction, error) {
	return sctransaction.NewTransaction(
		txb.Builder.Build(useAllInputs),
		txb.stateBlock,
		txb.requestBlocks,
	)
}

// ignores SC part
func (txb *Builder) BuildValueTransactionOnly(useAllInputs bool) *valuetransaction.Transaction {
	return txb.Builder.Build(useAllInputs)
}
