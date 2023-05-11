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
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type blockContext struct {
	emu      *emulator.EVMEmulator
	txs      []*types.Transaction
	receipts []*types.Receipt
}

// openBlockContext creates a new emulator instance before processing any
// requests in the ISC block. The purpose is to create a single Ethereum block
// for each ISC block.
func openBlockContext(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCaller(&isc.NilAgentID{}) // called from ISC VM
	bctx := &blockContext{
		emu: createEmulator(ctx, newL2Balance(ctx)),
	}
	ctx.Privileged().SetBlockContext(bctx)
	return nil
}

// closeBlockContext "mints" the Ethereum block after all requests in the ISC
// block have been processed.
func closeBlockContext(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCaller(&isc.NilAgentID{}) // called from ISC VM
	getBlockContext(ctx).mintBlock()
	return nil
}

func getBlockContext(ctx isc.Sandbox) *blockContext {
	return ctx.Privileged().BlockContext().(*blockContext)
}

func (bctx *blockContext) mintBlock() {
	// count txs where status = success (which are already stored in the pending block)
	txCount := uint(0)
	for i := range bctx.txs {
		if bctx.receipts[i].Status == types.ReceiptStatusSuccessful {
			txCount++
		}
	}

	// failed txs were not stored in the pending block -- store them now
	for i := range bctx.txs {
		if bctx.receipts[i].Status == types.ReceiptStatusSuccessful {
			continue
		}
		bctx.receipts[i].TransactionIndex = txCount
		bctx.emu.BlockchainDB().AddTransaction(bctx.txs[i], bctx.receipts[i])

		// we must also increment the nonce manually since the original request was reverted
		sender := evmutil.MustGetSender(bctx.txs[i])
		nonce := bctx.emu.StateDB().GetNonce(sender)
		bctx.emu.StateDB().SetNonce(sender, nonce+1)

		txCount++
	}

	bctx.emu.MintBlock()
}

func getTracer(ctx isc.Sandbox, bctx *blockContext) tracers.Tracer {
	tracer := ctx.EVMTracer()
	if tracer == nil {
		return nil
	}
	if int(tracer.TxIndex) != len(bctx.txs) {
		// some tx in this block is being traced but not the current one
		return nil
	}
	return tracer.Tracer
}

func gasLimits(ctx isc.SandboxBase) emulator.GasLimits {
	chainInfo := ctx.ChainInfo()
	return emulator.GasLimits{
		Block: gas.EVMBlockGasLimit(chainInfo.GasLimits, &chainInfo.GasFeePolicy.EVMGasRatio),
		Call:  gas.EVMCallGasLimit(chainInfo.GasLimits, &chainInfo.GasFeePolicy.EVMGasRatio),
	}
}

func createEmulator(ctx isc.Sandbox, l2Balance *l2Balance) *emulator.EVMEmulator {
	return emulator.NewEVMEmulator(
		evmStateSubrealm(ctx.State()),
		timestamp(ctx.Timestamp()),
		gasLimits(ctx),
		newMagicContract(ctx),
		l2Balance,
	)
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
