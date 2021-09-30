// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmchain

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

func evmStateSubrealm(state kv.KVStore) kv.KVStore {
	return subrealm.New(state, keyEVMState)
}

func createEmulator(ctx iscp.Sandbox) *evm.EVMEmulator {
	return evm.NewEVMEmulator(evmStateSubrealm(ctx.State()), timestamp(ctx), &iscpBackend{ctx})
}

func createEmulatorR(ctx iscp.SandboxView) *evm.EVMEmulator {
	return evm.NewEVMEmulator(evmStateSubrealm(buffered.NewBufferedKVStoreAccess(ctx.State())), timestamp(ctx), &iscpBackendR{ctx})
}

// timestamp returns the current timestamp in seconds since epoch
func timestamp(ctx iscp.SandboxBase) uint64 {
	tsNano := time.Duration(ctx.GetTimestamp()) * time.Nanosecond
	return uint64(tsNano / time.Second)
}

func result(value []byte) dict.Dict {
	if value == nil {
		return nil
	}
	return dict.Dict{FieldResult: value}
}

func blockResult(emu *evm.EVMEmulator, block *types.Block) (dict.Dict, error) {
	if block == nil {
		return nil, nil
	}
	return result(evmtypes.EncodeBlock(block)), nil
}

func txResult(emu *evm.EVMEmulator, tx *types.Transaction) (dict.Dict, error) {
	if tx == nil {
		return nil, nil
	}
	bc := emu.BlockchainDB()
	blockNumber := bc.GetBlockNumberByTxHash(tx.Hash())
	return dict.Dict{
		FieldTransaction: evmtypes.EncodeTransaction(tx),
		FieldBlockHash:   bc.GetBlockHashByBlockNumber(blockNumber).Bytes(),
		FieldBlockNumber: codec.EncodeUint64(blockNumber.Uint64()),
	}, nil
}

func txCountResult(emu *evm.EVMEmulator, block *types.Block) (dict.Dict, error) {
	if block == nil {
		return nil, nil
	}
	n := uint64(0)
	if block.NumberU64() != 0 {
		n = 1
	}
	return result(codec.EncodeUint64(n)), nil
}

func blockByNumber(ctx iscp.SandboxView) (*evm.EVMEmulator, *types.Block) {
	emu := createEmulatorR(ctx)
	blockNumber := paramBlockNumber(ctx, emu, true)
	return emu, emu.BlockchainDB().GetBlockByNumber(blockNumber)
}

func blockByHash(ctx iscp.SandboxView) (*evm.EVMEmulator, *types.Block) {
	emu := createEmulatorR(ctx)
	hash := common.BytesToHash(ctx.Params().MustGet(FieldBlockHash))
	return emu, emu.BlockchainDB().GetBlockByHash(hash)
}

func transactionByHash(ctx iscp.SandboxView) (*evm.EVMEmulator, *types.Transaction) {
	emu := createEmulatorR(ctx)
	txHash := common.BytesToHash(ctx.Params().MustGet(FieldTransactionHash))
	return emu, emu.BlockchainDB().GetTransactionByHash(txHash)
}

func transactionByBlockHash(ctx iscp.SandboxView) (*evm.EVMEmulator, *types.Transaction) {
	emu := createEmulatorR(ctx)
	blockHash := common.BytesToHash(ctx.Params().MustGet(FieldBlockHash))
	bc := emu.BlockchainDB()
	blockNumber := bc.GetBlockNumberByBlockHash(blockHash)
	if blockNumber == nil {
		return emu, nil
	}
	return emu, bc.GetTransactionByBlockNumber(blockNumber)
}

func transactionByBlockNumber(ctx iscp.SandboxView) (*evm.EVMEmulator, *types.Transaction) {
	emu := createEmulatorR(ctx)
	blockNumber := paramBlockNumber(ctx, emu, true)
	return emu, emu.BlockchainDB().GetTransactionByBlockNumber(blockNumber)
}

func paramBlockNumber(ctx iscp.SandboxView, emu *evm.EVMEmulator, allowPrevious bool) *big.Int {
	current := emu.BlockchainDB().GetNumber()
	if ctx.Params().MustHas(FieldBlockNumber) {
		blockNumber := new(big.Int).SetBytes(ctx.Params().MustGet(FieldBlockNumber))
		if blockNumber.Cmp(current) != 0 {
			assert.NewAssert(ctx.Log()).Require(allowPrevious, "unsupported operation")
		}
		return blockNumber
	}
	return current
}

func getFeeColor(ctx iscp.Sandbox) colored.Color {
	a := assert.NewAssert(ctx.Log())

	// call root contract view to get the feecolor
	feeInfo, err := ctx.Call(
		governance.Contract.Hname(),
		governance.FuncGetFeeInfo.Hname(),
		dict.Dict{governance.ParamHname: Contract.Hname().Bytes()},
		nil,
	)
	a.RequireNoError(err)
	feeColor, err := codec.DecodeColor(feeInfo.MustGet(governance.ParamFeeColor))
	a.RequireNoError(err)
	return feeColor
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
