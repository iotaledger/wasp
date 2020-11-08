package sctransaction

import (
	"errors"
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"

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
	stateColor balance.Color
	// number of newly minted tokens
	numMintedTokensByChain map[coretypes.ChainID]int64
	numMintedTokens        int64
	// number of requests
	numRequests int
}

func (tx *Transaction) calcProperties() (*Properties, error) {
	ret := &Properties{}
	if err := ret.analyzeSender(tx); err != nil {
		return nil, err
	}

	ret.countMintedTokens(tx)

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

func (prop *Properties) countMintedTokens(tx *Transaction) {
	prop.numMintedTokensByChain = make(map[coretypes.ChainID]int64)

	tx.Outputs().ForEach(func(addr address.Address, bals []*balance.Balance) bool {
		v := txutil.BalanceOfColor(bals, balance.ColorNew)
		if v != 0 {
			va, _ := prop.numMintedTokensByChain[(coretypes.ChainID)(addr)]
			prop.numMintedTokensByChain[(coretypes.ChainID)(addr)] = va + v
			prop.numMintedTokens += v
		}
		return true
	})
}

func (prop *Properties) analyzeStateBlock(tx *Transaction) error {
	stateBlock, ok := tx.State()
	prop.isState = ok
	if !ok {
		return nil
	}

	var err error

	if stateBlock.Color() != balance.ColorNew {
		prop.stateColor = stateBlock.Color()
		// it is not origin. Must contain exactly one output with value 1 of that color
		var v int64
		tx.Outputs().ForEach(func(addr address.Address, bals []*balance.Balance) bool {
			v += txutil.BalanceOfColor(bals, prop.stateColor)
			if v > 1 {
				err = errors.New("sc transaction must contain exactly one sc token output")
				return false
			}
			prop.chainID = (coretypes.ChainID)(addr)
			return true
		})
		if err != nil {
			return err
		}
		// TODO May change in the future
		if prop.chainID != (coretypes.ChainID)(prop.senderAddress) {
			return errors.New("SC token must move from the SC address to itself")
		}
		return nil
	}
	// it can be a smart contract origin transaction (color == new)
	// in this case transaction must contain number of requests + 1 newly minted token
	// in the same address
	if len(prop.numMintedTokensByChain) > 1 {
		return errors.New("in the origin transaction tokens can be minted only to 1 address")
	}
	// one address with minted tokens.
	for stateAddr := range prop.numMintedTokensByChain {
		prop.chainID = stateAddr
		break
	}
	prop.isOrigin = true
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
	prop.numRequests = len(tx.Requests())

	numReqByAddr := make(map[coretypes.ChainID]int64)
	for _, reqBlk := range tx.Requests() {
		n, _ := numReqByAddr[reqBlk.Target().ChainID()]
		numReqByAddr[reqBlk.Target().ChainID()] = n + 1
	}

	if prop.isOrigin {
		errWrongTokens := errors.New("wrong minted tokens and/or requests in the origin transaction")
		if len(numReqByAddr) != 1 {
			// must be exactly one target address for requests
			return errWrongTokens
		}
		if _, ok := numReqByAddr[prop.chainID]; !ok {
			// that one address must be address of the originated smart contract
			return errWrongTokens
		}
		numMinted, _ := prop.numMintedTokensByChain[prop.chainID]
		if numMinted != int64(len(tx.Requests())+1) {
			// number of minted must be one more that number of requests
			return errWrongTokens
		}
		return nil
	}
	// not origin transaction

	// IMPORTANT: number of minted tokens to an address of some smart contract must be exactly equal to
	// the number of requests to that smart contract.
	// Total number of minted tokens can be larger that the total number of requests, however the
	// rest of minted tokens must be in outputs different from any of the target smart contract address
	for targetAddr, numReq := range numReqByAddr {
		numMinted, _ := prop.numMintedTokensByChain[targetAddr]
		if numMinted != numReq {
			return fmt.Errorf("number of minted tokens to the SC address %s is not equal to the number of requests to that SC. Txid = %s",
				targetAddr.String(), tx.ID().String())
		}
	}
	return nil
}

func (prop *Properties) Sender() *address.Address {
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
