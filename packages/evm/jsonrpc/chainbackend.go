// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/iscp/color"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type ChainBackend interface {
	PostOnLedgerRequest(scName string, funName string, transfer color.Balances, args dict.Dict) error
	PostOffLedgerRequest(scName string, funName string, transfer color.Balances, args dict.Dict) error
	CallView(scName string, funName string, args dict.Dict) (dict.Dict, error)
	Signer() *ed25519.KeyPair
}
