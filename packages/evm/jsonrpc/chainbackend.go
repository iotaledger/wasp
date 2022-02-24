// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type ChainBackend interface {
	EstimateGasOnLedger(scName string, funName string, transfer *iscp.Assets, args dict.Dict) (uint64, *iscp.Assets, error)
	PostOnLedgerRequest(scName string, funName string, transfer *iscp.Assets, args dict.Dict, gasBudget uint64) error
	EstimateGasOffLedger(scName string, funName string, args dict.Dict) (uint64, *iscp.Assets, error)
	PostOffLedgerRequest(scName string, funName string, args dict.Dict, gasBudget uint64) error
	CallView(scName string, funName string, args dict.Dict) (dict.Dict, error)
	Signer() *cryptolib.KeyPair
}
