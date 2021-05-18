// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmchain

import (
	"fmt"
	"math"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
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

///////// TODO contract owner - FieldEvmOwner
///////// TODO withdrawal

const gasPerIota int64 = 1000 ///////// TODO gas per iota - FieldGasPerIota

func applyTransaction(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())

	var txData []byte
	if ctx.Params().MustHas(FieldTransactionData) {
		txData = ctx.Params().MustGet(FieldTransactionData)
	} else {
		getBlobFieldParams := dict.Dict(map[kv.Key][]byte{
			blob.ParamHash:  ctx.Params().MustGet(FieldTransactionDataBlobHash),
			blob.ParamField: []byte(FieldTransactionData),
		})
		r, err := ctx.Call(blob.Interface.Hname(), coretypes.Hn(blob.FuncGetBlobField), getBlobFieldParams, nil)
		a.RequireNoError(err)
		txData = r.MustGet(blob.ParamBytes)
	}

	tx := &types.Transaction{}
	err := tx.UnmarshalBinary(txData)
	a.RequireNoError(err)

	emu := emulator(ctx.State())
	defer emu.Close()

	err = emu.SendTransaction(tx)
	a.RequireNoError(err)

	//solidifies the pending block
	emu.Commit()

	receipt, err := emu.TransactionReceipt(tx.Hash())
	a.RequireNoError(err)

	receiptBytes, err := rlp.EncodeToBytes(receipt)
	a.RequireNoError(err)

	transferedIotas, _ := ctx.IncomingTransfer().Get(ledgerstate.ColorIOTA) // TODO figure out if there is a nice way to get FeeColor from sandbox

	iotasGasFee := uint64(math.Ceil(float64(receipt.GasUsed) / float64(gasPerIota)))

	if transferedIotas < iotasGasFee {
		//not enough gas
		return nil, fmt.Errorf("not enough iotas to pay the gas fees. spent: %d iotas", transferedIotas)
	}

	// refund unspend gas to the senders on-chain account
	iotasGasRefund := transferedIotas - iotasGasFee
	params := codec.MakeDict(map[string]interface{}{
		accounts.ParamAgentID: ctx.Caller(),
	})
	_, err = ctx.Call(accounts.Interface.Hname(), coretypes.Hn(accounts.FuncDeposit), params, coretypes.NewTransferIotas(iotasGasRefund))
	a.RequireNoError(err)

	ret := dict.New()
	ret.Set(FieldResult, receiptBytes)
	ret.Set(FieldGasFee, codec.EncodeUint64(iotasGasFee))

	return ret, nil
}

func getBalance(ctx coretypes.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())

	addr := common.BytesToAddress(ctx.Params().MustGet(FieldAddress))

	emu := emulatorR(ctx.State())
	defer emu.Close()

	state, err := emu.Blockchain().State()
	a.RequireNoError(err)
	bal := state.GetBalance(addr)

	ret := dict.New()
	ret.Set(FieldBalance, bal.Bytes())
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

	emu := emulatorR(ctx.State())
	defer emu.Close()

	nonce, err := emu.PendingNonceAt(addr)
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
