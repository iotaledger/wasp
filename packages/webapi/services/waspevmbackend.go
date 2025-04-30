// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package services defines the services for the evm backend in the webapi
package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/tracers"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chainutil"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// WaspEVMBackend is the implementation of [jsonrpc.ChainBackend] for the production environment.
type WaspEVMBackend struct {
	chain      chain.Chain
	nodePubKey *cryptolib.PublicKey
}

var _ jsonrpc.ChainBackend = &WaspEVMBackend{}

func NewWaspEVMBackend(ch chain.Chain, nodePubKey *cryptolib.PublicKey) *WaspEVMBackend {
	return &WaspEVMBackend{
		chain:      ch,
		nodePubKey: nodePubKey,
	}
}

func (b *WaspEVMBackend) FeePolicy(blockIndex uint32) (*gas.FeePolicy, error) {
	state, err := b.ISCStateByBlockIndex(blockIndex)
	if err != nil {
		return nil, err
	}
	ret, err := b.ISCCallView(state, governance.ViewGetFeePolicy.Message())
	if err != nil {
		return nil, err
	}
	return governance.ViewGetFeePolicy.DecodeOutput(ret)
}

func (b *WaspEVMBackend) EVMSendTransaction(tx *types.Transaction) error {
	// Ensure the transaction has more gas than the basic Ethereum tx fee.
	intrinsicGas, err := core.IntrinsicGas(tx.Data(), tx.AccessList(), nil, tx.To() == nil, true, true, true)
	if err != nil {
		return err
	}
	if tx.Gas() < intrinsicGas {
		return core.ErrIntrinsicGas
	}

	req, err := isc.NewEVMOffLedgerTxRequest(b.chain.ID(), tx)
	if err != nil {
		return err
	}
	b.chain.Log().LogDebugf("EVMSendTransaction, evm.tx.nonce=%v, evm.tx.hash=%v => isc.req.id=%v", tx.Nonce(), tx.Hash().Hex(), req.ID())
	if err := b.chain.ReceiveOffLedgerRequest(req, b.nodePubKey); err != nil {
		return fmt.Errorf("tx not added to the mempool: %v", err.Error())
	}

	return nil
}

func (b *WaspEVMBackend) EVMCall(anchor *isc.StateAnchor, callMsg ethereum.CallMsg, l1Params *parameters.L1Params) ([]byte, error) {
	return chainutil.EVMCall(
		anchor,
		l1Params,
		b.chain.Store(),
		b.chain.Processors(),
		b.chain.Log(),
		callMsg,
	)
}

func (b *WaspEVMBackend) EVMEstimateGas(anchor *isc.StateAnchor, callMsg ethereum.CallMsg, l1Params *parameters.L1Params) (uint64, error) {
	return chainutil.EVMEstimateGas(
		anchor,
		l1Params,
		b.chain.Store(),
		b.chain.Processors(),
		b.chain.Log(),
		callMsg,
	)
}

func (b *WaspEVMBackend) EVMTrace(
	anchor *isc.StateAnchor,
	blockTime time.Time,
	iscRequestsInBlock []isc.Request,
	enforceGasBurned []vm.EnforceGasBurned,
	tracer *tracers.Tracer,
	l1Params *parameters.L1Params,
) error {
	return chainutil.EVMTrace(
		anchor,
		l1Params,
		b.chain.Store(),
		b.chain.Processors(),
		b.chain.Log(),
		blockTime,
		iscRequestsInBlock,
		enforceGasBurned,
		tracer,
	)
}

func (b *WaspEVMBackend) ISCCallView(chainState state.State, msg isc.Message) (isc.CallArguments, error) {
	latestAnchor, err := b.ISCLatestAnchor()
	if err != nil {
		return nil, err
	}
	return chainutil.CallView(
		latestAnchor.ChainID(),
		chainState,
		b.chain.Processors(),
		b.chain.Log(),
		msg,
	)
}

func (b *WaspEVMBackend) ISCLatestAnchor() (*isc.StateAnchor, error) {
	latestAnchor, err := b.chain.LatestAnchor(chain.ActiveOrCommittedState)
	if err != nil {
		return nil, fmt.Errorf("could not get latest Anchor: %w", err)
	}
	return latestAnchor, nil
}

func (b *WaspEVMBackend) ISCLatestState() (state.State, error) {
	latestState, err := b.chain.LatestState(chain.ActiveOrCommittedState)
	if err != nil {
		return nil, fmt.Errorf("couldn't get latest block index: %w", err)
	}
	return latestState, nil
}

func (b *WaspEVMBackend) ISCStateByBlockIndex(blockIndex uint32) (state.State, error) {
	latestState, err := b.chain.LatestState(chain.ActiveOrCommittedState)
	if err != nil {
		return nil, fmt.Errorf("couldn't get latest state: %s", err.Error())
	}
	if latestState.BlockIndex() == blockIndex {
		return latestState, nil
	}
	return b.chain.Store().StateByIndex(blockIndex)
}

func (b *WaspEVMBackend) ISCStateByTrieRoot(trieRoot trie.Hash) (state.State, error) {
	return b.chain.Store().StateByTrieRoot(trieRoot)
}

func (b *WaspEVMBackend) ISCChainID() *isc.ChainID {
	chID := b.chain.ID()
	return &chID
}

var errNotImplemented = errors.New("method not implemented")

func (*WaspEVMBackend) RevertToSnapshot(int) error {
	return errNotImplemented
}

func (*WaspEVMBackend) TakeSnapshot() (int, error) {
	return 0, errNotImplemented
}
