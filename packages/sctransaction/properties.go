package sctransaction

import (
	"errors"
	"fmt"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/txutil"
)

type Properties struct {
	// TX ID
	txid valuetransaction.ID
	//
	numSignatures int
	// the only senderAddress of the SC transaction
	senderAddress address.Address
	// is it state transaction (== does it contain valid stateSection)
	isState bool
	// if isState == true: it states if it is the origin transaction
	isOrigin bool
	// if isState == true: chainID
	chainID coretypes.ChainID
	// chainAddress == chainID
	chainAddress address.Address
	// if isState == true: smart contract color
	stateColor      balance.Color
	numMintedTokens int64
	// number of requests
	numRequests int
	// data payload len
	dataPayloadSize uint32
}

func (tx *Transaction) calcProperties() (*Properties, error) {
	ret := &Properties{
		txid:            tx.ID(),
		dataPayloadSize: tx.DataPayloadSize(),
	}

	if tx.SignaturesValid() {
		ret.numSignatures = len(tx.Signatures())
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

func (prop *Properties) analyzeStateBlock(tx *Transaction) error {
	stateSection, ok := tx.State()
	prop.isState = ok
	if !ok {
		return nil
	}

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
		if v == 1 {
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
	var err error
	// validate all outputs w.r.t. request transfers
	tx.Transaction.Outputs().ForEach(func(addr address.Address, bals []*balance.Balance) bool {
		m, ok := reqTransfersByTargetChain[coretypes.ChainID(addr)]
		if !ok {
			// do not check outputs to outside addresses
			return true
		}
		outBalances := cbalances.NewFromBalances(bals)
		reqBalances := cbalances.NewFromMap(m)
		if addr != prop.chainAddress {
			// output to another chain
			// outputs and requests must be equal
			if !outBalances.Equal(reqBalances) {
				err = fmt.Errorf("mismatch between transfer data in request section and tx outputs")
				return false
			}
		}
		// output to the same chain
		// the request transfer must be subset of outputs and number of minted
		// request tokens must be equal to number of requests
		if outBalances.Balance(balance.ColorNew) != reqBalances.Balance(balance.ColorNew) {
			err = fmt.Errorf("wrong number of minted tokens in the output to the chain address")
			return false
		}
		// request balances must be contained in the chain balances
		if !outBalances.Includes(reqBalances) {
			err = fmt.Errorf("inconsisteny among request to self")
			return false
		}
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

// NumSignatures number of valid signatures
func (prop *Properties) NumSignatures() int {
	return prop.numSignatures
}

func (prop *Properties) String() string {
	ret := "---- Transaction:\n"
	ret += fmt.Sprintf("   txid: %s\n   num signatures: %d\n", prop.txid.String(), prop.numSignatures)
	ret += fmt.Sprintf("   requests: %d\n", prop.numRequests)
	ret += fmt.Sprintf("   senderAddress: %s\n", prop.senderAddress.String())
	ret += fmt.Sprintf("   isState: %v\n   isOrigin: %v\n", prop.isState, prop.isOrigin)
	ret += fmt.Sprintf("   chainAddress: %s\n", prop.chainAddress.String())
	ret += fmt.Sprintf("   chainID: %s\n   stateColor: %s\n", prop.chainID.String(), prop.stateColor.String())
	ret += fmt.Sprintf("   numMinted: %d\n", prop.numMintedTokens)
	ret += fmt.Sprintf("   data payload size: %d\n", prop.dataPayloadSize)
	return ret
}
