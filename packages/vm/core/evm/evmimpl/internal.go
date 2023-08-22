// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
	chainInfo := ctx.ChainInfo()
	return emulator.NewEVMEmulator(
		newL2StateForEmulator(ctx),
		timestamp(ctx.Timestamp()),
		gasLimits(chainInfo),
		chainInfo.BlockKeepAmount,
		newMagicContract(ctx),
	)
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

type l2StateForEmulator struct {
	kv.KVStore
	ctx isc.Sandbox
}

var _ emulator.L2State = &l2StateForEmulator{}

func newL2StateForEmulator(ctx isc.Sandbox) *l2StateForEmulator {
	return &l2StateForEmulator{
		KVStore: evm.EmulatorStateSubrealm(ctx.State()),
		ctx:     ctx,
	}
}

func (*l2StateForEmulator) Decimals() uint32 {
	return parameters.L1().BaseToken.Decimals
}

func (l2 *l2StateForEmulator) GetBalance(addr common.Address) uint64 {
	res := l2.ctx.CallView(
		accounts.Contract.Hname(),
		accounts.ViewBalanceBaseToken.Hname(),
		dict.Dict{accounts.ParamAgentID: isc.NewEthereumAddressAgentID(addr).Bytes()},
	)
	return codec.MustDecodeUint64(res.Get(accounts.ParamBalance), 0)
}

func (l2 *l2StateForEmulator) AddBalance(addr common.Address, amount uint64) {
	l2.ctx.Privileged().CreditToAccount(isc.NewEthereumAddressAgentID(addr), isc.NewAssetsBaseTokens(amount))
}

func (l2 *l2StateForEmulator) SubBalance(addr common.Address, amount uint64) {
	l2.ctx.Privileged().DebitFromAccount(isc.NewEthereumAddressAgentID(addr), isc.NewAssetsBaseTokens(amount))
}

func (l2 *l2StateForEmulator) TakeSnapshot() int {
	return l2.ctx.TakeStateSnapshot()
}

func (l2 *l2StateForEmulator) RevertToSnapshot(i int) {
	l2.ctx.RevertToStateSnapshot(i)
}
