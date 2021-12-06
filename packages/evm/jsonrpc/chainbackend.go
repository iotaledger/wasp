// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	_ "github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type ChainBackend interface {
	PostOnLedgerRequest(scName string, funName string, transfer *iscp.Assets, args dict.Dict) error
	PostOffLedgerRequest(scName string, funName string, transfer *iscp.Assets, args dict.Dict) error
	CallView(scName string, funName string, args dict.Dict) (dict.Dict, error)
	Signer() *cryptolib.KeyPair
}
