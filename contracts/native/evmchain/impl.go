// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmchain

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func emulator(state kv.KVStore) *evm.EVMEmulator {
	return evm.NewEVMEmulator(rawdb.NewDatabase(evm.NewKVAdapter(state)))
}

func emulatorR(state kv.KVStoreReader) *evm.EVMEmulator {
	return evm.NewEVMEmulator(rawdb.NewDatabase(evm.NewKVAdapter(buffered.NewBufferedKVStore(state))))
}

func initialize(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	genesisAlloc, err := DecodeGenesisAlloc(ctx.Params().MustGet(FieldGenesisAlloc))
	a.RequireNoError(err)
	evm.InitGenesis(rawdb.NewDatabase(evm.NewKVAdapter(ctx.State())), genesisAlloc)
	return nil, nil
}

func applyTransaction(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())

	tx := &types.Transaction{}
	err := tx.UnmarshalBinary(ctx.Params().MustGet(FieldTransactionData))
	a.RequireNoError(err)

	emu := emulator(ctx.State())
	defer emu.Close()

	err = emu.SendTransaction(tx)
	a.RequireNoError(err)
	emu.Commit()

	return nil, nil
}

func getBalance(ctx coretypes.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())

	addr := common.BytesToAddress(ctx.Params().MustGet(FieldAddress))

	var blockNumber *big.Int
	if ctx.Params().MustHas(FieldBlockNumber) {
		blockNumber = new(big.Int).SetBytes(ctx.Params().MustGet(FieldBlockNumber))
	}

	emu := emulatorR(ctx.State())
	defer emu.Close()

	bal, err := emu.BalanceAt(addr, blockNumber)
	a.RequireNoError(err)

	ret := dict.New()
	ret.Set(FieldBalance, bal.Bytes())
	return ret, nil
}

func getBlockNumber(ctx coretypes.SandboxView) (dict.Dict, error) {
	emu := emulatorR(ctx.State())
	defer emu.Close()

	n := emu.Blockchain().CurrentBlock().Number()

	ret := dict.New()
	ret.Set(FieldResult, n.Bytes())
	return ret, nil
}

func getBlockByNumber(ctx coretypes.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())

	var blockNumber *big.Int
	if ctx.Params().MustHas(FieldBlockNumber) {
		blockNumber = new(big.Int).SetBytes(ctx.Params().MustGet(FieldBlockNumber))
	}

	emu := emulatorR(ctx.State())
	defer emu.Close()

	block, err := emu.BlockByNumber(blockNumber)
	a.RequireNoError(err)

	ret := dict.New()
	if block != nil {
		ret.Set(FieldResult, EncodeBlock(block))
	}
	return ret, nil
}

func getReceipt(ctx coretypes.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())

	txHash := common.BytesToHash(ctx.Params().MustGet(FieldTransactionHash))

	emu := emulatorR(ctx.State())
	defer emu.Close()

	receipt, err := emu.TransactionReceipt(txHash)
	a.RequireNoError(err)

	receiptBytes, err := rlp.EncodeToBytes(receipt)
	a.RequireNoError(err)

	ret := dict.New()
	ret.Set(FieldResult, receiptBytes)
	return ret, nil
}

func getNonce(ctx coretypes.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())

	addr := common.BytesToAddress(ctx.Params().MustGet(FieldAddress))

	var blockNumber *big.Int
	if ctx.Params().MustHas(FieldBlockNumber) {
		blockNumber = new(big.Int).SetBytes(ctx.Params().MustGet(FieldBlockNumber))
	}

	emu := emulatorR(ctx.State())
	defer emu.Close()

	nonce, err := emu.NonceAt(addr, blockNumber)
	a.RequireNoError(err)

	ret := dict.New()
	ret.Set(FieldResult, codec.EncodeUint64(nonce))
	return ret, nil
}

func getCode(ctx coretypes.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	addr := common.BytesToAddress(ctx.Params().MustGet(FieldAddress))

	emu := emulatorR(ctx.State())
	defer emu.Close()

	code, err := emu.CodeAt(addr, nil)
	a.RequireNoError(err)

	ret := dict.New()
	ret.Set(FieldResult, code)
	return ret, nil
}

func callView(ctx coretypes.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())

	contractAddress := common.BytesToAddress(ctx.Params().MustGet(FieldAddress))
	callArguments := ctx.Params().MustGet(FieldCallArguments)

	emu := emulatorR(ctx.State())
	defer emu.Close()

	res, err := emu.CallContract(ethereum.CallMsg{
		To:   &contractAddress,
		Data: callArguments,
	}, nil)
	a.RequireNoError(err)

	ret := dict.New()
	ret.Set(FieldResult, res)
	return ret, nil
}
