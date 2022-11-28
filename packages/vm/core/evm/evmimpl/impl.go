// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/evm/solidity"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

var Processor = evm.Contract.Processor(initialize,
	evm.FuncOpenBlockContext.WithHandler(restricted(openBlockContext)),
	evm.FuncCloseBlockContext.WithHandler(restricted(closeBlockContext)),
	evm.FuncSendTransaction.WithHandler(restricted(applyTransaction)),
	evm.FuncEstimateGas.WithHandler(restricted(estimateGas)),
	evm.FuncRegisterERC20NativeToken.WithHandler(restricted(registerERC20NativeToken)),

	// views
	evm.FuncGetBalance.WithHandler(restrictedView(getBalance)),
	evm.FuncCallContract.WithHandler(restrictedView(callContract)),
	evm.FuncGetNonce.WithHandler(restrictedView(getNonce)),
	evm.FuncGetReceipt.WithHandler(restrictedView(getReceipt)),
	evm.FuncGetCode.WithHandler(restrictedView(getCode)),
	evm.FuncGetBlockNumber.WithHandler(restrictedView(getBlockNumber)),
	evm.FuncGetBlockByNumber.WithHandler(restrictedView(getBlockByNumber)),
	evm.FuncGetBlockByHash.WithHandler(restrictedView(getBlockByHash)),
	evm.FuncGetTransactionByHash.WithHandler(restrictedView(getTransactionByHash)),
	evm.FuncGetTransactionByBlockHashAndIndex.WithHandler(restrictedView(getTransactionByBlockHashAndIndex)),
	evm.FuncGetTransactionByBlockNumberAndIndex.WithHandler(restrictedView(getTransactionByBlockNumberAndIndex)),
	evm.FuncGetTransactionCountByBlockHash.WithHandler(restrictedView(getTransactionCountByBlockHash)),
	evm.FuncGetTransactionCountByBlockNumber.WithHandler(restrictedView(getTransactionCountByBlockNumber)),
	evm.FuncGetStorage.WithHandler(restrictedView(getStorage)),
	evm.FuncGetLogs.WithHandler(restrictedView(getLogs)),
	evm.FuncGetChainID.WithHandler(restrictedView(getChainID)),
)

func initialize(ctx isc.Sandbox) dict.Dict {
	genesisAlloc := core.GenesisAlloc{}
	var err error
	if ctx.Params().MustHas(evm.FieldGenesisAlloc) {
		genesisAlloc, err = evmtypes.DecodeGenesisAlloc(ctx.Params().MustGet(evm.FieldGenesisAlloc))
		ctx.RequireNoError(err)
	}

	gasLimit, err := codec.DecodeUint64(ctx.Params().MustGet(evm.FieldBlockGasLimit), evm.BlockGasLimitDefault)
	ctx.RequireNoError(err)

	blockKeepAmount, err := codec.DecodeInt32(ctx.Params().MustGet(evm.FieldBlockKeepAmount), evm.BlockKeepAmountDefault)
	ctx.RequireNoError(err)

	// add the standard ISC contract at arbitrary address 0x1074...
	deployMagicContractOnGenesis(genesisAlloc)

	// add the standard ERC20 contract
	genesisAlloc[iscmagic.ERC20BaseTokensAddress] = core.GenesisAccount{
		Code:    iscmagic.ERC20BaseTokensRuntimeBytecode,
		Storage: map[common.Hash]common.Hash{},
		Balance: &big.Int{},
	}
	addToPrivileged(ctx, iscmagic.ERC20BaseTokensAddress)

	// add the standard ERC721 contract
	genesisAlloc[iscmagic.ERC721NFTsAddress] = core.GenesisAccount{
		Code:    iscmagic.ERC721NFTsRuntimeBytecode,
		Storage: map[common.Hash]common.Hash{},
		Balance: &big.Int{},
	}
	addToPrivileged(ctx, iscmagic.ERC721NFTsAddress)

	chainID := evmtypes.MustDecodeChainID(ctx.Params().MustGet(evm.FieldChainID), evm.DefaultChainID)
	emulator.Init(
		evmStateSubrealm(ctx.State()),
		chainID,
		blockKeepAmount,
		gasLimit,
		timestamp(ctx),
		genesisAlloc,
		getBalanceFunc(ctx),
		getSubBalanceFunc(ctx),
		getAddBalanceFunc(ctx),
	)

	// storing hname as a terminal value of the contract's state nil key.
	// This way we will be able to retrieve commitment to the contract's state
	ctx.State().Set("", ctx.Contract().Bytes())

	ctx.Privileged().SubscribeBlockContext(evm.FuncOpenBlockContext.Hname(), evm.FuncCloseBlockContext.Hname())

	return nil
}

func applyTransaction(ctx isc.Sandbox) dict.Dict {
	// we only want to charge gas for the actual execution of the ethereum tx
	ctx.Privileged().GasBurnEnable(false)
	defer ctx.Privileged().GasBurnEnable(true)

	tx, err := evmtypes.DecodeTransaction(ctx.Params().MustGet(evm.FieldTransaction))
	ctx.RequireNoError(err)

	ctx.RequireCaller(isc.NewEthereumAddressAgentID(evmutil.MustGetSender(tx)))

	// next block will be minted when the ISC block is closed
	bctx := getBlockContext(ctx)

	ctx.Requiref(tx.ChainId().Uint64() == uint64(bctx.emu.BlockchainDB().GetChainID()), "chainId mismatch")

	// Send the tx to the emulator.
	// ISC gas burn will be enabled right before executing the tx, and disabled right after,
	// so that ISC magic calls are charged gas.
	receipt, result, err := bctx.emu.SendTransaction(tx, ctx.Privileged().GasBurnEnable)

	// burn EVM gas as ISC gas
	var gasErr error
	if result != nil {
		// convert burnt EVM gas to ISC gas
		gasRatio := getGasRatio(ctx)
		ctx.Privileged().GasBurnEnable(true)
		gasErr = panicutil.CatchPanic(
			func() {
				ctx.Gas().Burn(gas.BurnCodeEVM1P, evmtypes.EVMGasToISC(result.UsedGas, &gasRatio))
			},
		)
		ctx.Privileged().GasBurnEnable(false)
		if gasErr != nil {
			// out of gas when burning ISC gas, edit the EVM receipt so that it fails
			receipt.Status = types.ReceiptStatusFailed
		}
	}

	if receipt != nil { // receipt can be nil when "intrinsic gas too low" or not enough funds
		// If EVM execution was reverted we must revert the ISC request as well.
		// Failed txs will be stored when closing the block context.
		bctx.txs = append(bctx.txs, tx)
		bctx.receipts = append(bctx.receipts, receipt)
	}
	ctx.RequireNoError(err)
	ctx.RequireNoError(gasErr)
	ctx.RequireNoError(tryGetRevertError(result))

	return nil
}

func registerERC20NativeToken(ctx isc.Sandbox) dict.Dict {
	foundrySN := codec.MustDecodeUint32(ctx.Params().MustGet(evm.FieldFoundrySN))
	name := codec.MustDecodeString(ctx.Params().MustGet(evm.FieldTokenName))
	tickerSymbol := codec.MustDecodeString(ctx.Params().MustGet(evm.FieldTokenTickerSymbol))
	decimals := codec.MustDecodeUint8(ctx.Params().MustGet(evm.FieldTokenDecimals))

	{
		res := ctx.CallView(accounts.Contract.Hname(), accounts.ViewAccountFoundries.Hname(), dict.Dict{
			accounts.ParamAgentID: codec.EncodeAgentID(ctx.Caller()),
		})
		ctx.Requiref(res[kv.Key(codec.EncodeUint32(foundrySN))] != nil, "foundry sn %s not owned by caller", foundrySN)
	}

	// deploy the contract to the EVM state
	addr := iscmagic.ERC20NativeTokensAddress(foundrySN)
	emu := createEmulator(ctx)
	evmState := emu.StateDB()
	evmState.CreateAccount(addr)
	evmState.SetCode(addr, iscmagic.ERC20NativeTokensRuntimeBytecode)
	// see ERC20NativeTokens_storage.json
	evmState.SetState(addr, solidity.StorageSlot(0), solidity.StorageEncodeShortString(name))
	evmState.SetState(addr, solidity.StorageSlot(1), solidity.StorageEncodeShortString(tickerSymbol))
	evmState.SetState(addr, solidity.StorageSlot(2), solidity.StorageEncodeUint8(decimals))

	addToPrivileged(ctx, addr)

	return nil
}

func getBalance(ctx isc.SandboxView) dict.Dict {
	addr := common.BytesToAddress(ctx.Params().MustGet(evm.FieldAddress))
	emu := createEmulatorR(ctx)
	return result(emu.StateDB().GetBalance(addr).Bytes())
}

func getBlockNumber(ctx isc.SandboxView) dict.Dict {
	emu := createEmulatorR(ctx)
	return result(new(big.Int).SetUint64(emu.BlockchainDB().GetNumber()).Bytes())
}

func getBlockByNumber(ctx isc.SandboxView) dict.Dict {
	return blockResult(blockByNumber(ctx))
}

func getBlockByHash(ctx isc.SandboxView) dict.Dict {
	return blockResult(blockByHash(ctx))
}

func getTransactionByHash(ctx isc.SandboxView) dict.Dict {
	return txResult(transactionByHash(ctx))
}

func getTransactionByBlockHashAndIndex(ctx isc.SandboxView) dict.Dict {
	return txResult(transactionByBlockHashAndIndex(ctx))
}

func getTransactionByBlockNumberAndIndex(ctx isc.SandboxView) dict.Dict {
	return txResult(transactionByBlockNumberAndIndex(ctx))
}

func getTransactionCountByBlockHash(ctx isc.SandboxView) dict.Dict {
	return txCountResult(blockByHash(ctx))
}

func getTransactionCountByBlockNumber(ctx isc.SandboxView) dict.Dict {
	return txCountResult(blockByNumber(ctx))
}

func getReceipt(ctx isc.SandboxView) dict.Dict {
	txHash := common.BytesToHash(ctx.Params().MustGet(evm.FieldTransactionHash))
	emu := createEmulatorR(ctx)
	r := emu.BlockchainDB().GetReceiptByTxHash(txHash)
	if r == nil {
		return nil
	}
	return result(evmtypes.EncodeReceiptFull(r))
}

func getNonce(ctx isc.SandboxView) dict.Dict {
	emu := createEmulatorR(ctx)
	addr := common.BytesToAddress(ctx.Params().MustGet(evm.FieldAddress))
	return result(codec.EncodeUint64(emu.StateDB().GetNonce(addr)))
}

func getCode(ctx isc.SandboxView) dict.Dict {
	emu := createEmulatorR(ctx)
	addr := common.BytesToAddress(ctx.Params().MustGet(evm.FieldAddress))
	return result(emu.StateDB().GetCode(addr))
}

func getStorage(ctx isc.SandboxView) dict.Dict {
	emu := createEmulatorR(ctx)
	addr := common.BytesToAddress(ctx.Params().MustGet(evm.FieldAddress))
	key := common.BytesToHash(ctx.Params().MustGet(evm.FieldKey))
	data := emu.StateDB().GetState(addr, key)
	return result(data[:])
}

func getLogs(ctx isc.SandboxView) dict.Dict {
	q, err := evmtypes.DecodeFilterQuery(ctx.Params().MustGet(evm.FieldFilterQuery))
	ctx.RequireNoError(err)
	emu := createEmulatorR(ctx)
	logs := emu.FilterLogs(q)
	return result(evmtypes.EncodeLogs(logs))
}

func getChainID(ctx isc.SandboxView) dict.Dict {
	emu := createEmulatorR(ctx)
	return result(evmtypes.EncodeChainID(emu.BlockchainDB().GetChainID()))
}

func callContract(ctx isc.SandboxView) dict.Dict {
	callMsg, err := evmtypes.DecodeCallMsg(ctx.Params().MustGet(evm.FieldCallMsg))
	ctx.RequireNoError(err)
	emu := createEmulatorR(ctx)
	res, err := emu.CallContract(callMsg, nil)
	ctx.RequireNoError(err)
	ctx.RequireNoError(tryGetRevertError(res))
	return result(res.Return())
}

func tryGetRevertError(res *core.ExecutionResult) error {
	// try to include the revert reason in the error
	if res.Err == nil {
		return nil
	}
	if len(res.Revert()) > 0 {
		reason, errUnpack := abi.UnpackRevert(res.Revert())
		if errUnpack == nil {
			return fmt.Errorf("%s: %v", res.Err.Error(), reason)
		}
	}
	return res.Err
}

func estimateGas(ctx isc.Sandbox) dict.Dict {
	// we only want to charge gas for the actual execution of the ethereum tx
	ctx.Privileged().GasBurnEnable(false)
	defer ctx.Privileged().GasBurnEnable(true)

	callMsg, err := evmtypes.DecodeCallMsg(ctx.Params().MustGet(evm.FieldCallMsg))
	ctx.RequireNoError(err)
	ctx.RequireCaller(isc.NewEthereumAddressAgentID(callMsg.From))

	emu := createEmulator(ctx)

	res, err := emu.CallContract(callMsg, ctx.Privileged().GasBurnEnable)
	ctx.RequireNoError(err)
	ctx.RequireNoError(tryGetRevertError(res))

	gasRatio := getGasRatio(ctx)
	{
		// burn the used EVM gas as it would be done for a normal request call
		ctx.Privileged().GasBurnEnable(true)
		gasErr := panicutil.CatchPanic(
			func() {
				ctx.Gas().Burn(gas.BurnCodeEVM1P, evmtypes.EVMGasToISC(res.UsedGas, &gasRatio))
			},
		)
		ctx.Privileged().GasBurnEnable(false)
		ctx.RequireNoError(gasErr)
	}

	finalEvmGasUsed := evmtypes.ISCGasBurnedToEVM(ctx.Gas().Burned(), &gasRatio)

	return result(codec.EncodeUint64(finalEvmGasUsed))
}

func getGasRatio(ctx isc.SandboxBase) util.Ratio32 {
	gasRatioViewRes := ctx.CallView(governance.Contract.Hname(), governance.ViewGetEVMGasRatio.Hname(), nil)
	return codec.MustDecodeRatio32(gasRatioViewRes.MustGet(governance.ParamEVMGasRatio), evmtypes.DefaultGasRatio)
}
