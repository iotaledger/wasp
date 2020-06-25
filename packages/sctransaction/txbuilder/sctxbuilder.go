package txbuilder

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder/vtxbuilder"
)

type Builder struct {
	*vtxbuilder.Builder
	stateBlock    *sctransaction.StateBlock
	requestBlocks []*sctransaction.RequestBlock
}

func NewFromAddressBalances(addr *address.Address, addressBalances map[valuetransaction.ID][]*balance.Balance) (*Builder, error) {
	vtxb, err := vtxbuilder.NewFromAddressBalances(addr, addressBalances)
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

func (txb *Builder) AddStateBlock(stateBlock *sctransaction.StateBlock) {
	txb.stateBlock = stateBlock
}

func (txb *Builder) AddRequestBlock(reqBlk *sctransaction.RequestBlock) {
	txb.requestBlocks = append(txb.requestBlocks, reqBlk)
}
