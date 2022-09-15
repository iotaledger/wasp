// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
)

type ChainBackend interface {
	EVMSendTransaction(tx *types.Transaction) error
	EVMEstimateGas(callMsg ethereum.CallMsg) (uint64, error)
	ISCCallView(scName string, funName string, args dict.Dict) (dict.Dict, error)
	BaseToken() *parameters.BaseToken
}
