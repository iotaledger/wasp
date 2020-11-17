// tx builder for VM
package statetxbuilder

import (
	"errors"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type Builder struct {
	stateSection         *sctransaction.StateSection
	requestSections      []*sctransaction.RequestSection
	numRequestsProcessed map[balance.Color]int64
}

func New(chainColor balance.Color) (*Builder, error) {
	if chainColor == balance.ColorNew || chainColor == balance.ColorIOTA {
		return nil, errors.New("statetxbuilder.New: wrong chain color")
	}
	ret := &Builder{
		stateSection:         sctransaction.NewStateSection(sctransaction.NewStateSectionParams{Color: chainColor}),
		requestSections:      make([]*sctransaction.RequestSection, 0),
		numRequestsProcessed: make(map[balance.Color]int64),
	}
	return ret, nil
}

func (txb *Builder) Clone() *Builder {
	ret := &Builder{
		stateSection:    txb.stateSection.Clone(),
		requestSections: make([]*sctransaction.RequestSection, len(txb.requestSections)),
	}
	for i := range ret.requestSections {
		ret.requestSections[i] = txb.requestSections[i].Clone()
	}
	for col, n := range txb.numRequestsProcessed {
		ret.numRequestsProcessed[col] = n
	}
	return ret
}

func (txb *Builder) SetStateParams(stateIndex uint32, stateHash *hashing.HashValue, timestamp int64) error {
	txb.stateSection.WithStateParams(stateIndex, stateHash, timestamp)
	return nil
}

func (txb *Builder) RequestProcessed(reqid coretypes.RequestID) {
	color := balance.Color(*reqid.TransactionID())
	n, _ := txb.numRequestsProcessed[color]
	txb.numRequestsProcessed[color] = n + 1
}

// AddRequestSectionWithTransfer adds request block with the request
// token and adds respective outputs for the colored transfers
func (txb *Builder) AddRequestSection(req *sctransaction.RequestSection) {
	txb.requestSections = append(txb.requestSections, req)
}

func (txb *Builder) Build(chainAddress *address.Address, addressBalances map[valuetransaction.ID][]*balance.Balance) (*sctransaction.Transaction, error) {
	vtxb, err := newValueTxBuilder(*chainAddress, addressBalances)
	if err != nil {
		return nil, err
	}
	// move chain token
	err = vtxb.MoveTokensToAddress(*chainAddress, txb.stateSection.Color(), 1)
	if err != nil {
		return nil, err
	}
	// Erase request tokens
	for col, n := range txb.numRequestsProcessed {
		if err = vtxb.EraseColor(*chainAddress, col, n); err != nil {
			return nil, err
		}
	}
	for _, req := range txb.requestSections {
		var err error
		targetAddr := address.Address(req.Target().ChainID())
		// mint request token to the target address
		if err = vtxb.MintColor(targetAddr, balance.ColorIOTA, 1); err != nil {
			return nil, err
		}
		// move transfer tokens to the target address
		req.Transfer().Iterate(func(col balance.Color, bal int64) bool {
			err = vtxb.MoveTokensToAddress(targetAddr, col, bal)
			return err == nil
		})
		if err != nil {
			return nil, err
		}
	}
	return sctransaction.NewTransaction(
		vtxb.build(),
		txb.stateSection,
		txb.requestSections,
	)
}
