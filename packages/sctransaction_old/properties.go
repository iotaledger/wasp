// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package semantically analyzes parsed smart contract transaction
// return object with transaction properties or error if semantically incorrect
package sctransaction_old

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/txutil"
	"golang.org/x/xerrors"
)

// properties represents result of analysis and semantic check of the SC transaction essence

type properties struct {
	isState bool
	// if isState == true it states if it is the origin transaction, otherwise uninterpreted
	isOrigin bool
	// if isState == true: chainID
	chainID coretypes.ChainID
	// chainAddress == chainID
	chainAddress ledgerstate.Address
	// if isState == true: smart contract NFT color.
	stateColor ledgerstate.Color
	// stateHash from state section
	stateHash hashing.HashValue
	// number of requests
	numRequests int
	// serialized data payload size
	dataPayloadSize uint32
}

// CalcProperties analyzes the transaction and returns its properties
func calcProperties(tx *TransactionEssence) (coretypes.SCTransactionProperties, error) {
	ret := &properties{
		dataPayloadSize: uint32(len(tx.TransactionEssence.Payload().Bytes())),
	}
	if err := ret.analyzeStateBlock(tx); err != nil {
		return nil, err
	}
	if err := ret.analyzeRequestBlocks(tx); err != nil {
		return nil, err
	}
	return ret, nil
}

func (prop *properties) analyzeStateBlock(tx *TransactionEssence) error {
	stateSection, ok := tx.State()
	prop.isState = ok
	if !ok {
		return nil
	}

	prop.stateHash = stateSection.StateHash()

	var err error

	prop.isOrigin = stateSection.Color() == ledgerstate.ColorMint
	sectionColor := stateSection.Color()
	if sectionColor == ledgerstate.ColorIOTA {
		return xerrors.New("state section color can't be IOTAColor")
	}

	// must contain exactly one output with sectionColor. It can be NewColor for origin
	var v int64
	err = fmt.Errorf("can't find chain token output of color %s", sectionColor.String())
	tx.Outputs().ForEach(func(addr address.Address, bals []*balance.Balance) bool {
		v += txutil.BalanceOfColor(bals, sectionColor)
		if v > 1 {
			err = fmt.Errorf("can't be more than one chain token output of color %s", sectionColor.String())
			return false
		}
		if err != nil && v == 1 {
			prop.chainID = coretypes.ChainID(addr)
			prop.chainAddress = addr
			err = nil
		}
		return true
	})
	if err != nil {
		return err
	}
	if prop.isOrigin {
		prop.stateColor = balance.Color(prop.txid)
	} else {
		prop.stateColor = sectionColor
	}
	return nil
}

func (prop *properties) analyzeRequestBlocks(tx *TransactionEssence) error {
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
		chainid := req.Target().ChainID()
		m, ok := reqTransfersByTargetChain[chainid]
		if !ok {
			m = make(map[balance.Color]int64)
			reqTransfersByTargetChain[chainid] = m
		}
		req.Transfer().AddToMap(m)
		// add one request token
		numMinted, _ := m[balance.ColorNew]
		m[balance.ColorNew] = numMinted + 1
	}
	var err error
	// validate all outputs against request transfers
	tx.Transaction.Outputs().ForEach(func(addr address.Address, bals []*balance.Balance) bool {
		m, ok := reqTransfersByTargetChain[coretypes.ChainID(addr)]
		if !ok {
			// ignore outputs to outside addresses
			return true
		}
		diff := coretypes.NewFromBalances(bals).Diff(coretypes.NewFromMap(m))
		if prop.isState && addr == prop.chainAddress {
			if diff.Len() != 1 && diff.Balance(prop.stateColor) != 1 {
				// output to the self in the state transaction can't contain free tokens
				err = fmt.Errorf("wrong output to chain address in the state transaction")
				return false
			}
			return true
		}
		if diff.Len() == 0 {
			// exact match
			return true
		}
		if diff.Balance(balance.ColorNew) != 0 {
			// semantic rule: we require number of minted tokens to the target chain exactly equal
			// to the number of requests to that chain
			err = fmt.Errorf("wrong number of minted tokens in the output to the address %s", addr.String())
			return false
		}
		if !diff.NonNegative() {
			// metadata of transfer in the request ir wrong
			err = fmt.Errorf("mismatch between request metadata and outputs for address %s", addr.String())
			return false
		}
		// there are some free tokens for the address, i.e.
		// there are more tokens in the outputs of the value transaction than tokens in the
		// metadata. The Vm has to do something about it, otherwise they will become inaccessible
		prop.freeTokensByAddress[addr] = diff
		return true
	})
	return err
}

func (prop *properties) SenderAddress() ledgerstate.Address {
	return &prop.senderAddress
}

func (prop *properties) IsState() bool {
	return prop.isState
}

func (prop *properties) IsOrigin() bool {
	return prop.isState
}

func (prop *properties) MustChainID() *coretypes.ChainID {
	if !prop.isState {
		panic("MustChainID: must be a state transaction")
	}
	return &prop.chainID
}

func (prop *properties) MustStateColor() *ledgerstate.Color {
	if !prop.isState {
		panic("MustStateColor: must be a state transaction")
	}
	return &prop.stateColor
}

func (prop *properties) String() string {
	ret := "---- TransactionEssence:\n"
	ret += fmt.Sprintf("   requests: %d\n", prop.numRequests)
	ret += fmt.Sprintf("   senderAddress: %s\n", prop.senderAddress.String())
	ret += fmt.Sprintf("   isState: %v\n   isOrigin: %v\n", prop.isState, prop.isOrigin)
	ret += fmt.Sprintf("   chainAddress: %s\n", prop.chainAddress.String())
	ret += fmt.Sprintf("   chainID: %s\n   stateColor: %s\n", prop.chainID.String(), prop.stateColor.String())
	ret += fmt.Sprintf("   timestamp: %d\n    stateHash: %s\n", prop.timestamp, prop.stateHash.String())
	ret += fmt.Sprintf("   data payload size: %d\n", prop.dataPayloadSize)
	return ret
}
