// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
)

// SCTransactionProperties defines interface to the properties of
// syntactically and semantically valid smart contract transaction.
type SCTransactionProperties interface {
	// address of the transaction sender. It may be address of the wallet or chain address
	SenderAddress() *ledgerstate.Address
	// is it state transaction, i.e. transaction with the valid state section
	IsState() bool
	// is it origin transaction
	IsOrigin() bool
	// chain ID of the state section or panic if not a state transaction
	MustChainID() *ChainID
	// color of the state section or panic if not a state transaction
	MustStateColor() *ledgerstate.Color
	// string representation
	String() string
}
