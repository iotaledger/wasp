// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers"

	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// MintBlock "mints" the Ethereum block after all requests in the ISC
// block have been processed.
// IMPORTANT: Must only be called from the ISC VM
func MintBlock(evmPartition kv.KVStore, chainInfo *isc.ChainInfo, blockTimestamp time.Time) {
	createBlockchainDB(evmPartition, chainInfo).MintBlock(timestamp(blockTimestamp))
}

func getTracer(ctx isc.Sandbox) tracers.Tracer {
	tracer := ctx.EVMTracer()
	if tracer == nil {
		return nil
	}
	return tracer.Tracer
}

func createEmulator(ctx isc.Sandbox) *emulator.EVMEmulator {
	return emulator.NewEVMEmulator(newEmulatorContext(ctx))
}

func createBlockchainDB(evmPartition kv.KVStore, chainInfo *isc.ChainInfo) *emulator.BlockchainDB {
	return emulator.NewBlockchainDB(evm.EmulatorStateSubrealm(evmPartition), gasLimits(chainInfo).Block, chainInfo.BlockKeepAmount)
}

func saveExecutedTx(
	evmPartition kv.KVStore,
	chainInfo *isc.ChainInfo,
	tx *types.Transaction,
	receipt *types.Receipt,
) {
	createBlockchainDB(evmPartition, chainInfo).AddTransaction(tx, receipt)
	// make sure the nonce is incremented if the state was rolled back by the VM
	if receipt.Status != types.ReceiptStatusSuccessful {
		emulator.IncNonce(emulator.StateDBSubrealm(evm.EmulatorStateSubrealm(evmPartition)), evmutil.MustGetSender(tx))
	}
}

func gasLimits(chainInfo *isc.ChainInfo) emulator.GasLimits {
	return emulator.GasLimits{
		Block: gas.EVMBlockGasLimit(chainInfo.GasLimits, &chainInfo.GasFeePolicy.EVMGasRatio),
		Call:  gas.EVMCallGasLimit(chainInfo.GasLimits, &chainInfo.GasFeePolicy.EVMGasRatio),
	}
}

// timestamp returns the current timestamp in seconds since epoch
func timestamp(t time.Time) uint64 {
	return uint64(t.Unix())
}

func result(value []byte) dict.Dict {
	if value == nil {
		return nil
	}
	return dict.Dict{evm.FieldResult: value}
}

type emulatorContext struct {
	sandbox isc.Sandbox
}

var _ emulator.Context = &emulatorContext{}

func newEmulatorContext(sandbox isc.Sandbox) *emulatorContext {
	return &emulatorContext{
		sandbox: sandbox,
	}
}

func (ctx *emulatorContext) BlockKeepAmount() int32 {
	ret := int32(0)
	// do not charge gas for this, internal checks of the emulator require this function to run before executing the request
	ctx.WithoutGasBurn(func() {
		ret = ctx.sandbox.ChainInfo().BlockKeepAmount
	})
	return ret
}

func (ctx *emulatorContext) GasLimits() emulator.GasLimits {
	var ret emulator.GasLimits
	// do not charge gas for this, internal checks of the emulator require this function to run before executing the request
	ctx.WithoutGasBurn(func() {
		ret = gasLimits(ctx.sandbox.ChainInfo())
	})
	return ret
}

func (ctx *emulatorContext) MagicContracts() map[common.Address]vm.ISCMagicContract {
	return newMagicContract(ctx.sandbox)
}

func (ctx *emulatorContext) State() kv.KVStore {
	return evm.EmulatorStateSubrealm(ctx.sandbox.State())
}

func (ctx *emulatorContext) Timestamp() uint64 {
	return timestamp(ctx.sandbox.Timestamp())
}

func (*emulatorContext) BaseTokensDecimals() uint32 {
	return parameters.L1().BaseToken.Decimals
}

func (ctx *emulatorContext) GetBaseTokensBalance(addr common.Address) uint64 {
	ret := uint64(0)
	// do not charge gas for this, internal checks of the emulator require this function to run before executing the request
	ctx.WithoutGasBurn(func() {
		res := ctx.sandbox.CallView(
			accounts.Contract.Hname(),
			accounts.ViewBalanceBaseToken.Hname(),
			dict.Dict{accounts.ParamAgentID: isc.NewEthereumAddressAgentID(ctx.sandbox.ChainID(), addr).Bytes()},
		)
		ret = codec.MustDecodeUint64(res.Get(accounts.ParamBalance), 0)
	})
	return ret
}

func (ctx *emulatorContext) AddBaseTokensBalance(addr common.Address, amount uint64) {
	ctx.sandbox.Privileged().CreditToAccount(
		isc.NewEthereumAddressAgentID(ctx.sandbox.ChainID(), addr),
		isc.NewAssetsBaseTokens(amount),
	)
}

func (ctx *emulatorContext) SubBaseTokensBalance(addr common.Address, amount uint64) {
	ctx.sandbox.Privileged().DebitFromAccount(
		isc.NewEthereumAddressAgentID(ctx.sandbox.ChainID(), addr),
		isc.NewAssetsBaseTokens(amount),
	)
}

func (ctx *emulatorContext) TakeSnapshot() int {
	return ctx.sandbox.TakeStateSnapshot()
}

func (ctx *emulatorContext) RevertToSnapshot(i int) {
	ctx.sandbox.RevertToStateSnapshot(i)
}

func (ctx *emulatorContext) WithoutGasBurn(f func()) {
	prev := ctx.sandbox.Privileged().GasBurnEnabled()
	ctx.sandbox.Privileged().GasBurnEnable(false)
	f()
	ctx.sandbox.Privileged().GasBurnEnable(prev)
}
