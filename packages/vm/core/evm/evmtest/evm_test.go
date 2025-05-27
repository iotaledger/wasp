// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmerrors"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/packages/testutil/testdbhash"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/contracts/inccounter"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func TestStorageContract(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.EthereumAccountByIndexWithL2Funds(0)
	require.EqualValues(t, 1, env.getBlockNumber()) // evm block number is incremented along with ISC block index

	// deploy solidity `storage` contract
	storage := env.deployStorageContract(ethKey)
	require.EqualValues(t, 2, env.getBlockNumber())

	// call FuncCallView to call EVM contract's `retrieve` view, get 42
	require.EqualValues(t, 42, storage.retrieve())

	// call FuncSendTransaction with EVM tx that calls `store(43)`
	res, err := storage.store(43)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, res.EVMReceipt.Status)

	// call `retrieve` view, get 43
	require.EqualValues(t, 43, storage.retrieve())

	for _, v := range []uint32{44, 45, 46} {
		_, err = storage.store(v)
		require.NoError(t, err)
	}
	// try view call with blockNumber = -1 (latest block)
	{
		var v uint32
		require.NoError(t, storage.callView("retrieve", nil, &v, rpc.BlockNumberOrHashWithNumber(rpc.LatestBlockNumber)))
		require.EqualValues(t, 46, v)
	}
	// TODO: try the view call explicitly passing the EVM block
	// blockNumber := rpc.BlockNumber(env.getBlockNumber())
	// for _, i := range []uint32{0, 1, 2, 3} {
	// 	var v uint32
	// 	bn := blockNumber + rpc.BlockNumber(i)
	// 	require.NoErrorf(t, storage.callView("retrieve", nil, &v, rpc.BlockNumberOrHashWithNumber(bn)), "cannot fetch block %d", bn)
	// 	require.EqualValuesf(t, 43+i, v, "blockNumber %d should have counter=%d, got=%d", bn, 43+i, v)
	// }

	testdbhash.VerifyContractStateHash(env.solo, evm.Contract, "", t.Name())
}

func TestLowLevelCallRevert(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

	contract := env.DeployContract(ethKey, evmtest.RevertTestContractABI, evmtest.RevertTestContractBytecode)

	getCount := func() uint32 {
		var v uint32
		require.NoError(t, contract.callView("count", nil, &v))
		return v
	}

	require.Equal(t, uint32(0), getCount())

	_, err := contract.CallFn([]ethCallOptions{{gasLimit: 200000}}, "selfCallRevert")
	require.NoError(t, err)
	require.Equal(t, uint32(0), getCount())
}

func TestERC20Contract(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

	// deploy solidity `erc20` contract
	erc20 := env.deployERC20Contract(ethKey, "TestCoin", "TEST")

	// call `totalSupply` view
	{
		v := erc20.totalSupply()
		// 100 * 10^18
		expected := new(big.Int).Mul(big.NewInt(100), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
		require.Zero(t, v.Cmp(expected))
	}

	_, recipientAddress := generateEthereumKey(t)
	transferAmount := big.NewInt(1337)

	// call `transfer` => send 1337 TestCoin to recipientAddress
	res, err := erc20.transfer(recipientAddress, transferAmount)
	require.NoError(t, err)

	require.Equal(t, types.ReceiptStatusSuccessful, res.EVMReceipt.Status)
	require.Equal(t, 1, len(res.EVMReceipt.Logs))

	// call `balanceOf` view => check balance of recipient = 1337 TestCoin
	require.Zero(t, erc20.balanceOf(recipientAddress).Cmp(transferAmount))
}

func TestGetCode(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	erc20 := env.deployERC20Contract(ethKey, "TestCoin", "TEST")

	// get contract bytecode from EVM emulator
	retrievedBytecode := env.getCode(erc20.address)

	// ensure returned bytecode matches the expected runtime bytecode
	require.True(t, bytes.Equal(retrievedBytecode, evmtest.ERC20ContractRuntimeBytecode), "bytecode retrieved from the chain must match the deployed bytecode")
}

func TestGasCharged(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	storage := env.deployStorageContract(ethKey)

	// call `store(999)` with enough gas
	res, err := storage.store(999)
	require.NoError(t, err)
	t.Log("evm gas used:", res.EVMReceipt.GasUsed)
	t.Log("isc gas used:", res.ISCReceipt.GasBurned)
	t.Log("isc gas fee:", res.ISCReceipt.GasFeeCharged)
	require.Greater(t, res.EVMReceipt.GasUsed, uint64(0))
	require.Greater(t, res.ISCReceipt.GasBurned, uint64(0))
	require.Greater(t, res.ISCReceipt.GasFeeCharged, uint64(0))
}

func TestGasRatio(t *testing.T) {
	env := InitEVM(t)
	env.Chain.MustDepositBaseTokensToL2(solo.DefaultChainAdminBaseTokens, nil)

	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	storage := env.deployStorageContract(ethKey)

	require.Equal(t, gas.DefaultEVMGasRatio, env.getEVMGasRatio())

	res, err := storage.store(43)
	require.NoError(t, err)
	initialGasFee := res.ISCReceipt.GasFeeCharged

	// only the owner can call the setEVMGasRatio endpoint
	newGasRatio := util.Ratio32{A: gas.DefaultEVMGasRatio.A * 10, B: gas.DefaultEVMGasRatio.B}
	newUserWallet, _ := env.solo.NewKeyPairWithFunds()
	env.Chain.DepositBaseTokensToL2(1*isc.Million, newUserWallet)
	err = env.setEVMGasRatio(newGasRatio, iscCallOptions{wallet: newUserWallet})
	require.True(t, isc.VMErrorIs(err, vm.ErrUnauthorized))
	require.Equal(t, gas.DefaultEVMGasRatio, env.getEVMGasRatio())

	// current owner is able to set a new gasRatio
	err = env.setEVMGasRatio(newGasRatio, iscCallOptions{wallet: env.Chain.ChainAdmin})
	require.NoError(t, err)
	require.Equal(t, newGasRatio, env.getEVMGasRatio())

	// run an equivalent request and compare the gas fees
	res, err = storage.store(44)
	require.NoError(t, err)
	require.Greater(t, res.ISCReceipt.GasFeeCharged, initialGasFee)
}

// tests that the gas limits are correctly enforced based on the base tokens sent
func TestGasLimit(t *testing.T) {
	env := InitEVM(t)
	env.Chain.MustDepositBaseTokensToL2(solo.DefaultChainAdminBaseTokens, nil)

	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	storage := env.deployStorageContract(ethKey)

	// set a gas ratio such that evm gas cost in base tokens is larger than storage deposit cost
	err := env.setEVMGasRatio(util.Ratio32{A: 10, B: 1}, iscCallOptions{wallet: env.Chain.ChainAdmin})
	require.NoError(t, err)

	// estimate gas by sending a valid tx
	result, err := storage.store(123)
	require.NoError(t, err)
	gasBurned := result.ISCReceipt.GasBurned
	fee := result.ISCReceipt.GasFeeCharged
	t.Logf("gas: %d, fee: %d", gasBurned, fee)

	// send again with same gas limit but not enough base tokens
	notEnoughBaseTokensForGas := fee * 9 / 10
	ethKey2, _ := env.Chain.NewEthereumAccountWithL2Funds(notEnoughBaseTokensForGas)
	_, err = storage.store(124, ethCallOptions{sender: ethKey2})
	require.Error(t, err)
	require.Regexp(t, `\bgas\b`, err.Error())
}

func TestNotEnoughISCGas(t *testing.T) {
	env := InitEVM(t)
	env.Chain.MustDepositBaseTokensToL2(solo.DefaultChainAdminBaseTokens, nil)

	ethKey, ethAddress := env.Chain.NewEthereumAccountWithL2Funds()
	storage := env.deployStorageContract(ethKey)

	_, err := storage.store(43)
	require.NoError(t, err)

	// only the owner can call the setEVMGasRatio endpoint
	// set the ISC gas ratio VERY HIGH
	newGasRatio := util.Ratio32{A: gas.DefaultEVMGasRatio.A * 500, B: gas.DefaultEVMGasRatio.B}
	err = env.setEVMGasRatio(newGasRatio, iscCallOptions{wallet: env.Chain.ChainAdmin})
	require.NoError(t, err)
	require.Equal(t, newGasRatio, env.getEVMGasRatio())

	senderAddress := crypto.PubkeyToAddress(storage.defaultSender.PublicKey)
	nonce := env.getNonce(senderAddress)

	// try to issue a call to store(something) in EVM
	res, err := storage.store(44, ethCallOptions{
		gasLimit: 21204, // provide a gas limit because the estimation will fail
	})

	// the call must fail with "not enough gas"
	require.Error(t, err)
	require.Regexp(t, "out of gas", err)
	require.Equal(t, nonce+1, env.getNonce(senderAddress))

	// there must be an EVM receipt
	require.NotNil(t, res.EVMReceipt)
	require.Equal(t, res.EVMReceipt.Status, types.ReceiptStatusFailed)

	// no changes should persist
	require.EqualValues(t, 43, storage.retrieve())

	// check nonces are still in sync
	iscNonce := env.Chain.Nonce(isc.NewEthereumAddressAgentID(ethAddress))
	evmNonce := env.getNonce(ethAddress)
	require.EqualValues(t, iscNonce, evmNonce)
}

// ensure the amount of base tokens sent impacts the amount of gas used
func TestLoop(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	loop := env.deployLoopContract(ethKey)

	gasRatio := env.getEVMGasRatio()

	for _, gasLimit := range []uint64{200000, 400000} {
		baseTokensSent := coin.Value(gas.EVMGasToISC(gasLimit, &gasRatio))
		ethKey2, ethAddr2 := env.Chain.NewEthereumAccountWithL2Funds(baseTokensSent)
		require.EqualValues(t,
			env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr2)),
			baseTokensSent,
		)
		loop.loop(ethCallOptions{
			sender:   ethKey2,
			gasLimit: gasLimit,
		})
		// gas fee is charged regardless of result
		require.Less(t,
			env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr2)),
			baseTokensSent,
		)
	}
}

func TestLoopWithGasLeft(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	gasRatio := env.getEVMGasRatio()
	var usedGas []uint64
	for _, gasLimit := range []uint64{50000, 200000} {
		baseTokensSent := coin.Value(gas.EVMGasToISC(gasLimit, &gasRatio))
		ethKey2, _ := env.Chain.NewEthereumAccountWithL2Funds(baseTokensSent)
		res, err := iscTest.CallFn([]ethCallOptions{{
			sender:   ethKey2,
			gasLimit: gasLimit,
		}}, "loopWithGasLeft")
		require.NoError(t, err)
		require.NotEmpty(t, res.EVMReceipt.Logs)
		usedGas = append(usedGas, res.EVMReceipt.GasUsed)
	}
	require.Greater(t, usedGas[1], usedGas[0])
}

func TestEstimateGasWithoutFunds(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	callData, err := iscTest.abi.Pack("loopWithGasLeft")
	require.NoError(t, err)
	estimatedGas, err := env.evmChain.EstimateGas(ethereum.CallMsg{
		From: common.Address{},
		To:   &iscTest.address,
		Data: callData,
	}, nil)
	require.NoError(t, err)
	require.NotZero(t, estimatedGas)
	t.Log(estimatedGas)
}

func TestLoopWithGasLeftEstimateGas(t *testing.T) {
	env := InitEVM(t)
	ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	callData, err := iscTest.abi.Pack("loopWithGasLeft")
	require.NoError(t, err)
	estimatedGas, err := env.evmChain.EstimateGas(ethereum.CallMsg{
		From: ethAddr,
		To:   &iscTest.address,
		Data: callData,
	}, nil)
	require.NoError(t, err)
	require.NotZero(t, estimatedGas)
	t.Log(estimatedGas)

	gasRatio := env.getEVMGasRatio()
	baseTokensSent := coin.Value(gas.EVMGasToISC(estimatedGas, &gasRatio))
	ethKey2, _ := env.Chain.NewEthereumAccountWithL2Funds(baseTokensSent)
	res, err := iscTest.CallFn([]ethCallOptions{{
		sender:   ethKey2,
		gasLimit: estimatedGas,
	}}, "loopWithGasLeft")
	require.NoError(t, err)
	require.LessOrEqual(t, res.EVMReceipt.GasUsed, estimatedGas)
}

func TestEstimateContractGas(t *testing.T) {
	env := InitEVM(t)
	env.Chain.MustDepositBaseTokensToL2(solo.DefaultChainAdminBaseTokens, nil)

	ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()
	ethAgentID := isc.NewEthereumAddressAgentID(ethAddr)
	contract := env.deployERC20Contract(ethKey, "TEST", "tst")

	err := env.registerERC20Coin(env.Chain.ChainAdmin, coin.BaseTokenType)
	require.NoError(t, err)
	base := env.ERC20Coin(ethKey, coin.BaseTokenType)
	initialBalance := env.Chain.L2BaseTokens(ethAgentID)
	_, err = base.CallFn(nil, "transfer", contract.address, big.NewInt(int64(1*isc.Million)))
	require.NoError(t, err)
	require.EqualValues(t,
		1*isc.Million,
		env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(contract.address)),
	)
	require.LessOrEqual(t,
		env.Chain.L2BaseTokens(ethAgentID),
		initialBalance-1*isc.Million,
	)
	estimatedGas, err := env.evmChain.EstimateGas(ethereum.CallMsg{
		From: contract.address,
		To:   &ethAddr,
	}, nil)
	require.NoError(t, err)
	require.NotZero(t, estimatedGas)
}

func TestCallViewGasLimit(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	loop := env.deployLoopContract(ethKey)

	callArguments, err := loop.abi.Pack("loop")
	require.NoError(t, err)
	senderAddress := crypto.PubkeyToAddress(loop.defaultSender.PublicKey)
	callMsg := loop.callMsg(ethereum.CallMsg{
		From: senderAddress,
		Gas:  math.MaxUint64,
		Data: callArguments,
	})
	_, err = loop.chain.evmChain.CallContract(callMsg, nil)
	require.Contains(t, err.Error(), "out of gas")
}

func TestMagicContract(t *testing.T) {
	// deploy the evm contract, which starts an EVM chain and automatically
	// deploys the isc.sol EVM contract at address 0x10740000...
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

	// deploy the isc-test.sol EVM contract
	iscTest := env.deployISCTestContract(ethKey)

	// call the ISCTest.getChainId() view function of isc-test.sol which in turn:
	//  calls the ISC.getChainId() view function of isc.sol at 0x1074..., which:
	//   returns the ChainID of the underlying ISC chain
	chainID := iscTest.getChainID()

	require.True(t, env.Chain.ChainID.Equals(chainID))
}

func TestISCChainAdmin(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

	var ret struct {
		iscmagic.ISCAgentID
	}
	env.ISCMagicSandbox(ethKey).callView("getChainAdmin", nil, &ret)

	chainAdmin := env.Chain.AdminAgentID()
	require.True(t, chainAdmin.Equals(lo.Must(ret.Unwrap())))
}

func TestISCTimestamp(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

	var ret int64
	env.ISCMagicSandbox(ethKey).callView("getTimestampUnixSeconds", nil, &ret)

	require.WithinRange(t,
		time.Unix(ret, 0),
		time.Now().Add(-1*time.Hour),
		time.Now().Add(1*time.Hour),
	)
}

func TestISCCallView(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	var ret [][]byte
	chainAdmin := env.Chain.AdminAgentID()
	err := env.ISCMagicSandbox(ethKey).callView("callView", []any{
		iscmagic.WrapISCMessage(accounts.ViewBalance.Message(&chainAdmin)),
	}, &ret)
	require.NoError(t, err)
	require.NotEmpty(t, ret)
}

func TestISCTriggerEvent(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// call ISCTest.triggerEvent(string) function of isc-test.sol which in turn:
	//  calls the ISC.iscTriggerEvent(string) function of isc.sol at 0x1074..., which:
	//   triggers an ISC event with the given string parameter
	res, err := iscTest.triggerEvent("Hi from EVM!")
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, res.EVMReceipt.Status)
	events, err := env.Chain.GetEventsForBlock(env.Chain.GetLatestBlockInfo().BlockIndex)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, string(events[0].Payload), "Hi from EVM!")
}

func TestISCTriggerEventThenFail(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// test that triggerEvent() followed by revert() does not actually trigger the event
	_, err := iscTest.triggerEventFail("Hi from EVM!", ethCallOptions{
		gasLimit: 100_000, // skip estimate gas (which will fail)
	})
	require.Error(t, err)
	events, err := env.Chain.GetEventsForBlock(env.Chain.GetLatestBlockInfo().BlockIndex)
	require.NoError(t, err)
	require.Len(t, events, 0)
}

func TestISCEntropy(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// call the ISCTest.emitEntropy() function of isc-test.sol which in turn:
	//  calls ISC.iscEntropy() function of isc.sol at 0x1074..., which:
	//   returns the entropy value from the sandbox
	//  emits an EVM event (aka log) with the entropy value
	var entropy hashing.HashValue
	iscTest.CallFnExpectEvent(nil, "EntropyEvent", &entropy, "emitEntropy")

	require.NotEqualValues(t, hashing.NilHash, entropy)
}

func TestISCGetRequestID(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	var reqid isc.RequestID
	res := iscTest.CallFnExpectEvent(nil, "RequestIDEvent", &reqid, "emitRequestID")

	// check evm log is as expected
	require.NotEqualValues(t, res.EVMReceipt.Logs[0].TxHash, common.Hash{})
	require.NotEqualValues(t, res.EVMReceipt.Logs[0].BlockHash, common.Hash{})

	require.EqualValues(t, env.Chain.LastReceipt().DeserializedRequest().ID(), reqid)
}

func TestReceiptOfFailedTxDoesNotContainEvents(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	res, err := iscTest.CallFn([]ethCallOptions{{
		gasLimit: 21064,
	}}, "emitDummyEvent")
	require.Error(t, err)
	testmisc.RequireErrorToBe(t, err, "out of gas")
	require.Len(t, res.EVMReceipt.Logs, 0)
}

func TestISCGetSenderAccount(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	var sender struct {
		iscmagic.ISCAgentID
	}
	iscTest.CallFnExpectEvent(nil, "SenderAccountEvent", &sender, "emitSenderAccount")

	require.True(t, env.Chain.LastReceipt().DeserializedRequest().SenderAccount().Equals(lo.Must(sender.Unwrap())))
}

func TestSendNonPayableValueTX(t *testing.T) {
	env := InitEVM(t)

	ethKey, ethAddress := env.Chain.NewEthereumAccountWithL2Funds()

	// L2 balance of evm core contract is 0
	require.Zero(t, env.Chain.L2BaseTokens(isc.NewContractAgentID(evm.Contract.Hname())))

	// L2 balance of ISC magic contract (0x1074...) is 0
	require.Zero(t, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(iscmagic.Address)))

	// initial L2 balance of sender
	senderInitialBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddress))

	// call any function including some value
	value := util.BaseTokensDecimalsToEthereumDecimals(1*isc.Million, parameters.BaseTokenDecimals)

	sandbox := env.ISCMagicSandbox(ethKey)

	res, err := sandbox.CallFn(
		[]ethCallOptions{{sender: ethKey, value: value, gasLimit: 100_000}},
		"getSenderAccount",
	)
	require.Error(t, err, evmimpl.ErrPayingUnpayableMethod)

	// L2 balance of evm core contract is 0
	require.Zero(t, env.Chain.L2BaseTokens(isc.NewContractAgentID(evm.Contract.Hname())))
	// L2 balance of ISC magic contract (0x1074...) is 0
	require.Zero(t, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(iscmagic.Address)))
	// L2 balance of common account is: 0
	require.Zero(t, env.Chain.L2BaseTokens(isc.NewContractAgentID(0)))
	// L2 balance of sender is: initial-gasFeeCharged
	require.EqualValues(t, senderInitialBalance-res.ISCReceipt.GasFeeCharged, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddress)))
}

func TestSendPayableValueTX(t *testing.T) {
	env := InitEVM(t)

	ethKey, senderEthAddress := env.Chain.NewEthereumAccountWithL2Funds()
	_, receiver := env.solo.NewKeyPair()

	require.Zero(t, env.solo.L1BaseTokens(receiver))
	senderInitialBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(senderEthAddress))

	value := util.BaseTokensDecimalsToEthereumDecimals(1*isc.Million, parameters.BaseTokenDecimals)

	res, err := env.ISCMagicSandbox(ethKey).CallFn(
		[]ethCallOptions{{sender: ethKey, value: value, gasLimit: 100_000}},
		"transferToL1",
		receiver,
		iscmagic.WrapISCAssets(isc.NewEmptyAssets()),
	)
	require.NoError(t, err)

	valueInBaseTokens, bigRemainder := util.EthereumDecimalsToBaseTokenDecimals(
		value,
		parameters.BaseTokenDecimals,
	)
	require.Zero(t, bigRemainder.BitLen())

	// L2 balance of evm core contract is 0
	require.Zero(t, env.Chain.L2BaseTokens(isc.NewContractAgentID(evm.Contract.Hname())))
	// L2 balance of ISC magic contract (0x1074...) is 0 (!!important)
	require.Zero(t, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(iscmagic.Address)))
	// L2 balance of common account is: 0
	require.Zero(t, env.Chain.L2BaseTokens(isc.NewContractAgentID(0)))
	// L2 balance of sender is: initial - value sent in tx - gas fee
	require.EqualValues(t, senderInitialBalance-valueInBaseTokens-res.ISCReceipt.GasFeeCharged, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(senderEthAddress)))
	// L1 balance of receiver is `values sent in tx`
	require.EqualValues(t, valueInBaseTokens, env.solo.L1BaseTokens(receiver))
}

func TestSendBaseTokens(t *testing.T) {
	env := InitEVM(t)

	ethKey, ethAddress := env.Chain.EthereumAccountByIndexWithL2Funds(0)
	_, receiver := env.solo.NewKeyPair(env.solo.NewSeedFromTestNameAndTimestamp(t.Name()))

	iscTest := env.deployISCTestContract(ethKey)

	require.Zero(t, env.solo.L1BaseTokens(receiver))
	senderInitialBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddress))

	// transfer 1 mil from ethAddress L2 to receiver L1
	const transfer = 1 * isc.Million

	// attempt the operation without first calling `allow`
	_, err := iscTest.CallFn([]ethCallOptions{{
		gasLimit: 100_000, // skip estimate gas (which will fail)
	}}, "sendBaseTokens", receiver, uint64(transfer))
	require.Error(t, err)
	require.Contains(t, err.Error(), "remaining allowance insufficient")

	// allow ISCTest to take the tokens
	_, err = env.ISCMagicSandbox(ethKey).CallFn(
		[]ethCallOptions{{sender: ethKey}},
		"allow",
		iscTest.address,
		iscmagic.WrapISCAssets(isc.NewAssets(transfer)),
	)
	require.NoError(t, err)

	getAllowanceTo := func(target common.Address) *isc.Assets {
		var ret struct{ Allowance iscmagic.ISCAssets }
		env.ISCMagicSandbox(ethKey).callView("getAllowanceTo", []any{target}, &ret)
		return ret.Allowance.Unwrap()
	}

	// stored allowance should be == transfer
	require.EqualValues(t, transfer, getAllowanceTo(iscTest.address).BaseTokens())

	testdbhash.VerifyContractStateHash(env.solo, evm.Contract, "", t.Name())

	// attempt again
	const allAllowed = uint64(0)
	_, err = iscTest.CallFn(nil, "sendBaseTokens", receiver, allAllowed)
	require.NoError(t, err)
	require.GreaterOrEqual(t, env.solo.L1BaseTokens(receiver), coin.Value(transfer)-500) // 500 is the amount of tokens the contract will reserve to pay for the gas fees
	require.LessOrEqual(t, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddress)), senderInitialBalance-transfer)

	// allowance should be empty now
	require.True(t, getAllowanceTo(iscTest.address).IsEmpty())
}

func TestCannotDepleteAccount(t *testing.T) {
	env := InitEVM(t)

	ethKey, ethAddress := env.Chain.NewEthereumAccountWithL2Funds()
	_, receiver := env.solo.NewKeyPair()

	iscTest := env.deployISCTestContract(ethKey)

	require.Zero(t, env.solo.L1BaseTokens(receiver))
	senderInitialBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddress))

	// we eill attempt to transfer so much that we are left with no funds for gas
	transfer := senderInitialBalance - 300

	// allow ISCTest to take the tokens
	_, err := env.ISCMagicSandbox(ethKey).CallFn(
		[]ethCallOptions{{sender: ethKey}},
		"allow",
		iscTest.address,
		iscmagic.WrapISCAssets(isc.NewAssets(transfer)),
	)
	require.NoError(t, err)

	getAllowanceTo := func(target common.Address) *isc.Assets {
		var ret struct{ Allowance iscmagic.ISCAssets }
		env.ISCMagicSandbox(ethKey).callView("getAllowanceTo", []any{target}, &ret)
		return ret.Allowance.Unwrap()
	}

	// stored allowance should be == transfer
	require.Equal(t, transfer, getAllowanceTo(iscTest.address).BaseTokens())

	const allAllowed = uint64(0)
	_, err = iscTest.CallFn([]ethCallOptions{{
		gasLimit: 100_000, // skip estimate gas (which will fail)
	}}, "sendBaseTokens", receiver, allAllowed)
	require.ErrorContains(t, err, vm.ErrNotEnoughTokensLeftForGas.Error())
}

func TestISCCall(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	res, err := iscTest.CallFn(nil, "callInccounter")
	require.NoError(env.solo.T, err)
	require.Equal(env.solo.T, types.ReceiptStatusSuccessful, res.EVMReceipt.Status)

	r, err := env.Chain.CallView(inccounter.ViewGetCounter.Message())
	require.NoError(env.solo.T, err)
	require.EqualValues(t, 42, lo.Must(inccounter.ViewGetCounter.DecodeOutput(r)))
}

func TestFibonacciContract(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	fibo := env.deployFibonacciContract(ethKey)
	require.EqualValues(t, 2, env.getBlockNumber())

	res, err := fibo.fib(7)
	require.NoError(t, err)
	t.Log("evm gas used:", res.EVMReceipt.GasUsed)
	t.Log("isc gas used:", res.ISCReceipt.GasBurned)
	t.Log("Isc gas fee:", res.ISCReceipt.GasFeeCharged)
}

func TestEVMContractOwnsFundsL2Transfer(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// credit base tokens to the ISC test contract
	contractAgentID := isc.NewEthereumAddressAgentID(iscTest.address)
	env.Chain.GetL2FundsFromFaucet(contractAgentID)
	initialContractBalance := env.Chain.L2BaseTokens(contractAgentID)

	randAgentID := isc.NewAddressAgentID(cryptolib.NewRandomAddress())

	const nBaseTokens = 100
	allowance := isc.NewAssets(nBaseTokens)

	_, err := iscTest.CallFn(
		nil,
		"moveToAccount",
		iscmagic.WrapISCAgentID(randAgentID),
		iscmagic.WrapISCAssets(allowance),
	)
	require.NoError(t, err)

	env.Chain.AssertL2BaseTokens(randAgentID, nBaseTokens)
	env.Chain.AssertL2BaseTokens(contractAgentID, initialContractBalance-nBaseTokens)
}

func TestISCPanic(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

	iscTest := env.deployISCTestContract(ethKey)

	ret, err := iscTest.CallFn([]ethCallOptions{{
		gasLimit: 100_000, // skip estimate gas (which will fail)
	}}, "makeISCPanic")

	require.NotNil(t, ret.EVMReceipt) // evm receipt is produced

	require.Error(t, err)
	require.Equal(t, types.ReceiptStatusFailed, ret.EVMReceipt.Status)
	require.Contains(t, err.Error(), "not delegated to another chain admin")
}

func TestERC20BaseTokens(t *testing.T) {
	env := InitEVM(t)
	ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()

	env.Chain.DepositBaseTokensToL2(10*isc.Million, nil)
	err := env.registerERC20Coin(env.Chain.ChainAdmin, coin.BaseTokenType)
	require.NoError(t, err)

	erc20 := env.ERC20Coin(ethKey, coin.BaseTokenType)
	baseTokenCoinInfo := env.solo.L1CoinInfo(coin.BaseTokenType)

	{
		var name string
		require.NoError(t, erc20.callView("name", nil, &name))
		require.Equal(t, baseTokenCoinInfo.Name, name)
	}
	{
		var sym string
		require.NoError(t, erc20.callView("symbol", nil, &sym))
		require.Equal(t, baseTokenCoinInfo.Symbol, sym)
	}
	{
		var dec uint8
		require.NoError(t, erc20.callView("decimals", nil, &dec))
		require.EqualValues(t, baseTokenCoinInfo.Decimals, dec)
	}
	{
		var supply *big.Int
		require.NoError(t, erc20.callView("totalSupply", nil, &supply))
		// 4_600_000_000 is in the initial supply on IOTA
		require.Greater(t, supply.Uint64(), uint64(4_600_000_000))
	}
	{
		var balance *big.Int
		require.NoError(t, erc20.callView("balanceOf", []any{ethAddr}, &balance))
		require.EqualValues(t,
			env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr)),
			balance.Uint64(),
		)
	}
	{
		initialBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr))
		_, ethAddr2 := solo.NewEthereumAccount()
		_, err := erc20.CallFn(nil, "transfer", ethAddr2, big.NewInt(int64(1*isc.Million)))
		require.NoError(t, err)
		require.LessOrEqual(t,
			env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr)),
			initialBalance-1*isc.Million,
		)
		require.EqualValues(t,
			1*isc.Million,
			env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr2)),
		)
	}
	{
		initialBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr))
		ethKey2, ethAddr2 := env.Chain.NewEthereumAccountWithL2Funds()
		initialBalance2 := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr2))
		{
			_, err := erc20.CallFn(nil, "approve", ethAddr2, big.NewInt(int64(1*isc.Million)))
			require.NoError(t, err)
			require.Greater(t,
				env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr)),
				initialBalance-1*isc.Million,
			)
			require.EqualValues(t,
				initialBalance2,
				env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr2)),
			)
		}

		{
			var allowance *big.Int
			require.NoError(t, erc20.callView("allowance", []any{ethAddr, ethAddr2}, &allowance))
			require.EqualValues(t,
				1*isc.Million,
				allowance.Uint64(),
			)
		}
		{
			const amount = 100_000
			_, ethAddr3 := solo.NewEthereumAccount()
			_, err := erc20.CallFn([]ethCallOptions{{sender: ethKey2}}, "transferFrom", ethAddr, ethAddr3, big.NewInt(int64(amount)))
			require.NoError(t, err)
			require.Less(t,
				initialBalance-1*isc.Million,
				env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr)),
			)
			require.EqualValues(t,
				amount,
				env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr3)),
			)
			{
				var allowance *big.Int
				require.NoError(t, erc20.callView("allowance", []any{ethAddr, ethAddr2}, &allowance))
				require.EqualValues(t,
					1*isc.Million-amount,
					allowance.Uint64(),
				)
			}
		}
	}
}

func TestERC20Coin(t *testing.T) {
	t.Skip("TODO")
	testERC20Coin(nil, nil, coin.Type{}, "", "", 0, nil, nil)
	// env := InitEVM(t)
	//
	// const (
	// 	tokenName         = "ERC20 Native Token Test"
	// 	tokenTickerSymbol = "ERC20NT"
	// 	tokenDecimals     = 8
	// )
	//
	// foundryOwner, foundryOwnerAddr := env.solo.NewKeyPairWithFunds()
	// err := env.Chain.DepositBaseTokensToL2(env.solo.L1BaseTokens(foundryOwnerAddr)/2, foundryOwner)
	// require.NoError(t, err)
	//
	// supply := big.NewInt(int64(10 * isc.Million))
	// foundrySN, nativeTokenID, err := env.Chain.NewNativeTokenParams(supply).
	// 	WithUser(foundryOwner).
	// 	WithTokenName(tokenName).
	// 	WithTokenSymbol(tokenTickerSymbol).
	// 	WithTokenDecimals(tokenDecimals).
	// 	CreateFoundry()
	// require.NoError(t, err)
	// err = env.Chain.MintTokens(foundrySN, supply, foundryOwner)
	// require.NoError(t, err)
	//
	// // should not allow to register again
	// err = env.registerERC20NativeToken(foundryOwner, foundrySN, tokenName, tokenTickerSymbol, tokenDecimals)
	// require.ErrorContains(t, err, "already exists")
	//
	// ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()
	// ethAgentID := isc.NewEthereumAddressAgentID( ethAddr)
	//
	// err = env.Chain.SendFromL2ToL2Account(isc.NewAssets(0, iotago.NativeTokens{
	// 	&iotago.NativeToken{ID: nativeTokenID, Amount: supply},
	// }), ethAgentID, foundryOwner)
	// require.NoError(t, err)
	//
	// // there must be a Transfer event emitted from the ERC20NativeTokens contract
	// {
	// 	blockTxs := env.latestEVMTxs()
	// 	require.Len(t, blockTxs, 1)
	// 	tx := blockTxs[0]
	// 	receipt := env.evmChain.TransactionReceipt(tx.Hash())
	// 	require.Len(t, receipt.Logs, 1)
	// 	checkTransferEventERC20(
	// 		t,
	// 		receipt.Logs[0],
	// 		iscmagic.ERC20NativeTokensAddress(foundrySN),
	// 		common.Address{},
	// 		ethAddr,
	// 		supply,
	// 	)
	// }
	//
	// {
	// 	sandbox := env.ISCMagicSandbox(ethKey)
	// 	var addr common.Address
	// 	sandbox.callView("erc20NativeTokensAddress", []any{foundrySN}, &addr)
	// 	require.Equal(t, iscmagic.ERC20NativeTokensAddress(foundrySN), addr)
	// }
	//
	// erc20 := env.ERC20NativeTokens(ethKey, foundrySN)
	//
	// testERC20NativeTokens(
	// 	env,
	// 	erc20,
	// 	nativeTokenID,
	// 	tokenName, tokenTickerSymbol,
	// 	tokenDecimals,
	// 	supply,
	// 	ethAgentID,
	// )
}

func testERC20Coin(
	env *SoloChainEnv,
	erc20 *IscContractInstance,
	coinType coin.Type,
	tokenName, tokenTickerSymbol string,
	tokenDecimals uint8,
	supply *big.Int,
	ethAgentID isc.AgentID,
) {
	panic("TODO")
	/*
		t := env.t
		ethAddr := ethAgentID.(*isc.EthereumAddressAgentID).EthAddress()

		l2Balance := func(agentID isc.AgentID) uint64 {
			return env.Chain.L2NativeTokens(agentID, nativeTokenID).Uint64()
		}

		{
			var id struct{ iscmagic.NativeTokenID }
			require.NoError(t, erc20.callView("nativeTokenID", nil, &id))
			require.EqualValues(t, nativeTokenID[:], id.NativeTokenID.Data)
		}
		{
			var name string
			require.NoError(t, erc20.callView("name", nil, &name))
			require.Equal(t, tokenName, name)
		}
		{
			var sym string
			require.NoError(t, erc20.callView("symbol", nil, &sym))
			require.Equal(t, tokenTickerSymbol, sym)
		}
		{
			var dec uint8
			require.NoError(t, erc20.callView("decimals", nil, &dec))
			require.EqualValues(t, tokenDecimals, dec)
		}
		{
			var sup *big.Int
			require.NoError(t, erc20.callView("totalSupply", nil, &sup))
			require.Equal(t, supply.Uint64(), sup.Uint64())
		}
		{
			var balance *big.Int
			require.NoError(t, erc20.callView("balanceOf", []interface{}{ethAddr}, &balance))
			require.EqualValues(t,
				l2Balance(ethAgentID),
				balance.Uint64(),
			)
		}
		{
			initialBalance := l2Balance(ethAgentID)
			_, ethAddr2 := solo.NewEthereumAccount()
			eth2AgentID := isc.NewEthereumAddressAgentID( ethAddr2)
			_, err := erc20.CallFn(nil, "transfer", ethAddr2, big.NewInt(int64(1*isc.Million)))
			require.NoError(t, err)
			require.EqualValues(t,
				l2Balance(ethAgentID),
				initialBalance-1*isc.Million,
			)
			require.EqualValues(t,
				1*isc.Million,
				l2Balance(eth2AgentID),
			)
		}
		{
			initialBalance := l2Balance(ethAgentID)
			ethKey2, ethAddr2 := env.Chain.NewEthereumAccountWithL2Funds()
			eth2AgentID := isc.NewEthereumAddressAgentID( ethAddr2)
			initialBalance2 := l2Balance(eth2AgentID)
			{
				_, err := erc20.CallFn(nil, "approve", ethAddr2, big.NewInt(int64(1*isc.Million)))
				require.NoError(t, err)
				require.Greater(t,
					l2Balance(ethAgentID),
					initialBalance-1*isc.Million,
				)
				require.EqualValues(t,
					initialBalance2,
					l2Balance(eth2AgentID),
				)
			}

			{
				var allowance *big.Int
				require.NoError(t, erc20.callView("allowance", []interface{}{ethAddr, ethAddr2}, &allowance))
				require.EqualValues(t,
					1*isc.Million,
					allowance.Uint64(),
				)
			}
			{
				const amount = 100_000
				_, ethAddr3 := solo.NewEthereumAccount()
				eth3AgentID := isc.NewEthereumAddressAgentID( ethAddr3)
				_, err := erc20.CallFn([]ethCallOptions{{sender: ethKey2}}, "transferFrom", ethAddr, ethAddr3, big.NewInt(int64(amount)))
				require.NoError(t, err)
				require.Less(t,
					initialBalance-1*isc.Million,
					l2Balance(ethAgentID),
				)
				require.EqualValues(t,
					amount,
					l2Balance(eth3AgentID),
				)
				{
					var allowance *big.Int
					require.NoError(t, erc20.callView("allowance", []interface{}{ethAddr, ethAddr2}, &allowance))
					require.EqualValues(t,
						1*isc.Million-amount,
						allowance.Uint64(),
					)
				}
			}
		}
	*/
}

func TestERC20CoinLongName(t *testing.T) {
	t.Skip("TODO")
	/*
		env := InitEVM(t)

		var (
			tokenName         = strings.Repeat("A", 10_000)
			tokenTickerSymbol = "ERC20NT"
			tokenDecimals     = uint8(8)
		)

		foundryOwner, foundryOwnerAddr := env.solo.NewKeyPairWithFunds()
		err := env.Chain.DepositBaseTokensToL2(env.solo.L1BaseTokens(foundryOwnerAddr)/2, foundryOwner)
		require.NoError(t, err)

		supply := big.NewInt(int64(10 * isc.Million))

		foundrySN, _, err := env.Chain.NewNativeTokenParams(supply).
			WithUser(foundryOwner).
			WithTokenName(tokenName).
			WithTokenSymbol(tokenTickerSymbol).
			WithTokenDecimals(tokenDecimals).
			CreateFoundry()
		require.ErrorContains(t, err, "too long")
		require.Zero(t, foundrySN)
	*/
}

// test withdrawing ALL EVM balance to a L1 address via the magic contract
func TestEVMWithdrawAll(t *testing.T) {
	env := InitEVM(t)
	ethKey, ethAddress := env.Chain.NewEthereumAccountWithL2Funds()
	_, receiver := env.solo.NewKeyPair()

	tokensOnL2 := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddress))

	// try withdrawing all base tokens
	_, err := env.ISCMagicSandbox(ethKey).CallFn(
		[]ethCallOptions{{
			sender:   ethKey,
			gasLimit: 100_000, // provide a gas limit value as the estimation will fail
		}},
		"transferToL1",
		receiver,
		iscmagic.WrapISCAssets(isc.NewAssets(tokensOnL2)),
	)
	// request must fail with an error, and receiver should not receive any funds
	require.Error(t, err)
	require.Regexp(t, vm.ErrNotEnoughTokensLeftForGas.Error(), err.Error())
	iscReceipt := env.Chain.LastReceipt()
	require.Error(t, iscReceipt.Error.AsGoError())
	require.EqualValues(t, 0, env.solo.L1BaseTokens(receiver))

	// refresh since the last request consumed some gas
	tokensOnL2 = env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddress))
	transfer1 := tokensOnL2 / 2

	// retry the request above, but now leave some tokens to pay for the gas fees
	_, err = env.ISCMagicSandbox(ethKey).CallFn(
		[]ethCallOptions{{sender: ethKey}},
		"transferToL1",
		receiver,
		iscmagic.WrapISCAssets(isc.NewAssets(transfer1)),
	)
	require.NoError(t, err)

	iscReceipt = env.Chain.LastReceipt()
	require.NoError(t, iscReceipt.Error.AsGoError())
	gasFee := iscReceipt.GasFeeCharged
	require.EqualValues(t, tokensOnL2-transfer1-gasFee, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddress)))
	require.EqualValues(t, transfer1, env.solo.L1BaseTokens(receiver))

	// transfer all
	t.Skip(`TODO: this causes {failure MoveAbort(MoveLocation { module: ModuleId { address: 0000000000000000000000000000000000000000000000000000000000000002, name: Identifier("balance") }, function: 7, instruction: 10, function_name: Some("split") }, 2) in command 1}`)
	tokensOnL2 = env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddress))
	transfer2 := tokensOnL2 - gasFee
	_, err = env.ISCMagicSandbox(ethKey).CallFn(
		[]ethCallOptions{{sender: ethKey}},
		"transferToL1",
		receiver,
		iscmagic.WrapISCAssets(isc.NewAssets(transfer2)),
	)
	require.NoError(t, err)
	iscReceipt = env.Chain.LastReceipt()
	require.NoError(t, iscReceipt.Error.AsGoError())
	require.Zero(t, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddress)))
	require.EqualValues(t, transfer1+transfer2, env.solo.L1BaseTokens(receiver))
}

func TestEVMGasPriceMismatch(t *testing.T) {
	for _, v := range []struct {
		name          string
		gasPerToken   util.Ratio32
		evmGasRatio   util.Ratio32
		txGasPrice    *big.Int
		expectedError string
		gasBurned     uint64
		feeCharged    uint64
	}{
		{
			name:        "fees disabled, gas price nil",
			gasPerToken: util.Ratio32{A: 0, B: 0},
			evmGasRatio: util.Ratio32{A: 1, B: 1},
			txGasPrice:  nil,
			gasBurned:   gas.BurnCodeMinimumGasPerRequest1P.Cost(),
			feeCharged:  0,
		},
		{
			name:        "fees disabled, gas price 1",
			gasPerToken: util.Ratio32{A: 0, B: 0},
			evmGasRatio: util.Ratio32{A: 1, B: 1},
			txGasPrice:  big.NewInt(1),
			gasBurned:   gas.BurnCodeMinimumGasPerRequest1P.Cost(),
			feeCharged:  0,
		},
		{
			name:          "default policy, gas price nil",
			gasPerToken:   util.Ratio32{A: 100, B: 1}, // default: 1 base token = 100 gas units
			evmGasRatio:   util.Ratio32{A: 1, B: 1},   // default
			txGasPrice:    nil,
			expectedError: "insufficient gas price: got 0, minimum is 10000000",
			gasBurned:     168107,
			feeCharged:    1682,
		},
		{
			name:          "default policy, gas price too low",
			gasPerToken:   util.Ratio32{A: 100, B: 1}, // default: 1 base token = 100 gas units
			evmGasRatio:   util.Ratio32{A: 1, B: 1},   // default
			txGasPrice:    big.NewInt(9999999),
			expectedError: "insufficient gas price: got 9999999, minimum is 10000000",
			gasBurned:     168107,
			feeCharged:    1682,
		},
		{
			name:        "default policy, gas price just enough",
			gasPerToken: util.Ratio32{A: 100, B: 1}, // default: 1 base token = 100 gas units
			evmGasRatio: util.Ratio32{A: 1, B: 1},   // default
			txGasPrice:  big.NewInt(10000000),
			gasBurned:   25892,
			feeCharged:  259,
		},
		{
			name:        "default policy, gas price 2x",
			gasPerToken: util.Ratio32{A: 100, B: 1}, // default: 1 base token = 100 gas units
			evmGasRatio: util.Ratio32{A: 1, B: 1},   // default
			txGasPrice:  big.NewInt(2 * 10000000),
			gasBurned:   25892,
			feeCharged:  2 * 259,
		},
		{
			name:        "default policy, gas price 2x, evmGasRatio cheaper",
			gasPerToken: util.Ratio32{A: 100, B: 1}, // default: 1 base token = 100 gas units
			evmGasRatio: util.Ratio32{A: 1, B: 2},
			txGasPrice:  big.NewInt(2 * 10000000),
			gasBurned:   12951,
			feeCharged:  2*259/2 + 1,
		},
		{
			name:          "gas more expensive, gas price too low",
			gasPerToken:   util.Ratio32{A: 50, B: 1}, // 1 base token = 50 gas units
			evmGasRatio:   util.Ratio32{A: 1, B: 1},  // default
			txGasPrice:    big.NewInt(19999999),
			expectedError: "insufficient gas price: got 19999999, minimum is 20000000",
			gasBurned:     168107,
			feeCharged:    2*1682 - 1,
		},
		{
			name:        "gas more expensive, gas price just enough",
			gasPerToken: util.Ratio32{A: 50, B: 1}, // 1 base token = 50 gas units
			evmGasRatio: util.Ratio32{A: 1, B: 1},  // default
			txGasPrice:  big.NewInt(2 * 10000000),
			gasBurned:   25892,
			feeCharged:  2 * 259,
		},
		{
			name:        "gas more expensive, gas price 2x",
			gasPerToken: util.Ratio32{A: 50, B: 1}, // 1 base token = 50 gas units
			evmGasRatio: util.Ratio32{A: 1, B: 1},  // default
			txGasPrice:  big.NewInt(2 * 2 * 10000000),
			gasBurned:   25892,
			feeCharged:  2 * 2 * 259,
		},
	} {
		t.Run(v.name, func(t *testing.T) {
			env := InitEVM(t)
			env.Chain.DepositBaseTokensToL2(10*isc.Million, nil)
			feePolicy := env.Chain.GetGasFeePolicy()
			feePolicy.GasPerToken = v.gasPerToken
			feePolicy.EVMGasRatio = v.evmGasRatio
			err := env.setFeePolicy(*feePolicy)
			require.NoError(t, err)

			ethKey, senderAddress := env.Chain.NewEthereumAccountWithL2Funds()

			// deploy solidity `storage` contract
			storage := env.deployStorageContract(ethKey)

			// issue a tx with an arbitrary gas price
			valueToStore := uint32(888)
			callArguments, err := storage.abi.Pack("store", valueToStore)
			require.NoError(t, err)
			nonce := storage.chain.getNonce(senderAddress)
			unsignedTx := types.NewTransaction(
				nonce,
				storage.address,
				util.Big0,
				env.maxGasLimit(),
				v.txGasPrice,
				callArguments,
			)

			tx, err := types.SignTx(unsignedTx, evmutil.Signer(big.NewInt(int64(storage.chain.evmChain.ChainID()))), ethKey)
			require.NoError(t, err)

			err = storage.chain.evmChain.SendTransaction(tx)
			if v.expectedError != "" {
				require.Error(t, err)
				require.Equal(t, v.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}

			iscReceipt := env.Chain.LastReceipt()
			require.EqualValuesf(t, v.gasBurned, iscReceipt.GasBurned, "gas burned: %d", iscReceipt.GasBurned)
			require.EqualValuesf(t, v.feeCharged, iscReceipt.GasFeeCharged, "gas fee charged: %d", iscReceipt.GasFeeCharged)
		})
	}
}

func TestEVMIntrinsicGas(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	sandbox := env.ISCMagicSandbox(ethKey)
	res, err := sandbox.CallFn([]ethCallOptions{{
		sender:   ethKey,
		gasLimit: 1,
	}}, "getEntropy")
	require.ErrorContains(t, err, "intrinsic gas")
	require.NotZero(t, res.ISCReceipt.GasFeeCharged)
}

func TestEVMTransferBaseTokens(t *testing.T) {
	env := InitEVM(t)
	ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()
	_, someEthereumAddr := solo.NewEthereumAccount()
	someAgentID := isc.NewEthereumAddressAgentID(someEthereumAddr)

	sendTx := func(amount *big.Int) {
		nonce := env.getNonce(ethAddr)
		unsignedTx := types.NewTransaction(nonce, someEthereumAddr, amount, env.maxGasLimit(), env.evmChain.GasPrice(), []byte{})
		tx, err := types.SignTx(unsignedTx, evmutil.Signer(big.NewInt(int64(env.evmChainID))), ethKey)
		require.NoError(t, err)
		err = env.evmChain.SendTransaction(tx)
		require.NoError(t, err)
	}

	// try to transfer base tokens between 2 ethereum addresses

	// issue a tx with non-0 amount (try to send ETH/basetoken)
	// try sending 1 million base tokens (expressed in ethereum decimals)
	value := util.BaseTokensDecimalsToEthereumDecimals(
		1*isc.Million,
		parameters.BaseTokenDecimals,
	)
	sendTx(value)
	env.Chain.AssertL2BaseTokens(someAgentID, 1*isc.Million)
}

func TestSolidityTransferBaseTokens(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	_, someEthereumAddr := solo.NewEthereumAccount()
	someEthereumAgentID := isc.NewEthereumAddressAgentID(someEthereumAddr)

	iscTest := env.deployISCTestContract(ethKey)

	// try sending funds to `someEthereumAddr` by sending a "value tx" to the isc test contract
	oneMillionInEthDecimals := util.BaseTokensDecimalsToEthereumDecimals(
		1*isc.Million,
		parameters.BaseTokenDecimals,
	)

	_, err := iscTest.CallFn([]ethCallOptions{{
		sender: ethKey,
		value:  oneMillionInEthDecimals,
	}}, "sendTo", someEthereumAddr, oneMillionInEthDecimals)
	require.NoError(t, err)
	env.Chain.AssertL2BaseTokens(someEthereumAgentID, 1*isc.Million)

	// attempt to send more than the contract will have available
	twoMillionInEthDecimals := util.BaseTokensDecimalsToEthereumDecimals(
		2*isc.Million,
		parameters.BaseTokenDecimals,
	)

	_, err = iscTest.CallFn([]ethCallOptions{{
		sender: ethKey,
		value:  oneMillionInEthDecimals,
	}}, "sendTo", someEthereumAddr, twoMillionInEthDecimals)
	require.Error(t, err)
	env.Chain.AssertL2BaseTokens(someEthereumAgentID, 1*isc.Million)

	// fund the contract via a L1 wallet ISC transfer, then call `sendTo` to use those funds
	l1Wallet, _ := env.Chain.Env.NewKeyPairWithFunds()
	env.Chain.TransferAllowanceTo(
		isc.NewAssets(10*isc.Million),
		isc.NewEthereumAddressAgentID(iscTest.address),
		l1Wallet,
	)

	tenMillionInEthDecimals := util.BaseTokensDecimalsToEthereumDecimals(
		10*isc.Million,
		parameters.BaseTokenDecimals,
	)

	_, err = iscTest.CallFn([]ethCallOptions{{
		sender: ethKey,
	}}, "sendTo", someEthereumAddr, tenMillionInEthDecimals)
	require.NoError(t, err)
	env.Chain.AssertL2BaseTokens(someEthereumAgentID, 11*isc.Million)

	// send more than the balance
	_, err = iscTest.CallFn([]ethCallOptions{{
		sender:   ethKey,
		value:    tenMillionInEthDecimals.Mul(tenMillionInEthDecimals, big.NewInt(1000000)),
		gasLimit: 100_000, // provide a gas limit value as the estimation will fail
	}}, "sendTo", someEthereumAddr, big.NewInt(0))
	require.Error(t, err)
	env.Chain.AssertL2BaseTokens(someEthereumAgentID, 11*isc.Million)
}

func TestSendEntireBalance(t *testing.T) {
	env := InitEVM(t)
	ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()
	_, someEthereumAddr := solo.NewEthereumAccount()
	someEthereumAgentID := isc.NewEthereumAddressAgentID(someEthereumAddr)

	// send all initial
	initial := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr))
	// try sending funds to `someEthereumAddr` by sending a "value tx"
	initialBalanceInEthDecimals := util.BaseTokensDecimalsToEthereumDecimals(
		initial,
		parameters.BaseTokenDecimals,
	)

	unsignedTx := types.NewTransaction(0, someEthereumAddr, initialBalanceInEthDecimals, env.maxGasLimit(), env.evmChain.GasPrice(), []byte{})
	tx, err := types.SignTx(unsignedTx, evmutil.Signer(big.NewInt(int64(env.evmChainID))), ethKey)
	require.NoError(t, err)
	err = env.evmChain.SendTransaction(tx)
	// this will produce an error because there won't be tokens left in the account to pay for gas
	require.ErrorContains(t, err, vm.ErrNotEnoughTokensLeftForGas.Error())
	evmReceipt := env.evmChain.TransactionReceipt(tx.Hash())
	require.Equal(t, evmReceipt.Status, types.ReceiptStatusFailed)
	rec := env.Chain.LastReceipt()
	require.EqualValues(t, vm.ErrNotEnoughTokensLeftForGas.Error(), rec.ResolvedError)
	env.Chain.AssertL2BaseTokens(someEthereumAgentID, 0)

	// now try sending all balance, minus the funds needed for gas
	currentBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr))

	currentBalanceInEthDecimals := util.BaseTokensDecimalsToEthereumDecimals(
		currentBalance,
		parameters.BaseTokenDecimals,
	)

	estimatedGas, err := env.evmChain.EstimateGas(ethereum.CallMsg{
		From:  ethAddr,
		To:    &someEthereumAddr,
		Value: currentBalanceInEthDecimals,
		Data:  []byte{},
	}, nil)
	require.NoError(t, err)

	feePolicy := env.Chain.GetGasFeePolicy()
	gasPrice := feePolicy.DefaultGasPriceFullDecimals(parameters.BaseTokenDecimals)
	tokensForGasBudget := feePolicy.FeeFromGas(estimatedGas, gasPrice, parameters.BaseTokenDecimals)

	gasLimit := feePolicy.GasBudgetFromTokensWithGasPrice(tokensForGasBudget, gasPrice, parameters.BaseTokenDecimals)

	valueToSendInEthDecimals := util.BaseTokensDecimalsToEthereumDecimals(
		currentBalance-tokensForGasBudget,
		parameters.BaseTokenDecimals,
	)
	unsignedTx = types.NewTransaction(1, someEthereumAddr, valueToSendInEthDecimals, gasLimit, env.evmChain.GasPrice(), []byte{})
	tx, err = types.SignTx(unsignedTx, evmutil.Signer(big.NewInt(int64(env.evmChainID))), ethKey)
	require.NoError(t, err)
	err = env.evmChain.SendTransaction(tx)
	require.NoError(t, err)
	env.Chain.AssertL2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr), 0)
	env.Chain.AssertL2BaseTokens(someEthereumAgentID, currentBalance-tokensForGasBudget)
}

func TestSolidityRevertMessage(t *testing.T) {
	env := InitEVM(t)
	ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// test the revert reason is shown when invoking eth_call
	callData, err := iscTest.abi.Pack("testRevertReason")
	require.NoError(t, err)
	_, err = env.evmChain.CallContract(ethereum.CallMsg{
		From: ethAddr,
		To:   &iscTest.address,
		Gas:  100_000,
		Data: callData,
	}, nil)
	require.ErrorContains(t, err, "execution reverted")

	revertData, err := evmerrors.ExtractRevertData(err)
	require.NoError(t, err)
	revertString, err := abi.UnpackRevert(revertData)
	require.NoError(t, err)

	require.Equal(t, "foobar", revertString)

	res, err := iscTest.CallFn([]ethCallOptions{{
		gasLimit: 100_000, // needed because gas estimation would fail
	}}, "testRevertReason")
	require.Error(t, err)
	require.Regexp(t, `execution reverted: \w+`, res.ISCReceipt.ResolvedError)
}

func TestCallContractCannotCauseStackOverflow(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

	iscTest := env.deployISCTestContract(ethKey)

	// tx contract call
	ret, err := iscTest.CallFn([]ethCallOptions{{
		gasLimit: 100_000, // skip estimate gas (which will fail)
	}}, "testStackOverflow")

	require.ErrorContains(t, err, "unauthorized access")
	require.NotNil(t, ret.EVMReceipt) // evm receipt is produced

	// view call
	err = iscTest.callView("testStackOverflow", nil, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "unauthorized access")
}

func TestStaticCall(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	res, err := iscTest.CallFn([]ethCallOptions{{
		sender: ethKey,
	}}, "testStaticCall")
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, res.EVMReceipt.Status)
	events, err := env.Chain.GetEventsForBlock(env.Chain.GetLatestBlockInfo().BlockIndex)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, string(events[0].Payload), "non-static")
}

func TestSelfDestruct(t *testing.T) {
	// NOTE: since EIP-6780 self-destruct was deprecated
	env := InitEVM(t)
	ethKey, _ := env.Chain.EthereumAccountByIndexWithL2Funds(0)

	iscTest := env.deployISCTestContract(ethKey)
	iscTestAgentID := isc.NewEthereumAddressAgentID(iscTest.address)

	// send some tokens to the ISCTest contract
	{
		const baseTokensDepositFee = 500
		k, _ := env.solo.NewKeyPairWithFunds(env.solo.NewSeedFromTestNameAndTimestamp(t.Name()))
		err := env.Chain.SendFromL1ToL2AccountBaseTokens(baseTokensDepositFee, 1*isc.Million, iscTestAgentID, k)
		require.NoError(t, err)
		require.EqualValues(t, 1*isc.Million, env.Chain.L2BaseTokens(iscTestAgentID))
	}

	_, beneficiary := solo.EthereumAccountByIndex(1)

	require.NotEmpty(t, env.getCode(iscTest.address))

	_, err := iscTest.CallFn([]ethCallOptions{{sender: ethKey}}, "testSelfDestruct", beneficiary)
	require.NoError(t, err)

	// (EIP-6780) SELFDESTRUCT will recover all funds to the target but not delete the account,
	// except when called in the same transaction as creation
	require.NotEmpty(t, env.getCode(iscTest.address))
	require.Zero(t, env.Chain.L2BaseTokens(iscTestAgentID))
	require.EqualValues(t, 1*isc.Million, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(beneficiary)))

	testdbhash.VerifyContractStateHash(env.solo, evm.Contract, "", t.Name())
}

func TestSelfDestruct6780(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.EthereumAccountByIndexWithL2Funds(0)
	iscTest := env.deployISCTestContract(ethKey)

	var createContractAddr common.Address
	iscTest.CallFnExpectEvent(nil, "TestSelfDestruct6780ContractCreated", &createContractAddr, "testSelfDestruct6780")
	require.Empty(t, env.getCode(createContractAddr))
}

func TestChangeGasLimit(t *testing.T) {
	env := InitEVM(t)
	env.Chain.MustDepositBaseTokensToL2(solo.DefaultChainAdminBaseTokens, nil)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	storage := env.deployStorageContract(ethKey)

	var blockHashes []common.Hash
	for i := range 10 {
		res, err := storage.store(uint32(i))
		blockHashes = append(blockHashes, res.EVMReceipt.BlockHash)
		require.NoError(t, err)
	}

	{
		feePolicy := env.Chain.GetGasFeePolicy()
		feePolicy.EVMGasRatio.B *= 2
		err := env.setFeePolicy(*feePolicy)
		require.NoError(t, err)
	}

	for _, h := range blockHashes {
		b := env.evmChain.BlockByHash(h)
		require.Equal(t, b.Hash(), h)
	}
}

func TestChangeGasPerToken(t *testing.T) {
	env := InitEVM(t)
	env.Chain.MustDepositBaseTokensToL2(solo.DefaultChainAdminBaseTokens, nil)

	var fee coin.Value
	{
		ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
		storage := env.deployStorageContract(ethKey)
		res, err := storage.store(uint32(3))
		require.NoError(t, err)
		fee = res.ISCReceipt.GasFeeCharged
	}

	{
		feePolicy := env.Chain.GetGasFeePolicy()
		feePolicy.GasPerToken.B *= 2
		err := env.setFeePolicy(*feePolicy)
		require.NoError(t, err)
	}

	var fee2 coin.Value
	{
		ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
		storage := env.deployStorageContract(ethKey)
		res, err := storage.store(uint32(3))
		require.NoError(t, err)
		fee2 = res.ISCReceipt.GasFeeCharged
	}

	t.Log(fee, fee2)
	require.Greater(t, fee2, fee)
}

func TestGasPriceIgnoredInEstimateGas(t *testing.T) {
	env := InitEVM(t)

	var gasLimit []uint64

	for _, gasPrice := range []*big.Int{
		nil,
		big.NewInt(0),
		big.NewInt(10),
		big.NewInt(100),
	} {
		t.Run(fmt.Sprintf("%v", gasPrice), func(t *testing.T) { //nolint:gocritic // false positive
			ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
			storage := env.deployStorageContract(ethKey)

			gas, err := storage.estimateGas([]ethCallOptions{{
				sender:   ethKey,
				gasPrice: gasPrice,
			}}, "store", uint32(3))
			require.NoError(t, err)

			gasLimit = append(gasLimit, gas)
		})
	}

	t.Log("gas limit", gasLimit)
	require.Len(t, lo.Uniq(gasLimit), 1)
}

// calling views via eth_call must not cost gas (still has a maximum budget, but simple view calls should pass)
func TestEVMCallViewGas(t *testing.T) {
	env := InitEVM(t)

	// issue a view call from an account with no funds
	ethKey, _ := solo.NewEthereumAccount()

	var ret struct {
		iscmagic.ISCAgentID
	}
	err := env.ISCMagicSandbox(ethKey).callView("getChainAdmin", nil, &ret)
	require.NoError(t, err)
}

func TestGasPrice(t *testing.T) {
	env := InitEVM(t)
	env.Chain.DepositBaseTokensToL2(10*isc.Million, nil)

	price1 := env.evmChain.GasPrice().Uint64()
	require.NotZero(t, price1)

	{
		feePolicy := env.Chain.GetGasFeePolicy()
		feePolicy.GasPerToken.B *= 2 // 1 gas is paid with 2 tokens
		err := env.setFeePolicy(*feePolicy)
		require.NoError(t, err)
	}

	price2 := env.evmChain.GasPrice().Uint64()
	require.EqualValues(t, price1*2, price2)

	t.Skip(`TODO: this causes {failure MoveAbort(MoveLocation { module: ModuleId { address: 0000000000000000000000000000000000000000000000000000000000000002, name: Identifier("balance") }, function: 7, instruction: 10, function_name: Some("split") }, 2) in command 4}`)
	{
		feePolicy := env.Chain.GetGasFeePolicy()
		feePolicy.EVMGasRatio.A *= 2 // 1 EVM gas unit consumes 2 ISC gas units
		err := env.setFeePolicy(*feePolicy)
		require.NoError(t, err)
	}

	price3 := env.evmChain.GasPrice().Uint64()
	require.EqualValues(t, price2*2, price3)

	{
		feePolicy := env.Chain.GetGasFeePolicy()
		feePolicy.GasPerToken.A = 0
		feePolicy.GasPerToken.B = 0
		err := env.setFeePolicy(*feePolicy)
		require.NoError(t, err)
	}

	price4 := env.evmChain.GasPrice().Uint64()
	require.EqualValues(t, 0, price4)
}

func TestTraceTransaction(t *testing.T) {
	t.Skip("TODO: failed to get anchor content: version not found")
	env := InitEVM(t)
	ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()

	traceLatestTx := func() *jsonrpc.CallFrame {
		latestBlock, err := env.evmChain.BlockByNumber(nil)
		require.NoError(t, err)
		trace, err := env.evmChain.TraceTransaction(latestBlock.Transactions()[0].Hash(), &tracers.TraceConfig{})
		require.NoError(t, err)
		var ret jsonrpc.CallFrame
		err = json.Unmarshal(trace.(json.RawMessage), &ret)
		require.NoError(t, err)
		t.Log(ret)
		return &ret
	}
	{
		storage := env.deployStorageContract(ethKey)
		_, err := storage.store(43)
		require.NoError(t, err)
		trace := traceLatestTx()
		require.EqualValues(t, ethAddr, trace.From)
		require.EqualValues(t, storage.address, *trace.To)
		require.Empty(t, trace.Calls)
	}
	{
		iscTest := env.deployISCTestContract(ethKey)
		_, err := iscTest.triggerEvent("Hi from EVM!")
		require.NoError(t, err)
		trace := traceLatestTx()
		require.EqualValues(t, ethAddr, trace.From)
		require.EqualValues(t, iscTest.address, *trace.To)
		require.NotEmpty(t, trace.Calls)
	}
}

func TestMagicContractExamples(t *testing.T) {
	t.Skip("TODO")
	/*
		env := InitEVM(t)
		ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

		contract := env.deployERC20ExampleContract(ethKey)

		contractAgentID := isc.NewEthereumAddressAgentID( contract.address)
		env.Chain.GetL2FundsFromFaucet(contractAgentID)

		_, err := contract.CallFn(nil, "createFoundry", big.NewInt(1000000), uint64(10_000))
		require.NoError(t, err)

		_, err = contract.CallFn(nil, "registerToken", "TESTCOIN", "TEST", uint8(18), uint64(10_000))
		require.NoError(t, err)

		_, err = contract.CallFn(nil, "mint", big.NewInt(1000), uint64(10_000))
		require.NoError(t, err)

		ethKey2, _ := env.Chain.NewEthereumAccountWithL2Funds()
		isTestContract := env.deployISCTestContract(ethKey2)
		iscTestAgentID := isc.NewEthereumAddressAgentID( isTestContract.address)
		env.Chain.GetL2FundsFromFaucet(iscTestAgentID)

		_, err = isTestContract.CallFn(nil, "mint", uint32(1), big.NewInt(1000), uint64(10_000))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unauthorized")
	*/
}

func TestMagicContractExamplesWithNativeToken(t *testing.T) {
	t.Skip("TODO")
	/*
		env := InitEVM(t)
		ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

		contract := env.deployERC20ExampleContract(ethKey)

		contractAgentID := isc.NewEthereumAddressAgentID( contract.address)
		env.Chain.GetL2FundsFromFaucet(contractAgentID)

		_, err := contract.CallFn(nil, "createNativeTokenFoundry", "TESTCOIN", "TEST", uint8(18), big.NewInt(1000000), uint64(10_000))
		require.NoError(t, err)

		_, err = contract.CallFn(nil, "mint", big.NewInt(1000), uint64(10_000))
		require.NoError(t, err)

		ethKey2, _ := env.Chain.NewEthereumAccountWithL2Funds()
		isTestContract := env.deployISCTestContract(ethKey2)
		iscTestAgentID := isc.NewEthereumAddressAgentID( isTestContract.address)
		env.Chain.GetL2FundsFromFaucet(iscTestAgentID)

		_, err = isTestContract.CallFn(nil, "mint", uint32(1), big.NewInt(1000), uint64(10_000))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unauthorized")
	*/
}

func TestCaller(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)
	err := env.Chain.TransferAllowanceTo(
		isc.NewAssets(42),
		isc.NewEthereumAddressAgentID(iscTest.address),
		env.Chain.ChainAdmin,
	)
	require.NoError(t, err)

	var r []byte
	err = iscTest.callView("testCallViewCaller", nil, &r)
	require.NoError(t, err)
	require.EqualValues(t, 42, lo.Must(isc.CoinBalancesFromBytes(r)).BaseTokens())
}

func TestCustomError(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)
	_, err := iscTest.CallFn([]ethCallOptions{{
		gasLimit: 100000,
	}}, "revertWithCustomError")
	require.ErrorContains(t, err, "execution reverted")

	revertData, err := evmerrors.ExtractRevertData(err)
	require.NoError(t, err)

	args, err := evmerrors.UnpackCustomError(revertData, iscTest.abi.Errors["CustomError"])
	require.NoError(t, err)

	require.Len(t, args, 1)
	require.EqualValues(t, 42, args[0])
}

func TestEmitEventAndRevert(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)
	res, err := iscTest.CallFn([]ethCallOptions{{
		gasLimit: 100000,
	}}, "emitEventAndRevert")
	require.ErrorContains(t, err, "execution reverted")
	require.Empty(t, res.EVMReceipt.Logs)
}

func TestL1DepositEVM(t *testing.T) {
	env := InitEVM(t)
	// ensure that after a deposit to an EVM account, there is a tx/receipt for it to be auditable on the EVM side
	wallet, l1Addr := env.solo.NewKeyPairWithFunds()
	_, ethAddr := solo.NewEthereumAccount()
	const amount = 1 * isc.Million
	err := env.Chain.TransferAllowanceTo(
		isc.NewAssets(amount),
		isc.NewEthereumAddressAgentID(ethAddr),
		wallet,
	)
	require.NoError(t, err)

	bal, err := env.Chain.EVM().Balance(ethAddr, nil)
	require.NoError(t, err)

	// previous block must only have 1 tx, that corresponds to the deposit to ethAddr
	blockTxs := env.latestEVMTxs()
	require.Len(t, blockTxs, 1)
	tx := blockTxs[0]
	require.True(t, tx.GasPrice().Cmp(util.Big0) == 1)
	require.True(t, ethAddr == *tx.To())
	require.Zero(t, tx.Value().Cmp(bal))

	// assert txData has the expected information (<agentID sender> + assets)
	buf := (bytes.NewReader(tx.Data()))
	rr := rwutil.NewReader(buf)
	a := bcs.MustUnmarshalStream[isc.AgentID](buf)
	require.True(t, a.Equals(isc.NewAddressAgentID(l1Addr)))
	assets := bcs.MustUnmarshalStream[isc.Assets](buf)

	// blockIndex
	blockIndex := rr.ReadUint32()
	require.Equal(t, env.evmChain.BlockNumber().Uint64(), uint64(blockIndex))
	reqIndex := rr.ReadUint16()
	require.Zero(t, reqIndex)
	n, err := buf.Read([]byte{})
	require.Zero(t, n)
	require.ErrorIs(t, err, io.EOF)
	require.NoError(t, rr.Err)

	require.EqualValues(t,
		util.MustEthereumDecimalsToBaseTokenDecimalsExact(bal, parameters.BaseTokenDecimals),
		assets.BaseTokens())

	evmRec := env.Chain.EVM().TransactionReceipt(tx.Hash())
	require.NotNil(t, evmRec)
	require.Equal(t, types.ReceiptStatusSuccessful, evmRec.Status)
	iscRec := env.Chain.LastReceipt()
	feePolicy := env.Chain.GetGasFeePolicy()
	expectedGas := gas.ISCGasBudgetToEVM(iscRec.GasBurned, &feePolicy.EVMGasRatio)
	require.EqualValues(t, expectedGas, evmRec.GasUsed)

	// issue the same deposit again, assert txHashes do not collide

	err = env.Chain.TransferAllowanceTo(
		isc.NewAssets(amount),
		isc.NewEthereumAddressAgentID(ethAddr),
		wallet,
	)
	require.NoError(t, err)

	blockTxs2 := env.latestEVMTxs()
	require.Len(t, blockTxs2, 1)
	tx2 := blockTxs2[0]
	require.NotEqual(t, tx.Hash(), tx2.Hash())
}

func TestDecimalsConversion(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// call any function including 999999999 wei as value (which is just 1 wei short of 1 base token)
	lessThanOneGlow := new(big.Int).SetUint64(999999999)
	valueInBaseTokens, remainder := util.EthereumDecimalsToBaseTokenDecimals(
		lessThanOneGlow,
		parameters.BaseTokenDecimals,
	)
	t.Log(valueInBaseTokens)
	require.Zero(t, valueInBaseTokens)
	require.EqualValues(t, lessThanOneGlow.Uint64(), remainder.Uint64())

	_, err := iscTest.CallFn(
		[]ethCallOptions{{sender: ethKey, value: lessThanOneGlow, gasLimit: 100000}},
		"sendTo",
		iscTest.address,
		big.NewInt(0),
	)
	require.ErrorContains(t, err, "execution reverted")
}

func TestPreEIP155Transaction(t *testing.T) {
	env := InitEVM(t)
	ethKey, _ := env.Chain.EthereumAccountByIndexWithL2Funds(0)

	// use a signer without replay protection
	signer := types.HomesteadSigner{}

	tx, err := types.SignTx(
		types.NewContractCreation(0, big.NewInt(1_000_000_000), 1_000_000_000, env.evmChain.GasPrice(), nil),
		signer,
		ethKey,
	)
	require.NoError(t, err)

	err = env.evmChain.SendTransaction(tx)
	require.NoError(t, err)
}

func TestEVMEventOnFailedL1Deposit(t *testing.T) {
	t.Skip("!!! TODO")
	// env := InitEVM(t)
	// _, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()
	//
	// // set gas policy to a higher price (so that it can fails when charging ISC gas)
	// {
	// 	feePolicy := env.Chain.GetGasFeePolicy()
	// 	feePolicy.GasPerToken.A = 1
	// 	feePolicy.GasPerToken.B = 10
	// 	err := env.setFeePolicy(*feePolicy)
	// 	require.NoError(t, err)
	// }
	// // mint an NFT and send it to the chain
	// issuerWallet, issuerAddress := env.solo.NewKeyPairWithFunds()
	// metadata := []byte("foobar")
	// nft, _, err := env.solo.MintNFTL1(issuerWallet, issuerAddress, metadata)
	// require.NoError(t, err)
	// ethAgentID := isc.NewEthereumAddressAgentID( ethAddr)
	//
	// callParams := solo.NewCallParams(accounts.Contract.Name, accounts.FuncTransferAllowanceTo.Name, accounts.ParamAgentID, codec.Encode(ethAgentID)).
	// 	AddBaseTokens(1_000_000).
	// 	WithNFT(nft).
	// 	WithAllowance(isc.NewEmptyAssets().AddNFTs(nft.ID)).
	// 	WithMaxAffordableGasBudget()
	//
	// // do not include enough gas budget (but just enough to execute until the end)
	// _, estimatedReceipt, err := env.Chain.EstimateGasOnLedger(callParams, issuerWallet)
	// require.NoError(t, err)
	// callParams.WithGasBudget(estimatedReceipt.GasBurned - 1)
	//
	// _, err = env.Chain.PostRequestSync(callParams, issuerWallet)
	// require.Error(t, err)
	// require.Contains(t, err.Error(), "gas budget exceeded")
	//
	// // assert NO event is issued
	// logs := env.LastBlockEVMLogs()
	// require.Len(t, logs, 0)
}

// testEVMWithdrawWithFailedTx is a helper function that tests the withdraw functionality
// with a failing transaction in the same block.
// If withdrawFirst is true, the withdraw request is run before the failing request,
// otherwise the failing request is run first.
func testEVMWithdrawWithFailedTx(t *testing.T, withdrawFirst bool) {
	env := InitEVM(t)

	env.Chain.DepositAssetsToL2(isc.NewAssets(1*isc.Million), env.Chain.ChainAdmin)

	ethKey, senderEthAddress := env.Chain.NewEthereumAccountWithL2Funds()
	senderAgentID := isc.NewEthereumAddressAgentID(senderEthAddress)

	_, receiver := env.solo.NewKeyPair()

	l1Balance := env.Chain.Env.L1BaseTokens(receiver)
	l2Balance := env.Chain.L2BaseTokens(senderAgentID)
	const tokensWithdrawn = 10_000

	value := util.BaseTokensDecimalsToEthereumDecimals(tokensWithdrawn, parameters.BaseTokenDecimals)
	tx, err := env.ISCMagicSandbox(ethKey).MakeCallFn(
		[]ethCallOptions{{sender: ethKey, value: value, gasLimit: 100_000}},
		"transferToL1",
		receiver,
		iscmagic.WrapISCAssets(isc.NewEmptyAssets()),
	)
	require.NoError(t, err)
	withdrawRequest, err := isc.NewEVMOffLedgerTxRequest(env.Chain.ChainID, tx)
	require.NoError(t, err)

	failingRequest := solo.NewCallParams(accounts.FuncWithdraw.Message()).
		AddAllowance(isc.NewAssets(env.Chain.L2BaseTokens(env.Chain.AdminAgentID())*2)).
		WithMaxAffordableGasBudget().NewRequestOffLedger(env.Chain, env.Chain.ChainAdmin)

	// run both requests in a single block in the specified order
	var requests []isc.Request
	if withdrawFirst {
		requests = []isc.Request{withdrawRequest, failingRequest}
	} else {
		requests = []isc.Request{failingRequest, withdrawRequest}
	}
	_, res := env.Chain.RunOffLedgerRequests(requests)

	require.Len(t, res, 2)

	// Check that the correct receipt has the error
	if withdrawFirst {
		require.Nil(t, res[0].Receipt.Error)
		require.NotNil(t, res[1].Receipt.Error)
		// Gas fee is from the first receipt (withdraw request)
		gasFee := env.Chain.GetRequestReceiptsForBlock(env.Chain.LatestBlockIndex())[0].GasFeeCharged
		require.EqualValues(t, l2Balance-tokensWithdrawn-gasFee, env.Chain.L2BaseTokens(senderAgentID))
	} else {
		require.NotNil(t, res[0].Receipt.Error)
		require.Nil(t, res[1].Receipt.Error)
		// Gas fee is from the second receipt (withdraw request)
		gasFee := env.Chain.GetRequestReceiptsForBlock(env.Chain.LatestBlockIndex())[1].GasFeeCharged
		require.EqualValues(t, l2Balance-tokensWithdrawn-gasFee, env.Chain.L2BaseTokens(senderAgentID))
	}

	// In both cases, the L1 balance of the receiver should be increased by the withdrawn amount
	require.EqualValues(t, l1Balance+tokensWithdrawn, env.Chain.Env.L1BaseTokens(receiver))
}

func TestEVM_WithdrawWithFailedTxAfter(t *testing.T) {
	testEVMWithdrawWithFailedTx(t, true)
}

func TestEVM_WithdrawWithFailedTxBefore(t *testing.T) {
	testEVMWithdrawWithFailedTx(t, false)
}
