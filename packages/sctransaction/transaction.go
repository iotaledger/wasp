package sctransaction

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
)

// region ParsedTransaction //////////////////////////////////////////////////////////////////

// ParsedTransaction is a wrapper of ledgerstate.Transaction. It provides additional validation
// and methods for ISCP. Represents a set of parsed outputs with target of specific chainID
type ParsedTransaction struct {
	*ledgerstate.Transaction
	receivingChainID coretypes.ChainID
	senderAddr       ledgerstate.Address
	chainOutput      *ledgerstate.AliasOutput
	stateHash        hashing.HashValue
	requests         []*RequestOnLedger
}

// Parse analyzes value transaction and parses its data
func Parse(tx *ledgerstate.Transaction, sender ledgerstate.Address, receivingChainID coretypes.ChainID) *ParsedTransaction {
	ret := &ParsedTransaction{
		Transaction:      tx,
		receivingChainID: receivingChainID,
		senderAddr:       sender,
		requests:         make([]*RequestOnLedger, 0),
	}
	for _, out := range tx.Essence().Outputs() {
		if !out.Address().Equals(receivingChainID.AsAddress()) {
			continue
		}
		switch o := out.(type) {
		case *ledgerstate.ExtendedLockedOutput:
			ret.requests = append(ret.requests, RequestOnLedgerFromOutput(o, sender))
		case *ledgerstate.AliasOutput:
			h, err := hashing.HashValueFromBytes(o.GetStateData())
			if err == nil {
				ret.stateHash = h
			}
			ret.chainOutput = o
		default:
			continue
		}
	}
	return ret
}

// ChainOutput return chain output or nil if the transaction is not a state anchor
func (tx *ParsedTransaction) ChainOutput() *ledgerstate.AliasOutput {
	return tx.chainOutput
}

func (tx *ParsedTransaction) SenderAddress() ledgerstate.Address {
	return tx.senderAddr
}

func (tx *ParsedTransaction) Requests() []*RequestOnLedger {
	return tx.requests
}

func (tx *ParsedTransaction) ReceivingChainID() coretypes.ChainID {
	return tx.receivingChainID
}

// endregion /////////////////////////////////////////////////////////////////
