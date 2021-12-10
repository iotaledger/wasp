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
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// getEmulatorInBlockContext creates a new emulator instance if this is the first call to applyTransaction
// in the ISCP block; otherwise it returns the previously created instance. The purpose is to
// create a single Ethereum block for each ISCP block.
func getEmulatorInBlockContext(ctx iscp.Sandbox) *emulator.EVMEmulator {
	bctx := ctx.BlockContext(
		func(ctx iscp.Sandbox) interface{} { return createEmulator(ctx) },
		func(bctx interface{}) { bctx.(*emulator.EVMEmulator).MintBlock() },
	)
	return bctx.(*emulator.EVMEmulator)
}

func createEmulator(ctx iscp.Sandbox) *emulator.EVMEmulator {
	return emulator.NewEVMEmulator(evminternal.EVMStateSubrealm(ctx.State()), timestamp(ctx), &iscpBackend{ctx})
}

func createEmulatorR(ctx iscp.SandboxView) *emulator.EVMEmulator {
	return emulator.NewEVMEmulator(evminternal.EVMStateSubrealm(buffered.NewBufferedKVStoreAccess(ctx.State())), timestamp(ctx), &iscpBackendR{ctx})
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
	blockNumber, ok := bc.GetBlockNumberByTxHash(tx.Hash())
	if !ok {
		panic("cannot find block number of tx")
	}
	return dict.Dict{
		evm.FieldTransaction: evmtypes.EncodeTransaction(tx),
		evm.FieldBlockHash:   bc.GetBlockHashByBlockNumber(blockNumber).Bytes(),
		evm.FieldBlockNumber: codec.EncodeUint64(blockNumber),
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

	bc := emu.BlockchainDB()
	blockNumber, ok := bc.GetBlockNumberByBlockHash(blockHash)
	if !ok {
		return emu, nil
	}
	return emu, bc.GetTransactionByBlockNumberAndIndex(blockNumber, uint32(index))
}

func transactionByBlockNumberAndIndex(ctx iscp.SandboxView) (*emulator.EVMEmulator, *types.Transaction) {
	emu := createEmulatorR(ctx)
	blockNumber := paramBlockNumber(ctx, emu, true)

	a := assert.NewAssert(ctx.Log())
	index, err := codec.DecodeUint64(ctx.Params().MustGet(evm.FieldTransactionIndex), 0)
	a.RequireNoError(err)

	return emu, emu.BlockchainDB().GetTransactionByBlockNumberAndIndex(blockNumber, uint32(index))
}

func requireLatestBlock(ctx iscp.SandboxView, emu *emulator.EVMEmulator, allowPrevious bool, blockNumber uint64) uint64 {
	current := emu.BlockchainDB().GetNumber()
	if blockNumber != current {
		assert.NewAssert(ctx.Log()).Require(allowPrevious, "unsupported operation")
	}
	return blockNumber
}

func paramBlockNumber(ctx iscp.SandboxView, emu *emulator.EVMEmulator, allowPrevious bool) uint64 {
	current := emu.BlockchainDB().GetNumber()
	if ctx.Params().MustHas(evm.FieldBlockNumber) {
		blockNumber := new(big.Int).SetBytes(ctx.Params().MustGet(evm.FieldBlockNumber))
		return requireLatestBlock(ctx, emu, allowPrevious, blockNumber.Uint64())
	}
	return current
}

func paramBlockNumberOrHashAsNumber(ctx iscp.SandboxView, emu *emulator.EVMEmulator, allowPrevious bool) uint64 { // nolint:unparam
	if ctx.Params().MustHas(evm.FieldBlockHash) {
		a := assert.NewAssert(ctx.Log())
		blockHash := common.BytesToHash(ctx.Params().MustGet(evm.FieldBlockHash))
		header := emu.BlockchainDB().GetHeaderByHash(blockHash)
		a.Require(header != nil, "block not found")
		return requireLatestBlock(ctx, emu, allowPrevious, header.Number.Uint64())
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
