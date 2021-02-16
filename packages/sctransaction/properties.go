package sctransaction

import (
	"errors"
	"fmt"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/hashing"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/txutil"
)

// Properties represents result of analysis and semantic check of the SC transaction
// SC transaction is a value transaction with successfully parsed data payload
type Properties struct {
	// transaction ID
	txid valuetransaction.ID
	// senderAddress of the SC transaction. It is the only
	senderAddress address.Address
	// is it state transaction (== does it contain valid stateSection)
	isState bool
	// if isState == true it states if it is the origin transaction, otherwise uninterpreted
	isOrigin bool
	// if isState == true: chainID
	chainID coretypes.ChainID
	// chainAddress == chainID
	chainAddress address.Address
	// if isState == true: smart contract color
	stateColor balance.Color
	// timestamp from state section
	timestamp int64
	// stateHash from state section
	stateHash hashing.HashValue
	// number of requests
	numRequests int
	// data payload len
	dataPayloadSize uint32
	// number of minted tokens to any address - number of requests
	numFreeMintedTokens int64
	// free tokens: tokens with output to chain address - tokens transferred by requests - request tokens - chain token
	// In most cases it is empty, because all tokens should be transferred with requests
	// Free tokens normally should be returned to the sender
	freeTokensByAddress map[address.Address]coretypes.ColoredBalances
}

func (tx *Transaction) calcProperties() (*Properties, error) {
	ret := &Properties{
		txid:                tx.ID(),
		dataPayloadSize:     tx.DataPayloadSize(),
		freeTokensByAddress: make(map[address.Address]coretypes.ColoredBalances),
	}

	if !tx.SignaturesValid() {
		return nil, fmt.Errorf("invalid signatures")
	}
	if len(tx.Signatures()) > 1 {
		return nil, fmt.Errorf("number of signatures > 1")
	}
	if err := ret.analyzeSender(tx); err != nil {
		return nil, err
	}
	if err := ret.analyzeStateBlock(tx); err != nil {
		return nil, err
	}
	if err := ret.analyzeRequestBlocks(tx); err != nil {
		return nil, err
	}
	ret.calcNumMinted(tx)
	return ret, nil
}

func (prop *Properties) calcNumMinted(tx *Transaction) {
	prop.numFreeMintedTokens = 0
	tx.Transaction.Outputs().ForEach(func(addr address.Address, bals []*balance.Balance) bool {
		prop.numFreeMintedTokens += txutil.BalanceOfColor(bals, balance.ColorNew)
		return true
	})
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

func (prop *Properties) analyzeStateBlock(tx *Transaction) error {
	stateSection, ok := tx.State()
	prop.isState = ok
	if !ok {
		return nil
	}

	prop.timestamp = stateSection.timestamp
	prop.stateHash = stateSection.stateHash

	var err error

	prop.isOrigin = stateSection.Color() == balance.ColorNew
	sectionColor := stateSection.Color()
	if sectionColor == balance.ColorIOTA {
		return fmt.Errorf("state section color can't be IOTAColor")
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
		diff := cbalances.NewFromBalances(bals).Diff(cbalances.NewFromMap(m))
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
			err = fmt.Errorf("wrong number of minted tokens in the output to the address %s", addr.String())
			return false
		}
		if !diff.NonNegative() {
			err = fmt.Errorf("mismatch between request metadata and outputs for address %s", addr.String())
			return false
		}
		// there are some free tokens for the address
		prop.freeTokensByAddress[addr] = diff
		return true
	})
	return err
	// TODO free minted tokens
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

func (prop *Properties) ChainAddress() address.Address {
	return prop.chainAddress
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
	return prop.numFreeMintedTokens - int64(prop.numRequests)
}

func (prop *Properties) FreeTokensForAddress(addr address.Address) coretypes.ColoredBalances {
	if ret, ok := prop.freeTokensByAddress[addr]; ok {
		return ret
	}
	return cbalances.Nil
}

func (prop *Properties) String() string {
	ret := "---- Transaction:\n"
	ret += fmt.Sprintf("   requests: %d\n", prop.numRequests)
	ret += fmt.Sprintf("   senderAddress: %s\n", prop.senderAddress.String())
	ret += fmt.Sprintf("   isState: %v\n   isOrigin: %v\n", prop.isState, prop.isOrigin)
	ret += fmt.Sprintf("   chainAddress: %s\n", prop.chainAddress.String())
	ret += fmt.Sprintf("   chainID: %s\n   stateColor: %s\n", prop.chainID.String(), prop.stateColor.String())
	ret += fmt.Sprintf("   timestamp: %d\n    stateHash: %s\n", prop.timestamp, prop.stateHash.String())
	ret += fmt.Sprintf("   numMinted: %d\n", prop.numFreeMintedTokens)
	ret += fmt.Sprintf("   data payload size: %d\n", prop.dataPayloadSize)
	return ret
}
