// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/evm/evmutil"
	"github.com/iotaledger/wasp/v2/packages/evm/solidity"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/util/panicutil"
	iscvm "github.com/iotaledger/wasp/v2/packages/vm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm/iscmagic"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations/legacy"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

var Processor = evm.Contract.Processor(nil,
	evm.FuncSendTransaction.WithHandler(applyTransaction),
	evm.FuncCallContract.WithHandler(callContract),
	evm.FuncRegisterERC20Coin.WithHandler(registerERC20Coin),
	evm.FuncNewL1Deposit.WithHandler(newL1Deposit),

	// views
	evm.ViewGetChainID.WithHandler(getChainID),
)

// SetInitialState initializes the evm core contract and the Ethereum genesis
// block on a newly created ISC chain.
func SetInitialState(
	evmPartition kv.KVStore,
	evmChainID uint16,
	feePolicy *gas.FeePolicy,
	genesis *core.Genesis,
) {
	// Build base alloc with ISC magic contract
	genesisAlloc := types.GenesisAlloc{
		iscmagic.Address: types.Account{
			// Dummy code for size checks; never executed.
			Code:    common.Hex2Bytes("600180808053f3"),
			Storage: map[common.Hash]common.Hash{},
			Balance: nil,
		},
	}

	// Select gas ratio from fee policy (default if nil)
	var ratio util.Ratio32
	if feePolicy != nil {
		ratio = feePolicy.EVMGasRatio
	} else {
		ratio = gas.DefaultFeePolicy().EVMGasRatio
	}

	// Base gas limits
	gasLimits := gas.LimitsDefault
	evmGasLimit := emulator.GasLimits{
		Block: gas.EVMBlockGasLimit(gasLimits, &ratio),
		Call:  gas.EVMCallGasLimit(gasLimits, &ratio),
	}

	// Decide timestamp, gas limit override, and custom flag
	var ts uint64
	customGenesis := false
	if genesis != nil {
		ts = genesis.Timestamp
		customGenesis = true
		if genesis.GasLimit != 0 {
			evmGasLimit.Block = genesis.GasLimit
		}
	}

	emulator.Init(
		evm.EmulatorStateSubrealm(evmPartition),
		evmChainID,
		evmGasLimit,
		ts,
		genesisAlloc,
		customGenesis,
	)
}

var errChainIDMismatch = coreerrors.Register("chainId mismatch").Create()

func applyTransaction(ctx isc.Sandbox, tx *types.Transaction) {
	cannotBeCalledFromContracts(ctx)

	// We only want to charge gas for the actual execution of the ethereum tx.
	// ISC magic calls enable gas burning temporarily when called.
	ctx.Privileged().GasBurnEnable(false)
	defer ctx.Privileged().GasBurnEnable(true)

	ctx.RequireCaller(isc.NewEthereumAddressAgentID(evmutil.MustGetSender(tx)))

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
		receipt.Bloom = types.CreateBloom(receipt)
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
	ctx.Privileged().OnWriteReceipt(func(evmPartition kv.KVStore, _ uint64, _ *isc.VMError) {
		saveExecutedTx(evmPartition, chainInfo, tx, receipt)
	})

	// revert the changes in the state / txbuilder in case of error
	ctx.RequireNoError(revertErr)
}

var (
	errEVMAccountAlreadyExists = coreerrors.Register("cannot register ERC20Coin contract: EVM account already exists").Create()
	errUnknownCoin             = coreerrors.Register("unknown coin")
)

func registerERC20Coin(ctx isc.Sandbox, coinType coin.Type) {
	// deploy the contract to the EVM state
	addr := iscmagic.ERC20CoinAddress(coinType)
	emu := createEmulator(ctx)
	evmState := emu.StateDB()
	if evmState.Exist(addr) {
		panic(errEVMAccountAlreadyExists)
	}
	evmState.CreateAccount(addr)
	evmState.SetCode(addr, iscmagic.ERC20CoinRuntimeBytecode)
	// see ERC20Coin_storage.json
	info, ok := ctx.GetCoinInfo(coinType)
	if !ok {
		panic(errUnknownCoin)
	}
	for k, v := range solidity.StorageEncodeString(0, coinType.String()) {
		evmState.SetState(addr, k, v)
	}
	evmState.SetState(addr, solidity.StorageSlot(1), solidity.StorageEncodeShortString(info.Name))
	evmState.SetState(addr, solidity.StorageSlot(2), solidity.StorageEncodeShortString(info.Symbol))
	evmState.SetState(addr, solidity.StorageSlot(3), solidity.StorageEncodeUint8(info.Decimals))

	addToPrivileged(ctx.State(), addr)
}

func getChainID(ctx isc.SandboxView) uint16 {
	return emulator.GetChainIDFromBlockChainDBState(
		emulator.BlockchainDBSubrealmR(
			evm.EmulatorStateSubrealmR(ctx.StateR()),
		),
	)
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
		return iscvm.ErrEVMExecutionReverted.Create(hex.EncodeToString(res.Revert()))
	}
	return res.Err
}

// callContract is called from the jsonrpc eth_estimateGas and eth_call endpoints.
// The VM is in estimate gas mode, and any state mutations are discarded.
func callContract(ctx isc.Sandbox, callMsg ethereum.CallMsg) []byte {
	cannotBeCalledFromContracts(ctx)

	// We only want to charge gas for the actual execution of the ethereum tx.
	// ISC magic calls enable gas burning temporarily when called.
	ctx.Privileged().GasBurnEnable(false)
	defer ctx.Privileged().GasBurnEnable(true)

	ctx.RequireCaller(isc.NewEthereumAddressAgentID(callMsg.From))

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

	return res.ReturnData
}

func getEVMGasRatio(ctx isc.SandboxBase) util.Ratio32 {
	res := ctx.CallView(governance.ViewGetEVMGasRatio.Message())
	gasRatioViewRes, err := governance.ViewGetEVMGasRatio.DecodeOutput(res)
	ctx.RequireNoError(err)
	return gasRatioViewRes
}

func newL1Deposit(ctx isc.Sandbox, l1DepositOriginatorBytes isc.AgentID, toAddress common.Address, assets *isc.Assets) {
	// can only be called from the accounts contract
	ctx.RequireCaller(isc.NewContractAgentID(accounts.Contract.Hname()))
	// create a fake tx so that the operation is visible by the EVM
	AddDummyTxWithTransferEvents(ctx, toAddress, assets, legacy.AgentIDToBytes(ctx.SchemaVersion(), l1DepositOriginatorBytes), true)
}

func AddDummyTxWithTransferEvents(
	ctx isc.Sandbox,
	toAddress common.Address,
	assets *isc.Assets,
	txData []byte,
	isInRequestContext bool,
) {
	zeroAddress := common.Address{}
	logs := makeTransferEvents(ctx, zeroAddress, toAddress, assets)

	wei := util.BaseTokensDecimalsToEthereumDecimals(assets.BaseTokens(), newEmulatorContext(ctx).BaseTokensDecimals())
	if wei.Sign() == 0 && len(logs) == 0 {
		return
	}

	nonce := uint64(0)
	chainInfo := ctx.ChainInfo()
	gasPrice := chainInfo.GasFeePolicy.DefaultGasPriceFullDecimals(parameters.BaseTokenDecimals)

	// txData = txData+<assets>+[blockIndex + reqIndex]
	// the last part [ ] is needed so we don't produce txs with colliding hashes in the same or different blocks.
	txData = append(txData, legacy.AssetsToBytes(ctx.SchemaVersion(), assets)...)
	txData = append(txData, legacy.EncodeUint32ForDummyTX(ctx.SchemaVersion(), ctx.StateIndex())...)
	txData = append(txData, legacy.EncodeUint16ForDummyTX(ctx.SchemaVersion(), ctx.RequestIndex())...)

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

	receipt := &types.Receipt{
		Type:   types.LegacyTxType,
		Logs:   logs,
		Status: types.ReceiptStatusSuccessful,
	}
	receipt.Bloom = types.CreateBloom(receipt)

	callTracerHooks := func() {
		tracer := ctx.EVMTracer()
		if tracer != nil {
			if tracer.OnTxStart != nil {
				tracer.OnTxStart(nil, tx, common.Address{})
			}
			if tracer.OnEnter != nil {
				tracer.OnEnter(0, byte(vm.CALL), common.Address{}, toAddress, txData, 0, wei)
			}
			if tracer.OnLog != nil {
				for _, log := range logs {
					tracer.OnLog(log)
				}
			}
			if tracer.OnExit != nil {
				tracer.OnExit(0, nil, receipt.GasUsed, nil, false)
			}
			if tracer.OnTxEnd != nil {
				tracer.OnTxEnd(receipt, nil)
			}
		}
	}

	if !isInRequestContext {
		// called from outside vmrun, just add the tx without a gas value
		createBlockchainDB(ctx.State(), chainInfo).AddTransaction(tx, receipt)
		callTracerHooks()
		return
	}

	ctx.Privileged().OnWriteReceipt(func(evmPartition kv.KVStore, gasBurned uint64, vmError *isc.VMError) {
		if vmError != nil {
			return // do not issue deposit event if execution failed
		}
		receipt.GasUsed = gas.ISCGasBurnedToEVM(gasBurned, &chainInfo.GasFeePolicy.EVMGasRatio)
		blockchainDB := createBlockchainDB(evmPartition, chainInfo)
		receipt.CumulativeGasUsed = blockchainDB.GetPendingCumulativeGasUsed() + receipt.GasUsed
		blockchainDB.AddTransaction(tx, receipt)
		callTracerHooks()
	})
}

func makeTransferEvents(
	ctx isc.Sandbox,
	fromAddress, toAddress common.Address,
	assets *isc.Assets,
) []*types.Log {
	logs := make([]*types.Log, 0)
	stateDB := emulator.NewStateDB(newEmulatorContext(ctx))
	for coinType, value := range assets.Coins.Iterate() {
		if value == 0 {
			continue
		}
		// emit a Transfer event from the ERC20Coin / ERC20ExternalCoins contract
		erc20Address := iscmagic.ERC20CoinAddress(coinType)
		if stateDB.Exist(erc20Address) {
			logs = append(logs, makeTransferEventERC20(erc20Address, fromAddress, toAddress, value))
		}
	}
	return logs
}

var transferEventTopic = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

func makeTransferEventERC20(contractAddress, from, to common.Address, amount coin.Value) *types.Log {
	return &types.Log{
		Address: contractAddress,
		Topics: []common.Hash{
			transferEventTopic,
			evmutil.AddressToIndexedTopic(from),
			evmutil.AddressToIndexedTopic(to),
		},
		Data: evmutil.PackUint256(new(big.Int).SetUint64(uint64(amount))),
	}
}
