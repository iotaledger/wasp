package sctransaction

import (
	"errors"
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/txutil"
)

type Properties struct {
	// the only sender of the SC transaction
	sender address.Address
	// is it state transaction (== does it contain valid stateBlock)
	isState bool
	// if isState == true: it states if it is the origin transaction
	isOrigin bool
	// if isState == true: smart contract address
	stateAddress address.Address
	// if isState == true: smart contract color
	stateColor balance.Color
	// number of newly minted tokens
	numMintedTokensByAddr map[address.Address]int64
	numMintedTokens       int64
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
	// check if the sender is exactly one
	// only value transaction with one input address can be parsed as smart contract transactions
	// because we always need to deterministically identify the sender
	senderFound := false
	var err error
	tx.Transaction.Inputs().ForEachAddress(func(addr address.Address) bool {
		if senderFound {
			err = errors.New("smart contract transaction must contain exactly 1 input address")
			return false
		}
		prop.sender = addr
		senderFound = true
		return true
	})
	return err
}

func (prop *Properties) countMintedTokens(tx *Transaction) {
	prop.numMintedTokensByAddr = make(map[address.Address]int64)

	tx.Outputs().ForEach(func(addr address.Address, bals []*balance.Balance) bool {
		v := txutil.BalanceOfColor(bals, balance.ColorNew)
		if v != 0 {
			va, _ := prop.numMintedTokensByAddr[addr]
			prop.numMintedTokensByAddr[addr] = va + v
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
			prop.stateAddress = addr
			return true
		})
		if err != nil {
			return err
		}
		// TODO May change in the future
		if prop.stateAddress != prop.sender {
			return errors.New("SC token must move from the SC address to itself")
		}
		return nil
	}
	// it can be a smart contract origin transaction (color == new)
	// in this case transaction must contain number of requests + 1 newly minted token
	// in the same address
	if len(prop.numMintedTokensByAddr) > 1 {
		return errors.New("in the origin transaction tokens can be minted only to 1 address")
	}
	// one address with minted tokens.
	for stateAddr := range prop.numMintedTokensByAddr {
		prop.stateAddress = stateAddr
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

	numReqByAddr := make(map[address.Address]int64)
	for _, reqBlk := range tx.Requests() {
		n, _ := numReqByAddr[reqBlk.Address()]
		numReqByAddr[reqBlk.Address()] = n + 1
	}

	if prop.isOrigin {
		errWrongTokens := errors.New("wrong minted tokens and/or requests in the origin transaction")
		if len(numReqByAddr) != 1 {
			// must be exactly one target address for requests
			return errWrongTokens
		}
		if _, ok := numReqByAddr[prop.stateAddress]; !ok {
			// that one address must be address of the originated smart contract
			return errWrongTokens
		}
		numMinted, _ := prop.numMintedTokensByAddr[prop.stateAddress]
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
		numMinted, _ := prop.numMintedTokensByAddr[targetAddr]
		if numMinted != numReq {
			return fmt.Errorf("number of minted tokens to the SC address %s is not equal to the number of requests to that SC. Txid = %s",
				targetAddr.String(), tx.ID().String())
		}
	}
	return nil
}

func (prop *Properties) Sender() *address.Address {
	return &prop.sender
}

func (prop *Properties) IsState() bool {
	return prop.isState
}

func (prop *Properties) IsOrigin() bool {
	return prop.isState
}

func (prop *Properties) MustStateAddress() *address.Address {
	if !prop.isState {
		panic("MustStateAddress: must be a state transaction")
	}
	return &prop.stateAddress
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
