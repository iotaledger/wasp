// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/tracers"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
)

// ChainBackend provides access to the underlying ISC chain.
type ChainBackend interface {
	EVMSendTransaction(tx *types.Transaction) error
	EVMCall(aliasOutput *isc.AliasOutputWithID, callMsg ethereum.CallMsg) ([]byte, error)
	EVMEstimateGas(aliasOutput *isc.AliasOutputWithID, callMsg ethereum.CallMsg) (uint64, error)
	EVMTraceTransaction(aliasOutput *isc.AliasOutputWithID, blockTime time.Time, iscRequestsInBlock []isc.Request, txIndex uint64, tracer tracers.Tracer) error
	ISCChainID() *isc.ChainID
	ISCCallView(chainState state.State, scName string, funName string, args dict.Dict) (dict.Dict, error)
	ISCLatestAliasOutput() (*isc.AliasOutputWithID, error)
	ISCLatestState() state.State
	ISCStateByBlockIndex(blockIndex uint32) (state.State, error)
	ISCStateByTrieRoot(trieRoot trie.Hash) (state.State, error)
	BaseToken() *parameters.BaseToken
	TakeSnapshot() (int, error)
	RevertToSnapshot(int) error
}
