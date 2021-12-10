// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmlight

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/contracts/native/evm/evminternal"
	"github.com/iotaledger/wasp/contracts/native/evm/evmlight/emulator"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
)

// keyEVMState is the subrealm prefix for the EVM state
// If changed, make sure it does not collide with evm management keys
const keyEVMState = "s"

func evmStateSubrealm(state kv.KVStore) kv.KVStore {
	return subrealm.New(state, keyEVMState)
}

func createEmulator(ctx iscp.Sandbox) *emulator.EVMEmulator {
	return emulator.NewEVMEmulator(evmStateSubrealm(ctx.State()), timestamp(ctx), &iscpBackend{ctx})
}

func createEmulatorR(ctx iscp.SandboxView) *emulator.EVMEmulator {
	return emulator.NewEVMEmulator(evmStateSubrealm(buffered.NewBufferedKVStoreAccess(ctx.State())), timestamp(ctx), &iscpBackendR{ctx})
}

// timestamp returns the current timestamp in seconds since epoch
func timestamp(ctx iscp.SandboxBase) uint64 {
	tsNano := time.Duration(ctx.GetTimestamp()) * time.Nanosecond
	return uint64(tsNano / time.Second)
}

func blockResult(emu *emulator.EVMEmulator, block *types.Block) (dict.Dict, error) {
	if block == nil {
		return nil, nil
	}
	return evminternal.Result(evmtypes.EncodeBlock(block)), nil
}

func txResult(emu *emulator.EVMEmulator, tx *types.Transaction) (dict.Dict, error) {
	if tx == nil {
		return nil, nil
	}
	bc := emu.BlockchainDB()
	blockNumber := bc.GetBlockNumberByTxHash(tx.Hash())
	return dict.Dict{
		evm.FieldTransaction: evmtypes.EncodeTransaction(tx),
		evm.FieldBlockHash:   bc.GetBlockHashByBlockNumber(blockNumber).Bytes(),
		evm.FieldBlockNumber: codec.EncodeUint64(blockNumber.Uint64()),
	}, nil
}

func txCountResult(emu *emulator.EVMEmulator, block *types.Block) (dict.Dict, error) {
	if block == nil {
		return nil, nil
	}
	n := uint64(0)
	if block.NumberU64() != 0 {
		n = 1
	}
	return evminternal.Result(codec.EncodeUint64(n)), nil
}

func blockByNumber(ctx iscp.SandboxView) (*emulator.EVMEmulator, *types.Block) {
	emu := createEmulatorR(ctx)
	blockNumber := paramBlockNumber(ctx, emu, true)
	return emu, emu.BlockchainDB().GetBlockByNumber(blockNumber)
}

func blockByHash(ctx iscp.SandboxView) (*emulator.EVMEmulator, *types.Block) {
	emu := createEmulatorR(ctx)
	hash := common.BytesToHash(ctx.Params().MustGet(evm.FieldBlockHash))
	return emu, emu.BlockchainDB().GetBlockByHash(hash)
}

func transactionByHash(ctx iscp.SandboxView) (*emulator.EVMEmulator, *types.Transaction) {
	emu := createEmulatorR(ctx)
	txHash := common.BytesToHash(ctx.Params().MustGet(evm.FieldTransactionHash))
	return emu, emu.BlockchainDB().GetTransactionByHash(txHash)
}

func transactionByBlockHashAndIndex(ctx iscp.SandboxView) (*emulator.EVMEmulator, *types.Transaction) {
	emu := createEmulatorR(ctx)
	blockHash := common.BytesToHash(ctx.Params().MustGet(evm.FieldBlockHash))

	a := assert.NewAssert(ctx.Log())
	index, err := codec.DecodeUint64(ctx.Params().MustGet(evm.FieldTransactionIndex), 0)
	a.RequireNoError(err)
	// all blocks contain at most 1 tx
	if index > 0 {
		return emu, nil
	}

	bc := emu.BlockchainDB()
	blockNumber := bc.GetBlockNumberByBlockHash(blockHash)
	if blockNumber == nil {
		return emu, nil
	}
	return emu, bc.GetTransactionByBlockNumber(blockNumber)
}

func transactionByBlockNumberAndIndex(ctx iscp.SandboxView) (*emulator.EVMEmulator, *types.Transaction) {
	emu := createEmulatorR(ctx)
	blockNumber := paramBlockNumber(ctx, emu, true)

	a := assert.NewAssert(ctx.Log())
	index, err := codec.DecodeUint64(ctx.Params().MustGet(evm.FieldTransactionIndex), 0)
	a.RequireNoError(err)
	// all blocks contain at most 1 tx
	if index > 0 {
		return emu, nil
	}

	return emu, emu.BlockchainDB().GetTransactionByBlockNumber(blockNumber)
}

func requireLatestBlock(ctx iscp.SandboxView, emu *emulator.EVMEmulator, allowPrevious bool, blockNumber *big.Int) *big.Int {
	current := emu.BlockchainDB().GetNumber()
	if blockNumber.Cmp(current) != 0 {
		assert.NewAssert(ctx.Log()).Require(allowPrevious, "unsupported operation")
	}
	return blockNumber
}

func paramBlockNumber(ctx iscp.SandboxView, emu *emulator.EVMEmulator, allowPrevious bool) *big.Int {
	current := emu.BlockchainDB().GetNumber()
	if ctx.Params().MustHas(evm.FieldBlockNumber) {
		blockNumber := new(big.Int).SetBytes(ctx.Params().MustGet(evm.FieldBlockNumber))
		return requireLatestBlock(ctx, emu, allowPrevious, blockNumber)
	}
	return current
}

func paramBlockNumberOrHashAsNumber(ctx iscp.SandboxView, emu *emulator.EVMEmulator, allowPrevious bool) *big.Int { // nolint:unparam
	if ctx.Params().MustHas(evm.FieldBlockHash) {
		a := assert.NewAssert(ctx.Log())
		blockHash := common.BytesToHash(ctx.Params().MustGet(evm.FieldBlockHash))
		header := emu.BlockchainDB().GetHeaderByHash(blockHash)
		a.Require(header != nil, "block not found")
		return requireLatestBlock(ctx, emu, allowPrevious, header.Number)
	}
	return paramBlockNumber(ctx, emu, allowPrevious)
}

type iscpBackend struct {
	ctx iscp.Sandbox
}

var _ vm.ISCPBackend = &iscpBackend{}

func (i *iscpBackend) Event(s string) {
	i.ctx.Event(s)
}

func (i *iscpBackend) Entropy() [32]byte {
	return i.ctx.GetEntropy()
}

type iscpBackendR struct {
	ctx iscp.SandboxView
}

var _ vm.ISCPBackend = &iscpBackendR{}

func (i *iscpBackendR) Event(s string)    { panic("should not happen") }
func (i *iscpBackendR) Entropy() [32]byte { panic("should not happen") }
