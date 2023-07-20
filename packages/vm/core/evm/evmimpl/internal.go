// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"math/big"
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
	"github.com/iotaledger/wasp/packages/util"
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
		evm.EmulatorStateSubrealm(ctx.State()),
		timestamp(ctx.Timestamp()),
		gasLimits(chainInfo),
		chainInfo.BlockKeepAmount,
		newMagicContract(ctx),
		newL2Balance(ctx),
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
		emulator.NewStateDB(evm.EmulatorStateSubrealm(evmPartition), nil).IncNonce(evmutil.MustGetSender(tx))
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

type l2BalanceR struct {
	ctx isc.SandboxBase
}

func newL2BalanceR(ctx isc.SandboxBase) *l2BalanceR {
	return &l2BalanceR{
		ctx: ctx,
	}
}

type l2Balance struct {
	*l2BalanceR
	ctx isc.Sandbox
}

func newL2Balance(ctx isc.Sandbox) *l2Balance {
	return &l2Balance{
		l2BalanceR: newL2BalanceR(ctx),
		ctx:        ctx,
	}
}

func (b *l2BalanceR) Get(addr common.Address) *big.Int {
	res := b.ctx.CallView(
		accounts.Contract.Hname(),
		accounts.ViewBalanceBaseToken.Hname(),
		dict.Dict{accounts.ParamAgentID: isc.NewEthereumAddressAgentID(addr).Bytes()},
	)
	decimals := parameters.L1().BaseToken.Decimals
	ret := new(big.Int).SetUint64(codec.MustDecodeUint64(res.Get(accounts.ParamBalance), 0))
	return util.CustomTokensDecimalsToEthereumDecimals(ret, decimals)
}

func (b *l2BalanceR) Add(addr common.Address, amount *big.Int) {
	panic("should not be called")
}

func (b *l2BalanceR) Sub(addr common.Address, amount *big.Int) {
	panic("should not be called")
}

func assetsForFeeFromEthereumDecimals(amount *big.Int) *isc.Assets {
	decimals := parameters.L1().BaseToken.Decimals
	amt := util.EthereumDecimalsToCustomTokenDecimals(amount, decimals)
	return isc.NewAssetsBaseTokens(amt.Uint64())
}

func (b *l2Balance) Add(addr common.Address, amount *big.Int) {
	tokens := assetsForFeeFromEthereumDecimals(amount)
	b.ctx.Privileged().CreditToAccount(isc.NewEthereumAddressAgentID(addr), tokens)
}

func (b *l2Balance) Sub(addr common.Address, amount *big.Int) {
	tokens := assetsForFeeFromEthereumDecimals(amount)
	account := isc.NewEthereumAddressAgentID(addr)
	b.ctx.Privileged().DebitFromAccount(account, tokens)
}
