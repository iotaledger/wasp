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

	iotago "github.com/iotaledger/iota.go/v3"
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

var Processor = evm.Contract.Processor(nil,
	evm.FuncOpenBlockContext.WithHandler(restricted(openBlockContext)),
	evm.FuncCloseBlockContext.WithHandler(restricted(closeBlockContext)),
	evm.FuncSendTransaction.WithHandler(restricted(applyTransaction)),
	evm.FuncCallContract.WithHandler(restricted(callContract)),
	evm.FuncRegisterERC20NativeToken.WithHandler(restricted(registerERC20NativeToken)),
	evm.FuncRegisterERC20NativeTokenOnRemoteChain.WithHandler(restricted(registerERC20NativeTokenOnRemoteChain)),
	evm.FuncRegisterERC20ExternalNativeToken.WithHandler(registerERC20ExternalNativeToken),
	evm.FuncRegisterERC721NFTCollection.WithHandler(restricted(registerERC721NFTCollection)),

	// views
	evm.FuncGetBalance.WithHandler(restrictedView(getBalance)),
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
	evm.FuncGetERC20ExternalNativeTokenAddress.WithHandler(restrictedView(viewERC20ExternalNativeTokenAddress)),
)

func SetInitialState(state kv.KVStore, evmChainID uint16, blockKeepAmount int32) {
	// add the standard ISC contract at arbitrary address 0x1074...
	genesisAlloc := core.GenesisAlloc{}
	deployMagicContractOnGenesis(genesisAlloc)

	// add the standard ERC20 contract
	genesisAlloc[iscmagic.ERC20BaseTokensAddress] = core.GenesisAccount{
		Code:    iscmagic.ERC20BaseTokensRuntimeBytecode,
		Storage: map[common.Hash]common.Hash{},
		Balance: nil,
	}
	addToPrivileged(state, iscmagic.ERC20BaseTokensAddress)

	// add the standard ERC721 contract
	genesisAlloc[iscmagic.ERC721NFTsAddress] = core.GenesisAccount{
		Code:    iscmagic.ERC721NFTsRuntimeBytecode,
		Storage: map[common.Hash]common.Hash{},
		Balance: nil,
	}
	addToPrivileged(state, iscmagic.ERC721NFTsAddress)

	// chain always starts with default gas fee & limits configuration
	gasLimits := gas.LimitsDefault
	gasRatio := gas.DefaultFeePolicy().EVMGasRatio
	emulator.Init(
		evmStateSubrealm(state),
		evmChainID,
		blockKeepAmount,
		emulator.GasLimits{
			Block: gas.EVMBlockGasLimit(gasLimits, &gasRatio),
			Call:  gas.EVMCallGasLimit(gasLimits, &gasRatio),
		},
		0,
		genesisAlloc,
	)

	// subscription to block context is now done in `vmcontext/bootstrapstate.go`
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
		chainInfo := ctx.ChainInfo()
		ctx.Privileged().GasBurnEnable(true)
		gasErr = panicutil.CatchPanic(
			func() {
				ctx.Gas().Burn(
					gas.BurnCodeEVM1P,
					gas.EVMGasToISC(result.UsedGas, &chainInfo.GasFeePolicy.EVMGasRatio),
				)
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
	emu := getBlockContext(ctx).emu
	evmState := emu.StateDB()
	ctx.Requiref(!evmState.Exist(addr), "cannot register ERC20NativeTokens contract: EVM account already exists")
	evmState.CreateAccount(addr)
	evmState.SetCode(addr, iscmagic.ERC20NativeTokensRuntimeBytecode)
	// see ERC20NativeTokens_storage.json
	evmState.SetState(addr, solidity.StorageSlot(0), solidity.StorageEncodeShortString(name))
	evmState.SetState(addr, solidity.StorageSlot(1), solidity.StorageEncodeShortString(tickerSymbol))
	evmState.SetState(addr, solidity.StorageSlot(2), solidity.StorageEncodeUint8(decimals))

	addToPrivileged(ctx.State(), addr)

	return nil
}

func registerERC20NativeTokenOnRemoteChain(ctx isc.Sandbox) dict.Dict {
	foundrySN := codec.MustDecodeUint32(ctx.Params().MustGet(evm.FieldFoundrySN))
	name := codec.MustDecodeString(ctx.Params().MustGet(evm.FieldTokenName))
	tickerSymbol := codec.MustDecodeString(ctx.Params().MustGet(evm.FieldTokenTickerSymbol))
	decimals := codec.MustDecodeUint8(ctx.Params().MustGet(evm.FieldTokenDecimals))
	target := codec.MustDecodeAddress(ctx.Params().MustGet(evm.FieldTargetAddress))
	ctx.Requiref(target.Type() == iotago.AddressAlias, "target must be alias address")

	{
		res := ctx.CallView(accounts.Contract.Hname(), accounts.ViewAccountFoundries.Hname(), dict.Dict{
			accounts.ParamAgentID: codec.EncodeAgentID(ctx.Caller()),
		})
		ctx.Requiref(res[kv.Key(codec.EncodeUint32(foundrySN))] != nil, "foundry sn %s not owned by caller", foundrySN)
	}

	tokenScheme := func() iotago.TokenScheme {
		res := ctx.CallView(accounts.Contract.Hname(), accounts.ViewFoundryOutput.Hname(), dict.Dict{
			accounts.ParamFoundrySN: codec.EncodeUint32(foundrySN),
		})
		o := codec.MustDecodeOutput(res[accounts.ParamFoundryOutputBin])
		foundryOutput, ok := o.(*iotago.FoundryOutput)
		ctx.Requiref(ok, "expected foundry output")
		return foundryOutput.TokenScheme
	}()

	req := isc.RequestParameters{
		TargetAddress: target,
		Assets:        isc.NewEmptyAssets(),
		Metadata: &isc.SendMetadata{
			TargetContract: evm.Contract.Hname(),
			EntryPoint:     evm.FuncRegisterERC20ExternalNativeToken.Hname(),
			Params: dict.Dict{
				evm.FieldFoundrySN:          codec.EncodeUint32(foundrySN),
				evm.FieldTokenName:          codec.EncodeString(name),
				evm.FieldTokenTickerSymbol:  codec.EncodeString(tickerSymbol),
				evm.FieldTokenDecimals:      codec.EncodeUint8(decimals),
				evm.FieldFoundryTokenScheme: codec.EncodeTokenScheme(tokenScheme),
			},
		},
	}
	sd := ctx.EstimateRequiredStorageDeposit(req)
	ctx.TransferAllowedFunds(ctx.AccountID(), isc.NewAssetsBaseTokens(sd))
	req.Assets.AddBaseTokens(sd)
	ctx.Send(req)

	return nil
}

func registerERC20ExternalNativeToken(ctx isc.Sandbox) dict.Dict {
	caller, ok := ctx.Caller().(*isc.ContractAgentID)
	ctx.Requiref(ok, "sender must be an alias address")
	ctx.Requiref(!ctx.ChainID().Equals(caller.ChainID()), "foundry must be off-chain")
	alias := caller.ChainID().AsAliasAddress()

	name := codec.MustDecodeString(ctx.Params().MustGet(evm.FieldTokenName))
	tickerSymbol := codec.MustDecodeString(ctx.Params().MustGet(evm.FieldTokenTickerSymbol))
	decimals := codec.MustDecodeUint8(ctx.Params().MustGet(evm.FieldTokenDecimals))

	// TODO: We should somehow inspect the real FoundryOutput, but it is on L1.
	// Here we reproduce it from the given params (which we assume to be correct)
	// in order to derive the FoundryID
	foundrySN := codec.MustDecodeUint32(ctx.Params().MustGet(evm.FieldFoundrySN))
	tokenScheme := codec.MustDecodeTokenScheme(ctx.Params().MustGet(evm.FieldFoundryTokenScheme))
	simpleTS, ok := tokenScheme.(*iotago.SimpleTokenScheme)
	ctx.Requiref(ok, "only simple token scheme is supported")
	f := &iotago.FoundryOutput{
		SerialNumber: foundrySN,
		TokenScheme:  tokenScheme,
		Conditions: []iotago.UnlockCondition{&iotago.ImmutableAliasUnlockCondition{
			Address: &alias,
		}},
	}
	nativeTokenID, err := f.ID()
	ctx.RequireNoError(err)

	_, ok = getERC20ExternalNativeTokensAddress(ctx, nativeTokenID)
	ctx.Requiref(!ok, "native token already registered")

	emu := getBlockContext(ctx).emu
	evmState := emu.StateDB()

	addr, err := iscmagic.ERC20ExternalNativeTokensAddress(nativeTokenID, evmState.Exist)
	ctx.RequireNoError(err)

	addERC20ExternalNativeTokensAddress(ctx, nativeTokenID, addr)

	// deploy the contract to the EVM state
	evmState.CreateAccount(addr)
	evmState.SetCode(addr, iscmagic.ERC20ExternalNativeTokensRuntimeBytecode)
	// see ERC20ExternalNativeTokens_storage.json
	evmState.SetState(addr, solidity.StorageSlot(0), solidity.StorageEncodeShortString(name))
	evmState.SetState(addr, solidity.StorageSlot(1), solidity.StorageEncodeShortString(tickerSymbol))
	evmState.SetState(addr, solidity.StorageSlot(2), solidity.StorageEncodeUint8(decimals))
	for k, v := range solidity.StorageEncodeBytes(3, nativeTokenID[:]) {
		evmState.SetState(addr, k, v)
	}
	evmState.SetState(addr, solidity.StorageSlot(4), solidity.StorageEncodeUint256(simpleTS.MaximumSupply))

	addToPrivileged(ctx.State(), addr)

	return result(addr[:])
}

func viewERC20ExternalNativeTokenAddress(ctx isc.SandboxView) dict.Dict {
	nativeTokenID := codec.MustDecodeNativeTokenID(ctx.Params().MustGet(evm.FieldNativeTokenID))
	addr, ok := getERC20ExternalNativeTokensAddress(ctx, nativeTokenID)
	if !ok {
		return nil
	}
	return result(addr[:])
}

func registerERC721NFTCollection(ctx isc.Sandbox) dict.Dict {
	collectionID := codec.MustDecodeNFTID(ctx.Params().MustGet(evm.FieldNFTCollectionID))

	// The collection NFT must be deposited into the chain before registering. Afterwards it may be
	// withdrawn to L1.
	collection := func() *isc.NFT {
		res := ctx.CallView(accounts.Contract.Hname(), accounts.ViewNFTData.Hname(), dict.Dict{
			accounts.ParamNFTID: codec.EncodeNFTID(collectionID),
		})
		collection, err := isc.NFTFromBytes(res[accounts.ParamNFTData])
		ctx.RequireNoError(err)
		return collection
	}()

	metadata, err := isc.IRC27NFTMetadataFromBytes(collection.Metadata)
	ctx.RequireNoError(err, "cannot decode IRC27 collection NFT metadata")

	// deploy the contract to the EVM state
	addr := iscmagic.ERC721NFTCollectionAddress(collectionID)
	emu := getBlockContext(ctx).emu
	evmState := emu.StateDB()
	ctx.Requiref(!evmState.Exist(addr), "cannot register ERC721NFTCollection contract: EVM account already exists")
	evmState.CreateAccount(addr)
	evmState.SetCode(addr, iscmagic.ERC721NFTCollectionRuntimeBytecode)
	// see ERC721NFTCollection_storage.json
	evmState.SetState(addr, solidity.StorageSlot(2), solidity.StorageEncodeBytes32(collectionID[:]))
	for k, v := range solidity.StorageEncodeString(3, metadata.Name) {
		evmState.SetState(addr, k, v)
	}

	addToPrivileged(ctx.State(), addr)

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

// callContract is called from the jsonrpc eth_estimateGas and eth_call endpoints.
// The VM is in estimate gas mode, and any state mutations are discarded.
func callContract(ctx isc.Sandbox) dict.Dict {
	// we only want to charge gas for the actual execution of the ethereum tx
	ctx.Privileged().GasBurnEnable(false)
	defer ctx.Privileged().GasBurnEnable(true)

	callMsg, err := evmtypes.DecodeCallMsg(ctx.Params().MustGet(evm.FieldCallMsg))
	ctx.RequireNoError(err)
	ctx.RequireCaller(isc.NewEthereumAddressAgentID(callMsg.From))

	emu := getBlockContext(ctx).emu

	res, err := emu.CallContract(callMsg, ctx.Privileged().GasBurnEnable)
	ctx.RequireNoError(err)
	ctx.RequireNoError(tryGetRevertError(res))

	gasRatio := getGasRatio(ctx)
	{
		// burn the used EVM gas as it would be done for a normal request call
		ctx.Privileged().GasBurnEnable(true)
		gasErr := panicutil.CatchPanic(
			func() {
				ctx.Gas().Burn(gas.BurnCodeEVM1P, gas.EVMGasToISC(res.UsedGas, &gasRatio))
			},
		)
		ctx.Privileged().GasBurnEnable(false)
		ctx.RequireNoError(gasErr)
	}

	return result(res.ReturnData)
}

func getGasRatio(ctx isc.SandboxBase) util.Ratio32 {
	gasRatioViewRes := ctx.CallView(governance.Contract.Hname(), governance.ViewGetEVMGasRatio.Hname(), nil)
	return codec.MustDecodeRatio32(gasRatioViewRes.MustGet(governance.ParamEVMGasRatio), gas.DefaultEVMGasRatio)
}
