// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/assert"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type blockContext struct {
	emu       *emulator.EVMEmulator
	feePolicy *gas.FeePolicy
	txs       []*types.Transaction
	receipts  []*types.Receipt
}

// openBlockContext creates a new emulator instance before processing any
// requests in the ISC block. The purpose is to create a single Ethereum block
// for each ISC block.
func openBlockContext(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCaller(&isc.NilAgentID{}) // called from ISC VM
	feePolicy := getFeePolicy(ctx)
	bctx := &blockContext{feePolicy: feePolicy}
	bctx.emu = createEmulator(
		ctx,
		feePolicy,
		newL2Balance(ctx, func() *gas.FeePolicy { return bctx.feePolicy }),
	)
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
	bctx := ctx.Privileged().BlockContext().(*blockContext)
	bctx.feePolicy = getFeePolicy(ctx) // refresh in case it changed since the last tx
	return bctx
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

func createEmulator(ctx isc.Sandbox, feePolicy *gas.FeePolicy, l2Balance *l2Balance) *emulator.EVMEmulator {
	return emulator.NewEVMEmulator(
		evmStateSubrealm(ctx.State()),
		timestamp(ctx),
		emulator.GasLimits{
			Block: gas.EVMBlockGasLimit(&feePolicy.EVMGasRatio),
			Call:  gas.EVMCallGasLimit(&feePolicy.EVMGasRatio),
		},
		newMagicContract(ctx),
		l2Balance,
	)
}

func createEmulatorR(ctx isc.SandboxView) *emulator.EVMEmulator {
	feePolicy := getFeePolicy(ctx)
	return emulator.NewEVMEmulator(
		evmStateSubrealm(buffered.NewBufferedKVStore(ctx.StateR())),
		timestamp(ctx),
		emulator.GasLimits{
			Block: gas.EVMBlockGasLimit(&feePolicy.EVMGasRatio),
			Call:  gas.EVMCallGasLimit(&feePolicy.EVMGasRatio),
		},
		newMagicContractView(ctx),
		newL2BalanceR(ctx, func() *gas.FeePolicy { return feePolicy }),
	)
}

// timestamp returns the current timestamp in seconds since epoch
func timestamp(ctx isc.SandboxBase) uint64 {
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

func blockByNumber(ctx isc.SandboxView) (*emulator.EVMEmulator, *types.Block) {
	emu := createEmulatorR(ctx)
	blockNumber := paramBlockNumber(ctx, emu, true)
	return emu, emu.BlockchainDB().GetBlockByNumber(blockNumber)
}

func blockByHash(ctx isc.SandboxView) (*emulator.EVMEmulator, *types.Block) {
	emu := createEmulatorR(ctx)
	hash := common.BytesToHash(ctx.Params().MustGet(evm.FieldBlockHash))
	return emu, emu.BlockchainDB().GetBlockByHash(hash)
}

func transactionByHash(ctx isc.SandboxView) (*emulator.EVMEmulator, *types.Transaction) {
	emu := createEmulatorR(ctx)
	txHash := common.BytesToHash(ctx.Params().MustGet(evm.FieldTransactionHash))
	return emu, emu.BlockchainDB().GetTransactionByHash(txHash)
}

func transactionByBlockHashAndIndex(ctx isc.SandboxView) (*emulator.EVMEmulator, *types.Transaction) {
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

func transactionByBlockNumberAndIndex(ctx isc.SandboxView) (*emulator.EVMEmulator, *types.Transaction) {
	emu := createEmulatorR(ctx)
	blockNumber := paramBlockNumber(ctx, emu, true)

	a := assert.NewAssert(ctx.Log())
	index, err := codec.DecodeUint64(ctx.Params().MustGet(evm.FieldTransactionIndex), 0)
	a.RequireNoError(err)

	return emu, emu.BlockchainDB().GetTransactionByBlockNumberAndIndex(blockNumber, uint32(index))
}

func requireLatestBlock(ctx isc.SandboxView, emu *emulator.EVMEmulator, allowPrevious bool, blockNumber uint64) uint64 {
	current := emu.BlockchainDB().GetNumber()
	if blockNumber != current {
		assert.NewAssert(ctx.Log()).Requiref(allowPrevious, "unsupported operation, cannot query previous blocks")
	}
	return blockNumber
}

func paramBlockNumber(ctx isc.SandboxView, emu *emulator.EVMEmulator, allowPrevious bool) uint64 {
	current := emu.BlockchainDB().GetNumber()
	if ctx.Params().MustHas(evm.FieldBlockNumber) {
		blockNumber := new(big.Int).SetBytes(ctx.Params().MustGet(evm.FieldBlockNumber))
		return requireLatestBlock(ctx, emu, allowPrevious, blockNumber.Uint64())
	}
	return current
}

// TODO dropping "customtokens gas fee" might be the way to go
func getFeePolicy(ctx isc.SandboxBase) *gas.FeePolicy {
	res := ctx.CallView(
		governance.Contract.Hname(),
		governance.ViewGetFeePolicy.Hname(),
		nil,
	)
	var err error
	feePolicy, err := gas.FeePolicyFromBytes(res.MustGet(governance.ParamFeePolicyBytes))
	ctx.RequireNoError(err)
	return feePolicy
}

type l2BalanceR struct {
	ctx          isc.SandboxBase
	getFeePolicy func() *gas.FeePolicy
}

func newL2BalanceR(ctx isc.SandboxBase, getFeePolicy func() *gas.FeePolicy) *l2BalanceR {
	return &l2BalanceR{
		ctx:          ctx,
		getFeePolicy: getFeePolicy,
	}
}

type l2Balance struct {
	*l2BalanceR
	ctx isc.Sandbox
}

func newL2Balance(ctx isc.Sandbox, getFeePolicy func() *gas.FeePolicy) *l2Balance {
	return &l2Balance{
		l2BalanceR: newL2BalanceR(ctx, getFeePolicy),
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
	ret := new(big.Int).SetUint64(codec.MustDecodeUint64(res.MustGet(accounts.ParamBalance), 0))
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
