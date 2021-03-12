// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
)

// SCTransactionProperties defines interface to the properties of
// syntactically and semantically valid smart contract transaction.
type SCTransactionProperties interface {
	// address of the transaction sender. It may be address of the wallet or chain address
	SenderAddress() *address.Address
	// is it state transaction, i.e. transaction with the valid state section
	IsState() bool
	// is it origin transaction
	IsOrigin() bool
	// chain ID of the state section or panic if not a state transaction
	MustChainID() *ChainID
	// color of the state section or panic if not a state transaction
	MustStateColor() *balance.Color
	// number of minted tokens which are not request tokens
	NumFreeMintedTokens() int64
	// all tokens sent to the address but not included into the requests to that address
	// (normally 0. Needed for fallback processing otherwise those free tokens will be unaccounted
	// for and essentially lost
	FreeTokensForAddress(addr address.Address) ColoredBalances
	// string representation
	String() string
}
