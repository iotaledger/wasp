// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/samber/lo"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/evm/solidity"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

var Processor = evm.Contract.Processor(nil,
	evm.FuncSendTransaction.WithHandler(restricted(applyTransaction)),
	evm.FuncCallContract.WithHandler(restricted(callContract)),

	evm.FuncRegisterERC20NativeToken.WithHandler(registerERC20NativeToken),
	evm.FuncRegisterERC20NativeTokenOnRemoteChain.WithHandler(restricted(registerERC20NativeTokenOnRemoteChain)),
	evm.FuncRegisterERC20ExternalNativeToken.WithHandler(registerERC20ExternalNativeToken),
	evm.FuncRegisterERC721NFTCollection.WithHandler(restricted(registerERC721NFTCollection)),

	evm.FuncNewL1Deposit.WithHandler(newL1Deposit),
	evm.FuncNewL1Withdrawal.WithHandler(newL1Withdrawal),
	evm.FuncNewTransferBetweenL2Accounts.WithHandler(newTransferBetweenL2Accounts),

	// views
	evm.FuncGetERC20ExternalNativeTokenAddress.WithHandler(viewERC20ExternalNativeTokenAddress),
	evm.FuncGetChainID.WithHandler(getChainID),
)

// SetInitialState initializes the evm core contract and the Ethereum genesis
// block on a newly created ISC chain.
func SetInitialState(evmPartition kv.KVStore, evmChainID uint16, createBaseTokenMagicWrap bool) {
	// Ethereum genesis block configuration
	genesisAlloc := core.GenesisAlloc{}

	// add the ISC magic contract at address 0x10740000...00
	genesisAlloc[iscmagic.Address] = core.GenesisAccount{
		// Dummy code, because some contracts check the code size before calling
		// the contract.
		// The EVM code itself will never get executed; see type [magicContract].
		Code:    common.Hex2Bytes("600180808053f3"),
		Storage: map[common.Hash]common.Hash{},
		Balance: nil,
	}

	if createBaseTokenMagicWrap {
		// add the ERC20BaseTokens contract at address 0x10740100...00
		genesisAlloc[iscmagic.ERC20BaseTokensAddress] = core.GenesisAccount{
			Code:    iscmagic.ERC20BaseTokensRuntimeBytecode,
			Storage: map[common.Hash]common.Hash{},
			Balance: nil,
		}
		addToPrivileged(evmPartition, iscmagic.ERC20BaseTokensAddress)
	}

	// add the ERC721NFTs contract at address 0x10740300...00
	genesisAlloc[iscmagic.ERC721NFTsAddress] = core.GenesisAccount{
		Code:    iscmagic.ERC721NFTsRuntimeBytecode,
		Storage: map[common.Hash]common.Hash{},
		Balance: nil,
	}
	addToPrivileged(evmPartition, iscmagic.ERC721NFTsAddress)

	gasLimits := gas.LimitsDefault
	gasRatio := gas.DefaultFeePolicy().EVMGasRatio
	// create the Ethereum genesis block
	emulator.Init(
		evm.EmulatorStateSubrealm(evmPartition),
		evmChainID,
		emulator.GasLimits{
			Block: gas.EVMBlockGasLimit(gasLimits, &gasRatio),
			Call:  gas.EVMCallGasLimit(gasLimits, &gasRatio),
		},
		0,
		genesisAlloc,
	)
}

var errChainIDMismatch = coreerrors.Register("chainId mismatch").Create()

func applyTransaction(ctx isc.Sandbox) dict.Dict {
	// We only want to charge gas for the actual execution of the ethereum tx.
	// ISC magic calls enable gas burning temporarily when called.
	ctx.Privileged().GasBurnEnable(false)
	defer ctx.Privileged().GasBurnEnable(true)

	tx, err := evmtypes.DecodeTransaction(ctx.Params().Get(evm.FieldTransaction))
	ctx.RequireNoError(err)

	ctx.RequireCaller(isc.NewEthereumAddressAgentID(ctx.ChainID(), evmutil.MustGetSender(tx)))

	emu := createEmulator(ctx)

	if tx.Protected() && tx.ChainId().Uint64() != uint64(emu.BlockchainDB().GetChainID()) {
		panic(errChainIDMismatch)
	}

	// Execute the tx in the emulator.
	receipt, result, err := emu.SendTransaction(tx, getTracer(ctx), false)

	// Any gas burned by the EVM is converted to ISC gas units and burned as
	// ISC gas.
	chainInfo := ctx.ChainInfo()
	ctx.Privileged().GasBurnEnable(true)
	burnGasErr := panicutil.CatchPanic(
		func() {
			if result == nil {
				return
			}
			ctx.Gas().Burn(
				gas.BurnCodeEVM1P,
				gas.EVMGasToISC(result.UsedGas, &chainInfo.GasFeePolicy.EVMGasRatio),
			)
		},
	)
	ctx.Privileged().GasBurnEnable(false)

	// if any of these is != nil, the request will be reverted
	revertErr, _ := lo.Find(
		[]error{err, tryGetRevertError(result), burnGasErr},
		func(err error) bool { return err != nil },
	)
	if revertErr != nil {
		// mark receipt as failed
		receipt.Status = types.ReceiptStatusFailed
		// remove any events from the receipt
		receipt.Logs = make([]*types.Log, 0)
		receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
	}

	// amend the gas usage (to include any ISC gas burned in sandbox calls)
	{
		realGasUsed := gas.ISCGasBudgetToEVM(ctx.Gas().Burned(), &chainInfo.GasFeePolicy.EVMGasRatio)
		if realGasUsed > receipt.GasUsed {
			receipt.CumulativeGasUsed += realGasUsed - receipt.GasUsed
			receipt.GasUsed = realGasUsed
		}
	}

	// make sure we always store the EVM tx/receipt in the BlockchainDB, even
	// if the ISC request is reverted
	ctx.Privileged().OnWriteReceipt(func(evmPartition kv.KVStore, _ uint64) {
		saveExecutedTx(evmPartition, chainInfo, tx, receipt)
	})

	// revert the changes in the state / txbuilder in case of error
	ctx.RequireNoError(revertErr)

	return nil
}

var (
	errFoundryNotOwnedByCaller      = coreerrors.Register("foundry with serial number %d not owned by caller")
	errEVMAccountAlreadyExists      = coreerrors.Register("cannot register ERC20NativeTokens contract: EVM account already exists").Create()
	errEVMCanNotDecodeERC27Metadata = coreerrors.Register("cannot decode IRC27 collection NFT metadata")
)

func registerERC20NativeToken(ctx isc.Sandbox) dict.Dict {
	foundrySN := codec.MustDecodeUint32(ctx.Params().Get(evm.FieldFoundrySN))
	name := codec.MustDecodeString(ctx.Params().Get(evm.FieldTokenName))
	tickerSymbol := codec.MustDecodeString(ctx.Params().Get(evm.FieldTokenTickerSymbol))
	decimals := codec.MustDecodeUint8(ctx.Params().Get(evm.FieldTokenDecimals))

	{
		res := ctx.CallView(accounts.Contract.Hname(), accounts.ViewAccountFoundries.Hname(), dict.Dict{
			accounts.ParamAgentID: codec.EncodeAgentID(ctx.Caller()),
		})
		if res[kv.Key(codec.EncodeUint32(foundrySN))] == nil {
			panic(errFoundryNotOwnedByCaller.Create(foundrySN))
		}
	}

	// deploy the contract to the EVM state
	addr := iscmagic.ERC20NativeTokensAddress(foundrySN)
	emu := createEmulator(ctx)
	evmState := emu.StateDB()
	if evmState.Exist(addr) {
		panic(errEVMAccountAlreadyExists)
	}
	evmState.CreateAccount(addr)
	evmState.SetCode(addr, iscmagic.ERC20NativeTokensRuntimeBytecode)
	// see ERC20NativeTokens_storage.json
	evmState.SetState(addr, solidity.StorageSlot(0), solidity.StorageEncodeShortString(name))
	evmState.SetState(addr, solidity.StorageSlot(1), solidity.StorageEncodeShortString(tickerSymbol))
	evmState.SetState(addr, solidity.StorageSlot(2), solidity.StorageEncodeUint8(decimals))

	addToPrivileged(ctx.State(), addr)

	return nil
}

var (
	errTargetMustBeAlias   = coreerrors.Register("target must be alias address")
	errOutputMustBeFoundry = coreerrors.Register("expected foundry output")
)

func registerERC20NativeTokenOnRemoteChain(ctx isc.Sandbox) dict.Dict {
	foundrySN := codec.MustDecodeUint32(ctx.Params().Get(evm.FieldFoundrySN))
	name := codec.MustDecodeString(ctx.Params().Get(evm.FieldTokenName))
	tickerSymbol := codec.MustDecodeString(ctx.Params().Get(evm.FieldTokenTickerSymbol))
	decimals := codec.MustDecodeUint8(ctx.Params().Get(evm.FieldTokenDecimals))
	target := codec.MustDecodeAddress(ctx.Params().Get(evm.FieldTargetAddress))
	if target.Type() != iotago.AddressAlias {
		panic(errTargetMustBeAlias)
	}

	{
		res := ctx.CallView(accounts.Contract.Hname(), accounts.ViewAccountFoundries.Hname(), dict.Dict{
			accounts.ParamAgentID: codec.EncodeAgentID(ctx.Caller()),
		})
		if res[kv.Key(codec.EncodeUint32(foundrySN))] == nil {
			panic(errFoundryNotOwnedByCaller.Create(foundrySN))
		}
	}

	tokenScheme := func() iotago.TokenScheme {
		res := ctx.CallView(accounts.Contract.Hname(), accounts.ViewNativeToken.Hname(), dict.Dict{
			accounts.ParamFoundrySN: codec.EncodeUint32(foundrySN),
		})
		o := codec.MustDecodeOutput(res[accounts.ParamFoundryOutputBin])
		foundryOutput, ok := o.(*iotago.FoundryOutput)
		if !ok {
			panic(errOutputMustBeFoundry)
		}
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
			// FIXME why does this gas budget is higher than the allowance below
			GasBudget: 50 * gas.LimitsDefault.MinGasPerRequest,
		},
	}
	sd := ctx.EstimateRequiredStorageDeposit(req)
	// this request is sent by contract account,
	// so we move enough allowance for the gas fee below in the req.Assets.AddBaseTokens() function call
	ctx.TransferAllowedFunds(ctx.AccountID(), isc.NewAssetsBaseTokens(sd+10*gas.LimitsDefault.MinGasPerRequest))
	req.Assets.AddBaseTokens(sd + 10*gas.LimitsDefault.MinGasPerRequest)
	ctx.Send(req)

	return nil
}

var (
	errSenderMustBeAlias            = coreerrors.Register("sender must be alias address").Create()
	errFoundryMustBeOffChain        = coreerrors.Register("foundry must be off-chain").Create()
	errNativeTokenAlreadyRegistered = coreerrors.Register("native token already registered").Create()
)

func registerERC20ExternalNativeToken(ctx isc.Sandbox) dict.Dict {
	caller, ok := ctx.Caller().(*isc.ContractAgentID)
	if !ok {
		panic(errSenderMustBeAlias)
	}
	if ctx.ChainID().Equals(caller.ChainID()) {
		panic(errFoundryMustBeOffChain)
	}
	alias := caller.ChainID().AsAliasAddress()

	name := codec.MustDecodeString(ctx.Params().Get(evm.FieldTokenName))
	tickerSymbol := codec.MustDecodeString(ctx.Params().Get(evm.FieldTokenTickerSymbol))
	decimals := codec.MustDecodeUint8(ctx.Params().Get(evm.FieldTokenDecimals))

	// TODO: We should somehow inspect the real FoundryOutput, but it is on L1.
	// Here we reproduce it from the given params (which we assume to be correct)
	// in order to derive the FoundryID
	foundrySN := codec.MustDecodeUint32(ctx.Params().Get(evm.FieldFoundrySN))
	tokenScheme := codec.MustDecodeTokenScheme(ctx.Params().Get(evm.FieldFoundryTokenScheme))
	simpleTS, ok := tokenScheme.(*iotago.SimpleTokenScheme)
	if !ok {
		panic(errUnsupportedTokenScheme)
	}
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
	if ok {
		panic(errNativeTokenAlreadyRegistered)
	}

	emu := createEmulator(ctx)
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
	nativeTokenID := codec.MustDecodeNativeTokenID(ctx.Params().Get(evm.FieldNativeTokenID))
	addr, ok := getERC20ExternalNativeTokensAddress(ctx, nativeTokenID)
	if !ok {
		return nil
	}
	return result(addr[:])
}

func registerERC721NFTCollection(ctx isc.Sandbox) dict.Dict {
	collectionID := codec.MustDecodeNFTID(ctx.Params().Get(evm.FieldNFTCollectionID))

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

	RegisterERC721NFTCollectionByNFTId(ctx.State(), collection)

	return nil
}

func getChainID(ctx isc.SandboxView) dict.Dict {
	chainID := emulator.GetChainIDFromBlockChainDBState(
		emulator.BlockchainDBSubrealmR(
			evm.EmulatorStateSubrealmR(ctx.StateR()),
		),
	)
	return result(codec.EncodeUint16(chainID))
}

// include the revert reason in the error
func tryGetRevertError(res *core.ExecutionResult) error {
	if res == nil {
		return nil
	}
	if res.Err == nil {
		return nil
	}
	if len(res.Revert()) > 0 {
		return vm.ErrEVMExecutionReverted.Create(hex.EncodeToString(res.Revert()))
	}
	return res.Err
}

// callContract is called from the jsonrpc eth_estimateGas and eth_call endpoints.
// The VM is in estimate gas mode, and any state mutations are discarded.
func callContract(ctx isc.Sandbox) dict.Dict {
	// We only want to charge gas for the actual execution of the ethereum tx.
	// ISC magic calls enable gas burning temporarily when called.
	ctx.Privileged().GasBurnEnable(false)
	defer ctx.Privileged().GasBurnEnable(true)

	callMsg, err := evmtypes.DecodeCallMsg(ctx.Params().Get(evm.FieldCallMsg))
	ctx.RequireNoError(err)
	ctx.RequireCaller(isc.NewEthereumAddressAgentID(ctx.ChainID(), callMsg.From))

	emu := createEmulator(ctx)
	res, err := emu.CallContract(callMsg, ctx.Gas().EstimateGasMode())
	ctx.RequireNoError(err)
	ctx.RequireNoError(tryGetRevertError(res))

	gasRatio := getEVMGasRatio(ctx)
	{
		// burn the used EVM gas as it would be done for a normal request call
		ctx.Privileged().GasBurnEnable(true)
		gasErr := panicutil.CatchPanic(
			func() {
				if res != nil {
					ctx.Gas().Burn(gas.BurnCodeEVM1P, gas.EVMGasToISC(res.UsedGas, &gasRatio))
				}
			},
		)
		ctx.Privileged().GasBurnEnable(false)
		ctx.RequireNoError(gasErr)
	}

	return result(res.ReturnData)
}

func getEVMGasRatio(ctx isc.SandboxBase) util.Ratio32 {
	gasRatioViewRes := ctx.CallView(governance.Contract.Hname(), governance.ViewGetEVMGasRatio.Hname(), nil)
	return codec.MustDecodeRatio32(gasRatioViewRes.Get(governance.ParamEVMGasRatio), gas.DefaultEVMGasRatio)
}

func newL1Deposit(ctx isc.Sandbox) dict.Dict {
	// can only be called from the accounts contract
	ctx.RequireCaller(isc.NewContractAgentID(ctx.ChainID(), accounts.Contract.Hname()))
	params := ctx.Params()
	agentIDBytes := params.MustGetBytes(evm.FieldAgentIDDepositOriginator)
	targetAddress := common.BytesToAddress(params.MustGetBytes(evm.FieldAddress))
	assets, err := isc.AssetsFromBytes(params.MustGetBytes(evm.FieldAssets))
	ctx.RequireNoError(err, "unable to parse assets from params")

	// create a fake tx so that the deposit is visible by the EVM
	// discard remainder in decimals conversion
	wei := util.BaseTokensDecimalsToEthereumDecimals(assets.BaseTokens, newEmulatorContext(ctx).BaseTokensDecimals())
	nonce := uint64(0)
	// encode the txdata as <AgentID sender>+<Assets>+[blockIndex + reqIndex]
	// the last part [ ] is needed so we don't produce txs with colliding hashes in the same or different blocks.
	txData := []byte{}
	txData = append(txData, agentIDBytes...)
	txData = append(txData, assets.Bytes()...)
	txData = append(txData, codec.Encode(ctx.StateAnchor().StateIndex+1)...)
	txData = append(txData, codec.Encode(ctx.RequestIndex())...)
	chainInfo := ctx.ChainInfo()
	gasPrice := chainInfo.GasFeePolicy.DefaultGasPriceFullDecimals(parameters.L1().BaseToken.Decimals)
	tx := types.NewTx(
		&types.LegacyTx{
			Nonce:    nonce,
			To:       &targetAddress,
			Value:    wei,
			Gas:      0,
			GasPrice: gasPrice,
			Data:     txData,
		},
	)

	fromAddress := common.Address{}
	logs := make([]*types.Log, 0)
	for _, nt := range assets.NativeTokens {
		if nt.Amount.Sign() == 0 {
			continue
		}
		// emit a Transfer event from the ERC20NativeTokens / ERC20ExternalNativeTokens contract
		erc20Address, ok := findERC20NativeTokenContractAddress(ctx, nt.ID)
		if !ok {
			continue
		}
		logs = append(logs, makeTransferEventERC20(erc20Address, fromAddress, targetAddress, nt.Amount))
	}
	for _, nftID := range assets.NFTs {
		// if the NFT belongs to a collection, emit a Transfer event from the corresponding ERC721NFTCollection contract
		if nft := ctx.GetNFTData(nftID); nft != nil {
			if collectionNFTAddress, ok := nft.Issuer.(*iotago.NFTAddress); ok {
				collectionID := collectionNFTAddress.NFTID()
				erc721CollectionContractAddress := iscmagic.ERC721NFTCollectionAddress(collectionID)
				stateDB := emulator.NewStateDB(newEmulatorContext(ctx))
				if stateDB.Exist(erc721CollectionContractAddress) {
					logs = append(logs, makeTransferEventERC721(erc721CollectionContractAddress, fromAddress, targetAddress, iscmagic.WrapNFTID(nftID).TokenID()))
					continue
				}
			}
		}
		// otherwise, emit a Transfer event from the ERC721NFTs contract
		logs = append(logs, makeTransferEventERC721(iscmagic.ERC721NFTsAddress, fromAddress, targetAddress, iscmagic.WrapNFTID(nftID).TokenID()))
	}

	receipt := &types.Receipt{
		Type:   types.LegacyTxType,
		Logs:   logs,
		Status: types.ReceiptStatusSuccessful,
	}
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	ctx.Privileged().OnWriteReceipt(func(evmPartition kv.KVStore, gasBurned uint64) {
		receipt.GasUsed = gas.ISCGasBurnedToEVM(gasBurned, &chainInfo.GasFeePolicy.EVMGasRatio)
		receipt.CumulativeGasUsed = createBlockchainDB(evmPartition, chainInfo).GetPendingCumulativeGasUsed() + receipt.GasUsed
		createBlockchainDB(evmPartition, ctx.ChainInfo()).AddTransaction(tx, receipt)
	})

	return nil
}

func newL1Withdrawal(ctx isc.Sandbox) dict.Dict {
	// can only be called from the accounts contract
	ctx.RequireCaller(isc.NewContractAgentID(ctx.ChainID(), accounts.Contract.Hname()))
	params := ctx.Params()
	targetAddressBytes := params.MustGetBytes(evm.FieldAgentIDWithdrawalTarget)
	senderAddress := common.BytesToAddress(params.MustGetBytes(evm.FieldAddress))
	assets, err := isc.AssetsFromBytes(params.MustGetBytes(evm.FieldAssets))
	ctx.RequireNoError(err, "unable to parse assets from params")

	// create a fake tx so that the withdrawal is visible by the EVM
	wei := util.BaseTokensDecimalsToEthereumDecimals(assets.BaseTokens, newEmulatorContext(ctx).BaseTokensDecimals())
	nonce := uint64(0)
	// encode the txdata as <target>+<Assets>+[blockIndex + reqIndex]
	// the last part [ ] is needed so we don't produce txs with colliding hashes in the same or different blocks.
	txData := []byte{}
	txData = append(txData, targetAddressBytes...)
	txData = append(txData, assets.Bytes()...)
	txData = append(txData, codec.Encode(ctx.StateAnchor().StateIndex+1)...)
	txData = append(txData, codec.Encode(ctx.RequestIndex())...)
	chainInfo := ctx.ChainInfo()
	gasPrice := chainInfo.GasFeePolicy.DefaultGasPriceFullDecimals(parameters.L1().BaseToken.Decimals)
	toAddress := common.Address{}
	tx := types.NewTx(
		&types.LegacyTx{
			Nonce:    nonce,
			To:       &toAddress,
			Value:    wei,
			Gas:      0,
			GasPrice: gasPrice,
			Data:     txData,
		},
	)

	logs := make([]*types.Log, 0)
	for _, nt := range assets.NativeTokens {
		if nt.Amount.Sign() == 0 {
			continue
		}
		// emit a Transfer event from the ERC20NativeTokens / ERC20ExternalNativeTokens contract
		erc20Address, ok := findERC20NativeTokenContractAddress(ctx, nt.ID)
		if !ok {
			continue
		}
		logs = append(logs, makeTransferEvent(erc20Address, senderAddress, toAddress, nt.Amount))
	}
	for _, nftID := range assets.NFTs {
		// if the NFT belongs to a collection, emit a Transfer event from the corresponding ERC721NFTCollection contract
		if nft := ctx.GetNFTData(nftID); nft != nil {
			if collectionNFTAddress, ok := nft.Issuer.(*iotago.NFTAddress); ok {
				collectionID := collectionNFTAddress.NFTID()
				erc721CollectionContractAddress := iscmagic.ERC721NFTCollectionAddress(collectionID)
				stateDB := emulator.NewStateDB(newEmulatorContext(ctx))
				if stateDB.Exist(erc721CollectionContractAddress) {
					logs = append(logs, makeTransferEvent(erc721CollectionContractAddress, senderAddress, toAddress, iscmagic.WrapNFTID(nftID).TokenID()))
					continue
				}
			}
		}
		// otherwise, emit a Transfer event from the ERC721NFTs contract
		logs = append(logs, makeTransferEvent(iscmagic.ERC721NFTsAddress, senderAddress, toAddress, iscmagic.WrapNFTID(nftID).TokenID()))
	}

	if len(logs) == 0 {
		// no need for the fake tx with wei == 0 and no events
		return nil
	}

	receipt := &types.Receipt{
		Type:   types.LegacyTxType,
		Logs:   logs,
		Status: types.ReceiptStatusSuccessful,
	}
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	ctx.Privileged().OnWriteReceipt(func(evmPartition kv.KVStore, gasBurned uint64) {
		receipt.GasUsed = gas.ISCGasBurnedToEVM(gasBurned, &chainInfo.GasFeePolicy.EVMGasRatio)
		receipt.CumulativeGasUsed = createBlockchainDB(evmPartition, chainInfo).GetPendingCumulativeGasUsed() + receipt.GasUsed
		createBlockchainDB(evmPartition, ctx.ChainInfo()).AddTransaction(tx, receipt)
	})
	return nil
}

func newTransferBetweenL2Accounts(ctx isc.Sandbox) dict.Dict {
	// can only be called from the accounts contract
	ctx.RequireCaller(isc.NewContractAgentID(ctx.ChainID(), accounts.Contract.Hname()))
	params := ctx.Params()
	fromAddress := common.BytesToAddress(params.MustGetBytes(evm.FieldFromAddress))
	toAddress := common.BytesToAddress(params.MustGetBytes(evm.FieldToAddress))
	assets, err := isc.AssetsFromBytes(params.MustGetBytes(evm.FieldAssets))
	ctx.RequireNoError(err, "unable to parse assets from params")

	// create a fake tx so that the operation is visible by the EVM
	wei := util.BaseTokensDecimalsToEthereumDecimals(assets.BaseTokens, newEmulatorContext(ctx).BaseTokensDecimals())
	nonce := uint64(0)
	// encode the txdata as <from>+<Assets>+[blockIndex + reqIndex]
	// the last part [ ] is needed so we don't produce txs with colliding hashes in the same or different blocks.
	txData := []byte{}
	txData = append(txData, fromAddress.Bytes()...)
	txData = append(txData, assets.Bytes()...)
	txData = append(txData, codec.Encode(ctx.StateAnchor().StateIndex+1)...)
	txData = append(txData, codec.Encode(ctx.RequestIndex())...)
	chainInfo := ctx.ChainInfo()
	gasPrice := chainInfo.GasFeePolicy.DefaultGasPriceFullDecimals(parameters.L1().BaseToken.Decimals)
	tx := types.NewTx(
		&types.LegacyTx{
			Nonce:    nonce,
			To:       &toAddress,
			Value:    wei,
			Gas:      0,
			GasPrice: gasPrice,
			Data:     txData,
		},
	)

	logs := make([]*types.Log, 0)
	for _, nt := range assets.NativeTokens {
		if nt.Amount.Sign() == 0 {
			continue
		}
		// emit a Transfer event from the ERC20NativeTokens / ERC20ExternalNativeTokens contract
		erc20Address, ok := findERC20NativeTokenContractAddress(ctx, nt.ID)
		if !ok {
			continue
		}
		logs = append(logs, makeTransferEvent(erc20Address, fromAddress, toAddress, nt.Amount))
	}
	for _, nftID := range assets.NFTs {
		// if the NFT belongs to a collection, emit a Transfer event from the corresponding ERC721NFTCollection contract
		if nft := ctx.GetNFTData(nftID); nft != nil {
			if collectionNFTAddress, ok := nft.Issuer.(*iotago.NFTAddress); ok {
				collectionID := collectionNFTAddress.NFTID()
				erc721CollectionContractAddress := iscmagic.ERC721NFTCollectionAddress(collectionID)
				stateDB := emulator.NewStateDB(newEmulatorContext(ctx))
				if stateDB.Exist(erc721CollectionContractAddress) {
					logs = append(logs, makeTransferEvent(erc721CollectionContractAddress, fromAddress, toAddress, iscmagic.WrapNFTID(nftID).TokenID()))
					continue
				}
			}
		}
		// otherwise, emit a Transfer event from the ERC721NFTs contract
		logs = append(logs, makeTransferEvent(iscmagic.ERC721NFTsAddress, fromAddress, toAddress, iscmagic.WrapNFTID(nftID).TokenID()))
	}

	if len(logs) == 0 {
		// no need for the fake tx with wei == 0 and no events
		return nil
	}

	receipt := &types.Receipt{
		Type:   types.LegacyTxType,
		Logs:   logs,
		Status: types.ReceiptStatusSuccessful,
	}
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	ctx.Privileged().OnWriteReceipt(func(evmPartition kv.KVStore, gasBurned uint64) {
		receipt.GasUsed = gas.ISCGasBurnedToEVM(gasBurned, &chainInfo.GasFeePolicy.EVMGasRatio)
		receipt.CumulativeGasUsed = createBlockchainDB(evmPartition, chainInfo).GetPendingCumulativeGasUsed() + receipt.GasUsed
		createBlockchainDB(evmPartition, ctx.ChainInfo()).AddTransaction(tx, receipt)
	})
	return nil
}

var transferEventTopic = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

func makeTransferEventERC20(contractAddress, from, to common.Address, uint256Data *big.Int) *types.Log {
	return &types.Log{
		Address: contractAddress,
		Topics: []common.Hash{
			transferEventTopic,
			evmutil.AddressToIndexedTopic(from),
			evmutil.AddressToIndexedTopic(to),
		},
		Data: lo.Must((abi.Arguments{{Type: lo.Must(abi.NewType("uint256", "", nil))}}).Pack(uint256Data)),
	}
}

func makeTransferEventERC721(contractAddress, from, to common.Address, uint256Data *big.Int) *types.Log {
	tokenIDPacked := lo.Must((abi.Arguments{{Type: lo.Must(abi.NewType("uint256", "", nil))}}).Pack(uint256Data))
	var tokenIDTopic common.Hash
	copy(tokenIDTopic[:], tokenIDPacked) // same len

	return &types.Log{
		Address: contractAddress,
		Topics: []common.Hash{
			transferEventTopic, // event topic
			{},                 // indexed `from` address
			addrToTopic,        // indexed `to` address
			tokenIDTopic,       // indexed `tokenId`
		},
	}
}
