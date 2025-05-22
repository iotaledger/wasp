// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/tracers"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// ChainBackend provides access to the underlying ISC chain.
type ChainBackend interface {
	EVMSendTransaction(tx *types.Transaction) error
	EVMCall(anchor *isc.StateAnchor, callMsg ethereum.CallMsg, l1Params *parameters.L1Params) ([]byte, error)
	EVMEstimateGas(anchor *isc.StateAnchor, callMsg ethereum.CallMsg, l1Params *parameters.L1Params) (uint64, error)
	EVMTrace(
		anchor *isc.StateAnchor,
		blockTime time.Time,
		iscRequestsInBlock []isc.Request,
		enforceGasBurned []vm.EnforceGasBurned,
		tracer *tracers.Tracer,
		l1Params *parameters.L1Params,
	) error
	FeePolicy(blockIndex uint32) (*gas.FeePolicy, error)
	ISCChainID() *isc.ChainID
	ISCCallView(chainState state.State, msg isc.Message) (isc.CallArguments, error)
	ISCLatestState() (*isc.StateAnchor, state.State, error)
	ISCStateByBlockIndex(blockIndex uint32) (state.State, error)
	ISCStateByTrieRoot(trieRoot trie.Hash) (state.State, error)
	TakeSnapshot() (int, error)
	RevertToSnapshot(int) error
}
