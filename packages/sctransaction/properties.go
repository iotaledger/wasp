package sctransaction

import (
	"errors"
	"github.com/iotaledger/wasp/packages/coretypes"
	accounts "github.com/iotaledger/wasp/packages/vm/balances"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/txutil"
)

type Properties struct {
	// the only senderAddress of the SC transaction
	senderAddress address.Address
	// is it state transaction (== does it contain valid stateBlock)
	isState bool
	// if isState == true: it states if it is the origin transaction
	isOrigin bool
	// if isState == true: chainID
	chainID coretypes.ChainID
	// if isState == true: smart contract color
	stateColor      balance.Color
	numMintedTokens int64
	// number of requests
	numRequests int
}

func (tx *Transaction) calcProperties() (*Properties, error) {
	ret := &Properties{}
	if err := ret.analyzeSender(tx); err != nil {
		return nil, err
	}
	if err := ret.analyzeStateBlock(tx); err != nil {
		return nil, err
	}
	if err := ret.analyzeRequestBlocks(tx); err != nil {
		return nil, err
	}
	return ret, nil
}

func (prop *Properties) analyzeSender(tx *Transaction) error {
	// check if the senderAddress is exactly one
	// only value transaction with one input address can be parsed as smart contract transactions
	// because we always need to deterministically identify the senderAddress
	senderFound := false
	var err error
	tx.Transaction.Inputs().ForEachAddress(func(addr address.Address) bool {
		if senderFound {
			err = errors.New("smart contract transaction must contain exactly 1 input address")
			return false
		}
		prop.senderAddress = addr
		senderFound = true
		return true
	})
	return err
}

var ErrWrongChainToken = errors.New("sc transaction must contain exactly one chain token output")

func (prop *Properties) analyzeStateBlock(tx *Transaction) error {
	stateSection, ok := tx.State()
	prop.isState = ok
	if !ok {
		return nil
	}

	var err error

	prop.isOrigin = stateSection.Color() == balance.ColorNew
	sectionColor := stateSection.Color()

	// must contain exactly one output with sectionColor. It ca be NewColor for origin
	var v int64
	tx.Outputs().ForEach(func(addr address.Address, bals []*balance.Balance) bool {
		v += txutil.BalanceOfColor(bals, sectionColor)
		if v > 1 {
			err = ErrWrongChainToken
			return false
		}
		prop.chainID = (coretypes.ChainID)(addr)
		return true
	})
	if v != 1 {
		return ErrWrongChainToken
	}
	if err != nil {
		return err
	}
	prop.stateColor = (balance.Color)(tx.Transaction.ID())
	return nil
}

func (prop *Properties) analyzeRequestBlocks(tx *Transaction) error {
	if !prop.isState && len(tx.Requests()) == 0 {
		return errors.New("smart contract transaction which does not contain state block must contain at least one request")
	}
	if len(tx.Requests()) == 0 {
		return nil
	}
	if prop.isOrigin {
		return errors.New("origin transaction should not contain requests")
	}
	prop.numRequests = len(tx.Requests())

	// sum up transfers of requests by target chain
	reqTransfersByTargetChain := make(map[coretypes.ChainID]map[balance.Color]int64)
	for _, req := range tx.Requests() {
		chainid := req.targetContractID.ChainID()
		m, ok := reqTransfersByTargetChain[chainid]
		if !ok {
			m = make(map[balance.Color]int64)
			reqTransfersByTargetChain[chainid] = m
		}
		req.transfer.AddToMap(m)
		// add one request token
		numMinted, _ := m[balance.ColorNew]
		m[balance.ColorNew] = numMinted + 1
	}
	// check if transfers from requests equal to transfers by output address
	for chainid, m := range reqTransfersByTargetChain {
		bals, ok := tx.OutputBalancesByAddress(address.Address(chainid))
		if !ok {
			return errors.New("can't find outputs for request section")
		}
		txBals := accounts.NewColoredBalancesFromMap(txutil.BalancesToMap(bals))
		reqBals := accounts.NewColoredBalancesFromMap(m)
		if !txBals.Equal(reqBals) {
			return errors.New("mismatch between transfer data in request section and tx outputs")
		}
	}
	// TODO free minted tokens
	return nil
}

func (prop *Properties) SenderAddress() *address.Address {
	return &prop.senderAddress
}

func (prop *Properties) IsState() bool {
	return prop.isState
}

func (prop *Properties) IsOrigin() bool {
	return prop.isState
}

func (prop *Properties) MustChainID() *coretypes.ChainID {
	if !prop.isState {
		panic("MustChainID: must be a state transaction")
	}
	return &prop.chainID
}

func (prop *Properties) MustStateColor() *balance.Color {
	if !prop.isState {
		panic("MustStateColor: must be a state transaction")
	}
	return &prop.stateColor
}

// NumFreeMintedTokens return total minted tokens minus number of requests
func (prop *Properties) NumFreeMintedTokens() int64 {
	if prop.isOrigin {
		return 0
	}
	return prop.numMintedTokens - int64(prop.numRequests)
}
