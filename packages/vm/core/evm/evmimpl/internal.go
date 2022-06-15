// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type blockContext struct {
	emu      *emulator.EVMEmulator
	txs      []*types.Transaction
	receipts []*types.Receipt
}

// getBlockContext creates a new emulator instance if this is the first call to applyTransaction
// in the ISC block; otherwise it returns the previously created instance. The purpose is to
// create a single Ethereum block for each ISC block.
func getBlockContext(ctx iscp.Sandbox) *blockContext {
	bctx := ctx.Privileged().BlockContext(
		func(ctx iscp.Sandbox) interface{} { return &blockContext{emu: createEmulator(ctx)} },
		func(bctx interface{}) { bctx.(*blockContext).close() },
	)
	return bctx.(*blockContext)
}

func (bctx *blockContext) close() {
	// count txs where status = success (which are already stored in the pending block)
	txCount := uint(0)
	for i := range bctx.txs {
		if bctx.receipts[i].Status == types.ReceiptStatusSuccessful {
			txCount++
		}
	}

	// failed txs were not stored in the pending block -- store them now
	for i := range bctx.txs {
		if bctx.receipts[i].Status != types.ReceiptStatusSuccessful {
			bctx.receipts[i].TransactionIndex = txCount
			bctx.emu.BlockchainDB().AddTransaction(bctx.txs[i], bctx.receipts[i])
			txCount++
		}
	}

	bctx.emu.MintBlock()
}

func createEmulator(ctx iscp.Sandbox) *emulator.EVMEmulator {
	return emulator.NewEVMEmulator(
		evmStateSubrealm(ctx.State()),
		timestamp(ctx),
		newISCContract(ctx),
		getBalanceFunc(ctx),
	)
}

func createEmulatorR(ctx iscp.SandboxView) *emulator.EVMEmulator {
	return emulator.NewEVMEmulator(
		evmStateSubrealm(buffered.NewBufferedKVStoreAccess(ctx.State())),
		timestamp(ctx),
		newISCContractView(ctx),
		getBalanceFunc(ctx),
	)
}

// timestamp returns the current timestamp in seconds since epoch
func timestamp(ctx iscp.SandboxBase) uint64 {
	return uint64(ctx.Timestamp().Unix())
}

func result(value []byte) dict.Dict {
	if value == nil {
		return nil
	}
	return dict.Dict{evm.FieldResult: value}
}

func blockResult(emu *emulator.EVMEmulator, block *types.Block) dict.Dict {
	if block == nil {
		return nil
	}
	return result(evmtypes.EncodeBlock(block))
}

func txResult(emu *emulator.EVMEmulator, tx *types.Transaction) dict.Dict {
	if tx == nil {
		return nil
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
	}
}

func txCountResult(emu *emulator.EVMEmulator, block *types.Block) dict.Dict {
	if block == nil {
		return nil
	}
	n := uint64(0)
	if block.NumberU64() != 0 {
		n = 1
	}
	return result(codec.EncodeUint64(n))
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
		assert.NewAssert(ctx.Log()).Requiref(allowPrevious, "unsupported operation")
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

func paramBlockNumberOrHashAsNumber(ctx iscp.SandboxView, emu *emulator.EVMEmulator, allowPrevious bool) uint64 {
	if ctx.Params().MustHas(evm.FieldBlockHash) {
		a := assert.NewAssert(ctx.Log())
		blockHash := common.BytesToHash(ctx.Params().MustGet(evm.FieldBlockHash))
		header := emu.BlockchainDB().GetHeaderByHash(blockHash)
		a.Requiref(header != nil, "block not found")
		return requireLatestBlock(ctx, emu, allowPrevious, header.Number.Uint64())
	}
	return paramBlockNumber(ctx, emu, allowPrevious)
}

func getBalanceFunc(ctx iscp.SandboxBase) emulator.BalanceFunc {
	res := ctx.CallView(
		governance.Contract.Hname(),
		governance.ViewGetFeePolicy.Hname(),
		nil,
	)
	feePolicy, err := gas.FeePolicyFromBytes(res.MustGet(governance.ParamFeePolicyBytes))
	ctx.RequireNoError(err)
	if feePolicy.GasFeeTokenID == nil {
		return func(addr common.Address) *big.Int {
			res := ctx.CallView(
				accounts.Contract.Hname(),
				accounts.ViewBalanceIotas.Hname(),
				dict.Dict{accounts.ParamAgentID: iscp.NewEthereumAddressAgentID(addr).Bytes()},
			)
			return new(big.Int).SetUint64(codec.MustDecodeUint64(res.MustGet(accounts.ParamBalance), 0))
		}
	} else {
		return func(addr common.Address) *big.Int {
			res := ctx.CallView(
				accounts.Contract.Hname(),
				accounts.ViewBalanceNativeToken.Hname(),
				dict.Dict{
					accounts.ParamAgentID:       iscp.NewEthereumAddressAgentID(addr).Bytes(),
					accounts.ParamNativeTokenID: feePolicy.GasFeeTokenID[:],
				},
			)
			return new(big.Int).SetBytes(res.MustGet(accounts.ParamBalance))
		}
	}
}
