package txbuilder

import (
	"errors"
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
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

// CreateOriginStateSection
// - initializes origin state transaction of the chain with originAddress in the builder.
// - mints chain token, sets origin state hash
// - sets state index and timestamp to 0
func (txb *Builder) CreateOriginStateSection(stateHash hashing.HashValue, originAddress *address.Address) error {
	if txb.stateBlock != nil {
		return errors.New("can't set state block twice")
	}
	if err := txb.MintColor(*originAddress, balance.ColorIOTA, 1); err != nil {
		return err
	}
	txb.stateBlock = sctransaction.NewStateSection(sctransaction.NewStateSectionParams{
		Color:      balance.ColorNew,
		BlockIndex: 0,
		StateHash:  stateHash,
		Timestamp:  0,
	})
	return nil
}

func (txb *Builder) SetStateParams(stateIndex uint32, stateHash hashing.HashValue, timestamp int64) error {
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
	tran := req.Transfer()
	if tran != nil {
		tran.Iterate(func(col balance.Color, bal int64) bool {
			if err = txb.MoveTokensToAddress(targetAddr, col, bal); err != nil {
				return false
			}
			return true
		})
	}
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
