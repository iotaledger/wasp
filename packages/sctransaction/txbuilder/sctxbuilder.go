package txbuilder

import (
	"errors"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder/vtxbuilder"
	"github.com/iotaledger/wasp/packages/util"
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

func (txb *Builder) AddOriginStateBlock(stateHash *hashing.HashValue, scAddress *address.Address) error {
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

func (txb *Builder) AddStateBlock(stateBlock *sctransaction.StateBlock) error {
	if stateBlock.Color() == balance.ColorNew {
		return errors.New("can't use 'ColorNew'")
	}
	if txb.GetInputBalance(stateBlock.Color()) != 1 {
		return errorWrongScToken
	}
	foundAddress := false
	var scAddress address.Address
	txb.ForEachInputBalance(func(oid *valuetransaction.OutputID, bals []*balance.Balance) bool {
		if util.BalanceOfColor(bals, stateBlock.Color()) == 1 {
			scAddress = oid.Address()
			foundAddress = true
			return false
		}
		return true
	})
	if !foundAddress {
		return errorWrongScToken
	}
	if err := txb.MoveToAddress(scAddress, stateBlock.Color(), 1); err != nil {
		return err
	}
	txb.stateBlock = stateBlock
	return nil
}

func (txb *Builder) AddRequestBlock(reqBlk *sctransaction.RequestBlock) error {
	if err := txb.MintColor(reqBlk.Address(), balance.ColorIOTA, 1); err != nil {
		return err
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
