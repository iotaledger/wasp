// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmchain

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

var (
	FaucetKey, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	FaucetAddress = crypto.PubkeyToAddress(FaucetKey.PublicKey)

	FaucetSupply = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))
	genesisAlloc = map[common.Address]core.GenesisAccount{
		FaucetAddress: {Balance: FaucetSupply},
	}
)

func emulator(state kv.KVStore) *evm.EVMEmulator {
	return evm.NewEVMEmulator(rawdb.NewDatabase(evm.NewKVAdapter(state)), genesisAlloc)
}

func emulatorR(state kv.KVStoreReader) *evm.EVMEmulator {
	return evm.NewEVMEmulator(rawdb.NewDatabase(evm.NewKVAdapter(buffered.NewBufferedKVStore(state))), genesisAlloc)
}

func initialize(ctx coretypes.Sandbox) (dict.Dict, error) {
	// this commits the genesis block
	emu := emulator(ctx.State())
	defer emu.Close()
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

	emu := emulatorR(ctx.State())
	defer emu.Close()

	state, err := emu.Blockchain().State()
	a.RequireNoError(err)
	bal := state.GetBalance(addr)

	ret := dict.New()
	ret.Set(FieldBalance, bal.Bytes())
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
