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
	stateBlock    *sctransaction.StateSection
	requestBlocks []*sctransaction.RequestSection
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
		requestBlocks: make([]*sctransaction.RequestSection, 0),
	}, nil
}

func NewFromOutputBalances(outputBalances map[valuetransaction.OutputID][]*balance.Balance) (*Builder, error) {
	vtxb, err := vtxbuilder.NewFromOutputBalances(outputBalances)
	if err != nil {
		return nil, err
	}
	return &Builder{
		Builder:       vtxb,
		requestBlocks: make([]*sctransaction.RequestSection, 0),
	}, nil
}

func (txb *Builder) Clone() *Builder {
	ret := &Builder{
		Builder:       txb.Builder.Clone(),
		stateBlock:    txb.stateBlock.Clone(),
		requestBlocks: make([]*sctransaction.RequestSection, len(txb.requestBlocks)),
	}
	for i := range ret.requestBlocks {
		ret.requestBlocks[i] = txb.requestBlocks[i].Clone()
	}
	return ret
}

// CreateStateSection assumes txb contain balances of the smart contract with 'color'.
// It adds state block and moves smart contract token to the same address.
// State block will have 0 state index, 0 timestamp, nil stateHash
// The function is used by VM wrapper to create new state transaction
func (txb *Builder) CreateStateSection(color balance.Color) error {
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
	if err := txb.MoveTokensToAddress(scAddress, color, 1); err != nil {
		return err
	}
	txb.stateBlock = sctransaction.NewStateSection(sctransaction.NewStateSectionParams{
		Color: color,
	})
	return nil
}

// CreateOriginStateSection
// - initializes origin state transaction of the chain with originAddress in the builder.
// - mints chain token, sets origin state hash
// - sets state index and timestamp to 0
func (txb *Builder) CreateOriginStateSection(stateHash *hashing.HashValue, originAddress *address.Address) error {
	if txb.stateBlock != nil {
		return errors.New("can't set state block twice")
	}
	if err := txb.MintColor(*originAddress, balance.ColorIOTA, 1); err != nil {
		return err
	}
	txb.stateBlock = sctransaction.NewStateSection(sctransaction.NewStateSectionParams{
		Color:      balance.ColorNew,
		BlockIndex: 0,
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

// AddRequestSectionWithTransfer adds request block with the request
// token and adds respective outputs for the colored transfers
func (txb *Builder) AddRequestSection(req *sctransaction.RequestSection) error {
	targetAddr := (address.Address)(req.Target().ChainID())
	if err := txb.MintColor(targetAddr, balance.ColorIOTA, 1); err != nil {
		return err
	}
	var err error
	req.Transfer().Iterate(func(col balance.Color, bal int64) bool {
		if err = txb.MoveTokensToAddress(targetAddr, col, bal); err != nil {
			return false
		}
		return true
	})
	txb.requestBlocks = append(txb.requestBlocks, req)
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
