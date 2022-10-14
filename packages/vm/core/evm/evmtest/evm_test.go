// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtest

import (
	"bytes"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func TestStorageContract(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	require.EqualValues(t, 1, env.getBlockNumber()) // evm block number is incremented along with ISC block index

	// deploy solidity `storage` contract
	storage := env.deployStorageContract(ethKey)
	require.EqualValues(t, 2, env.getBlockNumber())

	// call FuncCallView to call EVM contract's `retrieve` view, get 42
	require.EqualValues(t, 42, storage.retrieve())

	// call FuncSendTransaction with EVM tx that calls `store(43)`
	res, err := storage.store(43)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, res.evmReceipt.Status)
	require.EqualValues(t, 3, env.getBlockNumber())

	// call `retrieve` view, get 43
	require.EqualValues(t, 43, storage.retrieve())
}

func TestERC20Contract(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

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

	require.Equal(t, types.ReceiptStatusSuccessful, res.evmReceipt.Status)
	require.Equal(t, 1, len(res.evmReceipt.Logs))

	// call `balanceOf` view => check balance of recipient = 1337 TestCoin
	require.Zero(t, erc20.balanceOf(recipientAddress).Cmp(transferAmount))
}

func TestGetCode(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	erc20 := env.deployERC20Contract(ethKey, "TestCoin", "TEST")

	// get contract bytecode from EVM emulator
	retrievedBytecode := env.getCode(erc20.address)

	// ensure returned bytecode matches the expected runtime bytecode
	require.True(t, bytes.Equal(retrievedBytecode, evmtest.ERC20ContractRuntimeBytecode), "bytecode retrieved from the chain must match the deployed bytecode")
}

func TestGasCharged(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	storage := env.deployStorageContract(ethKey)

	// call `store(999)` with enough gas
	res, err := storage.store(999)
	require.NoError(t, err)
	t.Log("evm gas used:", res.evmReceipt.GasUsed)
	t.Log("isc gas used:", res.iscReceipt.GasBurned)
	t.Log("isc gas fee:", res.iscReceipt.GasFeeCharged)
	require.Greater(t, res.evmReceipt.GasUsed, uint64(0))
	require.Greater(t, res.iscReceipt.GasBurned, uint64(0))
	require.Greater(t, res.iscReceipt.GasFeeCharged, uint64(0))
}

func TestGasRatio(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	storage := env.deployStorageContract(ethKey)

	require.Equal(t, evmtypes.DefaultGasRatio, env.getGasRatio())

	res, err := storage.store(43)
	require.NoError(t, err)
	initialGasFee := res.iscReceipt.GasFeeCharged

	// only the owner can call the setGasRatio endpoint
	newGasRatio := util.Ratio32{A: evmtypes.DefaultGasRatio.A * 10, B: evmtypes.DefaultGasRatio.B}
	newUserWallet, _ := env.solo.NewKeyPairWithFunds()
	err = env.setGasRatio(newGasRatio, iscCallOptions{wallet: newUserWallet})
	require.True(t, isc.VMErrorIs(err, vm.ErrUnauthorized))
	require.Equal(t, evmtypes.DefaultGasRatio, env.getGasRatio())

	// current owner is able to set a new gasRatio
	err = env.setGasRatio(newGasRatio, iscCallOptions{wallet: env.soloChain.OriginatorPrivateKey})
	require.NoError(t, err)
	require.Equal(t, newGasRatio, env.getGasRatio())

	// run an equivalent request and compare the gas fees
	res, err = storage.store(44)
	require.NoError(t, err)
	require.Greater(t, res.iscReceipt.GasFeeCharged, initialGasFee)
}

// tests that the gas limits are correctly enforced based on the base tokens sent
func TestGasLimit(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	storage := env.deployStorageContract(ethKey)

	// set a gas ratio such that evm gas cost in base tokens is larger than storage deposit cost
	err := env.setGasRatio(util.Ratio32{A: 10, B: 1}, iscCallOptions{wallet: env.soloChain.OriginatorPrivateKey})
	require.NoError(t, err)

	// estimate gas by sending a valid tx
	result, err := storage.store(123)
	require.NoError(t, err)
	gasBurned := result.iscReceipt.GasBurned
	fee := result.iscReceipt.GasFeeCharged
	t.Logf("gas: %d, fee: %d", gasBurned, fee)

	// send again with same gas limit but not enough base tokens
	notEnoughBaseTokensForGas := fee * 9 / 10
	ethKey2, _ := env.soloChain.NewEthereumAccountWithL2Funds(notEnoughBaseTokensForGas)
	_, err = storage.store(124, ethCallOptions{sender: ethKey2})
	require.Error(t, err)
	require.Regexp(t, `\bgas\b`, err.Error())
}

func TestNotEnoughISCGas(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	storage := env.deployStorageContract(ethKey)

	_, err := storage.store(43)
	require.NoError(t, err)

	// only the owner can call the setGasRatio endpoint
	// set the ISC gas ratio VERY HIGH
	newGasRatio := util.Ratio32{A: evmtypes.DefaultGasRatio.A * 5000, B: evmtypes.DefaultGasRatio.B}
	err = env.setGasRatio(newGasRatio, iscCallOptions{wallet: env.soloChain.OriginatorPrivateKey})
	require.NoError(t, err)
	require.Equal(t, newGasRatio, env.getGasRatio())

	senderAddress := crypto.PubkeyToAddress(storage.defaultSender.PublicKey)
	nonce := env.getNonce(senderAddress)

	// try to issue a call to store(something) in EVM
	res, err := storage.store(44, ethCallOptions{
		gasLimit: 21204, // provide a gas limit because the estimation will fail
	})

	// the call must fail with "not enough gas"
	require.Error(t, err)
	require.Regexp(t, "gas budget exceeded", err)
	require.Equal(t, nonce+1, env.getNonce(senderAddress))

	// there must be an EVM receipt
	require.NotNil(t, res.evmReceipt)
	require.Equal(t, res.evmReceipt.Status, types.ReceiptStatusFailed)

	// no changes should persist
	require.EqualValues(t, 43, storage.retrieve())
}

// ensure the amount of base tokens sent impacts the amount of gas used
func TestLoop(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	loop := env.deployLoopContract(ethKey)

	gasRatio := env.getGasRatio()

	for _, gasLimit := range []uint64{200000, 400000} {
		baseTokensSent := evmtypes.EVMGasToISC(gasLimit, &gasRatio)
		ethKey2, ethAddr2 := env.soloChain.NewEthereumAccountWithL2Funds(baseTokensSent)
		require.EqualValues(t,
			env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr2)),
			baseTokensSent,
		)
		loop.loop(ethCallOptions{
			sender:   ethKey2,
			gasLimit: gasLimit,
		})
		// gas fee is charged regardless of result
		require.Less(t,
			env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr2)),
			baseTokensSent,
		)
	}
}

func TestMagicContract(t *testing.T) {
	// deploy the evm contract, which starts an EVM chain and automatically
	// deploys the isc.sol EVM contract at address 0x1074
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	// deploy the isc-test.sol EVM contract
	iscTest := env.deployISCTestContract(ethKey)

	// call the ISCTest.getChainId() view function of isc-test.sol which in turn:
	//  calls the ISC.getChainId() view function of isc.sol at 0x1074, which:
	//   returns the ChainID of the underlying ISC chain
	chainID := iscTest.getChainID()

	require.True(t, env.soloChain.ChainID.Equals(chainID))
}

func TestISCChainOwnerID(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	ret := new(iscmagic.ISCAgentID)
	env.MagicContract(ethKey).callView("getChainOwnerID", nil, &ret)

	chainOwnerID := env.soloChain.OriginatorAgentID
	require.True(t, chainOwnerID.Equals(ret.MustUnwrap()))
}

func TestISCTimestamp(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	var ret int64
	env.MagicContract(ethKey).callView("getTimestampUnixSeconds", nil, &ret)

	require.EqualValues(t, env.soloChain.GetLatestBlockInfo().Timestamp.Unix(), ret)
}

func TestISCCallView(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	ret := new(iscmagic.ISCDict)
	env.MagicContract(ethKey).callView("callView", []interface{}{
		accounts.Contract.Hname(),
		accounts.ViewBalance.Hname(),
		&iscmagic.ISCDict{Items: []iscmagic.ISCDictItem{{
			Key:   []byte(accounts.ParamAgentID),
			Value: env.soloChain.OriginatorAgentID.Bytes(),
		}}},
	}, &ret)

	require.NotEmpty(t, ret.Unwrap())
}

func TestISCNFTData(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	// mint an NFT and send it to the chain
	issuerWallet, issuerAddress := env.solo.NewKeyPairWithFunds()
	metadata := []byte("foobar")
	nft, _, err := env.solo.MintNFTL1(issuerWallet, issuerAddress, []byte("foobar"))
	require.NoError(t, err)
	_, err = env.soloChain.PostRequestSync(
		solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).
			AddBaseTokens(100000).
			WithNFT(nft).
			WithMaxAffordableGasBudget().
			WithSender(nft.ID.ToAddress()),
		issuerWallet,
	)
	require.NoError(t, err)

	// call getNFTData from EVM
	ret := new(iscmagic.ISCNFT)
	env.MagicContract(ethKey).callView(
		"getNFTData",
		[]interface{}{iscmagic.WrapNFTID(nft.ID)},
		&ret,
	)

	require.EqualValues(t, nft.ID, ret.MustUnwrap().ID)
	require.True(t, issuerAddress.Equal(ret.MustUnwrap().Issuer))
	require.EqualValues(t, metadata, ret.MustUnwrap().Metadata)
}

func TestISCTriggerEvent(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// call ISCTest.triggerEvent(string) function of isc-test.sol which in turn:
	//  calls the ISC.iscTriggerEvent(string) function of isc.sol at 0x1074, which:
	//   triggers an ISC event with the given string parameter
	res, err := iscTest.triggerEvent("Hi from EVM!")
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, res.evmReceipt.Status)
	ev, err := env.soloChain.GetEventsForBlock(env.soloChain.GetLatestBlockInfo().BlockIndex)
	require.NoError(t, err)
	require.Len(t, ev, 1)
	require.Contains(t, ev[0], "Hi from EVM!")
}

func TestISCTriggerEventThenFail(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// test that triggerEvent() followed by revert() does not actually trigger the event
	_, err := iscTest.triggerEventFail("Hi from EVM!", ethCallOptions{
		gasLimit: 100_000, // skip estimate gas (which will fail)
	})
	require.Error(t, err)
	ev, err := env.soloChain.GetEventsForBlock(env.soloChain.GetLatestBlockInfo().BlockIndex)
	require.NoError(t, err)
	require.Len(t, ev, 0)
}

func TestISCEntropy(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// call the ISCTest.emitEntropy() function of isc-test.sol which in turn:
	//  calls ISC.iscEntropy() function of isc.sol at 0x1074, which:
	//   returns the entropy value from the sandbox
	//  emits an EVM event (aka log) with the entropy value
	var entropy hashing.HashValue
	iscTest.callFnExpectEvent(nil, "EntropyEvent", &entropy, "emitEntropy")

	require.NotEqualValues(t, hashing.NilHash, entropy)
}

func TestISCGetRequestID(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	reqID := new(isc.RequestID)
	res := iscTest.callFnExpectEvent(nil, "RequestIDEvent", &reqID, "emitRequestID")

	// check evm log is as expected
	require.NotEqualValues(t, res.evmReceipt.Logs[0].TxHash, common.Hash{})
	require.NotEqualValues(t, res.evmReceipt.Logs[0].BlockHash, common.Hash{})

	require.EqualValues(t, env.soloChain.LastReceipt().DeserializedRequest().ID(), *reqID)
}

func TestISCGetSenderAccount(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	sender := new(iscmagic.ISCAgentID)
	iscTest.callFnExpectEvent(nil, "SenderAccountEvent", &sender, "emitSenderAccount")

	require.EqualValues(t, iscmagic.WrapISCAgentID(env.soloChain.LastReceipt().DeserializedRequest().SenderAccount()), *sender)
}

func TestSendBaseTokens(t *testing.T) {
	env := initEVM(t, inccounter.Processor)

	ethKey, ethAddress := env.soloChain.NewEthereumAccountWithL2Funds()
	_, receiver := env.solo.NewKeyPair()

	iscTest := env.deployISCTestContract(ethKey)

	require.Zero(t, env.solo.L1BaseTokens(receiver))
	senderInitialBalance := env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddress))

	// transfer 1 mil from ethAddress L2 to receiver L1
	transfer := 1 * isc.Million

	// attempt the operation without first calling `allow`
	_, err := iscTest.callFn([]ethCallOptions{{
		gasLimit: 100_000, // skip estimate gas (which will fail)
	}}, "sendBaseTokens", iscmagic.WrapL1Address(receiver), transfer)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not previously allowed")

	// allow ISCTest to take the tokens
	_, err = env.MagicContract(ethKey).callFn(
		[]ethCallOptions{{sender: ethKey}},
		"allow",
		iscTest.address,
		iscmagic.WrapISCAllowance(isc.NewAllowanceBaseTokens(transfer)),
	)
	require.NoError(t, err)

	getAllowanceTo := func(target common.Address) *isc.Allowance {
		var ret struct{ Allowance iscmagic.ISCAllowance }
		env.MagicContract(ethKey).callView("getAllowanceTo", []interface{}{target}, &ret)
		return ret.Allowance.Unwrap()
	}

	// stored allowance should be == transfer
	require.Equal(t, transfer, getAllowanceTo(iscTest.address).Assets.BaseTokens)

	// attempt again
	const allAllowed = uint64(0)
	_, err = iscTest.callFn(nil, "sendBaseTokens", iscmagic.WrapL1Address(receiver), allAllowed)
	require.NoError(t, err)
	require.GreaterOrEqual(t, env.solo.L1BaseTokens(receiver), transfer-500) // 500 is the amount of tokens the contract will reserve to pay for the gas fees
	require.LessOrEqual(t, env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddress)), senderInitialBalance-transfer)

	// allowance should be empty now
	require.True(t, getAllowanceTo(iscTest.address).IsEmpty())
}

func TestSendAsNFT(t *testing.T) {
	// TODO: how to send an NFT to an ethereum address on L2?
	t.SkipNow()
	/*
		env := initEVM(t, inccounter.Processor)
		ethKey, ethAddress := env.soloChain.NewEthereumAccountWithL2Funds()
		iscTest := env.deployISCTestContract(ethKey)

		err := env.soloChain.DeployContract(nil, inccounter.Contract.Name, inccounter.Contract.ProgramHash)
		require.NoError(t, err)

		// mint an NFT and send to chain
		env.soloChain.MustDepositBaseTokensToL2(10*isc.Mi, nil) // for gas
		issuerWallet, issuerAddress := env.solo.NewKeyPairWithFunds()
		metadata := []byte("foobar")
		nftInfo, err := env.solo.MintNFTL1(issuerWallet, issuerAddress, metadata)
		require.NoError(t, err)

		_, err = iscTest.callFn([]ethCallOptions{{
			iota: iscCallOptions{
				wallet: issuerWallet,
				before: func(cp *solo.CallParams) {
					cp.AddBaseTokens(100000).
						WithNFT(&isc.NFT{
							ID:       nftInfo.NFTID,
							Issuer:   issuerAddress,
							Metadata: metadata,
						}).
						AddAllowanceNFTs(nftInfo.NFTID).
						WithMaxAffordableGasBudget()
				},
			},
		}}, "callSendAsNFT", iscmagic.WrapBaseTokensNFTID(nftInfo.NFTID))
		require.NoError(t, err)
	*/
}

func TestISCCall(t *testing.T) {
	env := initEVM(t, inccounter.Processor)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	err := env.soloChain.DeployContract(nil, inccounter.Contract.Name, inccounter.Contract.ProgramHash)
	require.NoError(t, err)
	iscTest := env.deployISCTestContract(ethKey)

	res, err := iscTest.callFn(nil, "callInccounter")
	require.NoError(env.solo.T, err)
	require.Equal(env.solo.T, types.ReceiptStatusSuccessful, res.evmReceipt.Status)

	r, err := env.soloChain.CallView(
		inccounter.Contract.Name,
		inccounter.ViewGetCounter.Name,
	)
	require.NoError(env.solo.T, err)
	require.EqualValues(t, 42, codec.MustDecodeInt64(r.MustGet(inccounter.VarCounter)))
}

func TestFibonacciContract(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	fibo := env.deployFibonacciContract(ethKey)
	require.EqualValues(t, 2, env.getBlockNumber())

	res, err := fibo.fib(7)
	require.NoError(t, err)
	t.Log("evm gas used:", res.evmReceipt.GasUsed)
	t.Log("isc gas used:", res.iscReceipt.GasBurned)
	t.Log("Isc gas fee:", res.iscReceipt.GasFeeCharged)
}

func TestEVMContractOwnsFundsL2Transfer(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// credit base tokens to the ISC test contract
	contractAgentID := isc.NewEthereumAddressAgentID(iscTest.address)
	env.soloChain.GetL2FundsFromFaucet(contractAgentID)
	initialContractBalance := env.soloChain.L2BaseTokens(contractAgentID)

	randAgentID := isc.NewAgentID(tpkg.RandEd25519Address())

	nBaseTokens := uint64(100)
	allowance := isc.NewAllowanceBaseTokens(nBaseTokens)

	_, err := iscTest.callFn(
		nil,
		"moveToAccount",
		iscmagic.WrapISCAgentID(randAgentID),
		iscmagic.WrapISCAllowance(allowance),
	)
	require.NoError(t, err)

	env.soloChain.AssertL2BaseTokens(randAgentID, nBaseTokens)
	env.soloChain.AssertL2BaseTokens(contractAgentID, initialContractBalance-nBaseTokens)
}

func TestISCPanic(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	iscTest := env.deployISCTestContract(ethKey)

	ret, err := iscTest.callFn([]ethCallOptions{{
		gasLimit: 100_000, // skip estimate gas (which will fail)
	}}, "makeISCPanic")

	require.NotNil(t, ret.evmReceipt) // evm receipt is produced

	require.Error(t, err)
	require.Contains(t, err.Error(), "execution reverted")
}

func TestISCSendWithArgs(t *testing.T) {
	env := initEVM(t, inccounter.Processor)
	err := env.soloChain.DeployContract(nil, inccounter.Contract.Name, inccounter.Contract.ProgramHash)
	require.NoError(t, err)

	checkCounter := func(c int) {
		ret, err := env.soloChain.CallView(inccounter.Contract.Name, inccounter.ViewGetCounter.Name)
		require.NoError(t, err)
		counter := codec.MustDecodeUint64(ret.MustGet(inccounter.VarCounter))
		require.EqualValues(t, c, counter)
	}
	checkCounter(0)

	ethKey, ethAddr := env.soloChain.NewEthereumAccountWithL2Funds()
	senderInitialBalance := env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr))

	sendBaseTokens := 700 * isc.Million

	ret, err := env.MagicContract(ethKey).callFn(
		nil,
		"send",
		iscmagic.WrapL1Address(env.soloChain.ChainID.AsAddress()),
		iscmagic.WrapISCFungibleTokens(*isc.NewFungibleBaseTokens(sendBaseTokens)),
		false, // auto adjust SD
		iscmagic.WrapISCSendMetadata(isc.SendMetadata{
			TargetContract: inccounter.Contract.Hname(),
			EntryPoint:     inccounter.FuncIncCounter.Hname(),
			Params:         dict.Dict{},
			Allowance:      isc.NewEmptyAllowance(),
			GasBudget:      math.MaxUint64,
		}),
		iscmagic.ISCSendOptions{},
	)
	require.NoError(t, err)
	require.Nil(t, ret.iscReceipt.Error)

	senderFinalBalance := env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr))
	require.Less(t, senderFinalBalance, senderInitialBalance-sendBaseTokens)
	time.Sleep(1 * time.Second) // wait a bit for the request going out of EVM to be processed by ISC

	// assert inc counter was incremented
	checkCounter(1)
}

func TestERC20BaseTokens(t *testing.T) {
	env := initEVM(t)
	ethKey, ethAddr := env.soloChain.NewEthereumAccountWithL2Funds()

	erc20 := env.ERC20BaseTokens(ethKey)

	{
		var name string
		erc20.callView("name", nil, &name)
		require.Equal(t, parameters.L1().BaseToken.Name, name)
	}
	{
		var sym string
		erc20.callView("symbol", nil, &sym)
		require.Equal(t, parameters.L1().BaseToken.TickerSymbol, sym)
	}
	{
		var dec uint8
		erc20.callView("decimals", nil, &dec)
		require.EqualValues(t, parameters.L1().BaseToken.Decimals, dec)
	}
	{
		var supply *big.Int
		erc20.callView("totalSupply", nil, &supply)
		require.Equal(t, parameters.L1().Protocol.TokenSupply, supply.Uint64())
	}
	{
		var balance *big.Int
		erc20.callView("balanceOf", []interface{}{ethAddr}, &balance)
		require.EqualValues(t,
			env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr)),
			balance.Uint64(),
		)
	}
	{
		initialBalance := env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr))
		_, ethAddr2 := solo.NewEthereumAccount()
		_, err := erc20.callFn(nil, "transfer", ethAddr2, big.NewInt(int64(1*isc.Million)))
		require.NoError(t, err)
		require.LessOrEqual(t,
			env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr)),
			initialBalance-1*isc.Million,
		)
		require.EqualValues(t,
			1*isc.Million,
			env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr2)),
		)
	}
	{
		initialBalance := env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr))
		ethKey2, ethAddr2 := env.soloChain.NewEthereumAccountWithL2Funds()
		initialBalance2 := env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr2))
		{
			_, err := erc20.callFn(nil, "approve", ethAddr2, big.NewInt(int64(1*isc.Million)))
			require.NoError(t, err)
			require.Greater(t,
				env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr)),
				initialBalance-1*isc.Million,
			)
			require.EqualValues(t,
				initialBalance2,
				env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr2)),
			)
		}

		{
			var allowance *big.Int
			erc20.callView("allowance", []interface{}{ethAddr, ethAddr2}, &allowance)
			require.EqualValues(t,
				1*isc.Million,
				allowance.Uint64(),
			)
		}
		{
			const amount = 100_000
			_, ethAddr3 := solo.NewEthereumAccount()
			_, err := erc20.callFn([]ethCallOptions{{sender: ethKey2}}, "transferFrom", ethAddr, ethAddr3, big.NewInt(int64(amount)))
			require.NoError(t, err)
			require.Less(t,
				initialBalance-1*isc.Million,
				env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr)),
			)
			require.EqualValues(t,
				amount,
				env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr3)),
			)
			{
				var allowance *big.Int
				erc20.callView("allowance", []interface{}{ethAddr, ethAddr2}, &allowance)
				require.EqualValues(t,
					1*isc.Million-amount,
					allowance.Uint64(),
				)
			}
		}
	}
}

// test withdrawing ALL EVM balance to a L1 address via the magic contract
func TestEVMWithdrawAll(t *testing.T) {
	env := initEVM(t)
	ethKey, ethAddress := env.soloChain.NewEthereumAccountWithL2Funds()
	_, receiver := env.solo.NewKeyPair()

	tokensToWithdraw := env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddress))

	// try withdrawing all base tokens
	metadata := iscmagic.WrapISCSendMetadata(
		isc.SendMetadata{
			TargetContract: inccounter.Contract.Hname(),
			EntryPoint:     inccounter.FuncIncCounter.Hname(),
			Params:         dict.Dict{},
			Allowance:      isc.NewEmptyAllowance(),
			GasBudget:      math.MaxUint64,
		},
	)
	_, err := env.MagicContract(ethKey).callFn(
		[]ethCallOptions{{sender: ethKey}},
		"send",
		iscmagic.WrapL1Address(receiver),
		iscmagic.WrapISCFungibleTokens(*isc.NewFungibleBaseTokens(tokensToWithdraw)),
		false,
		metadata,
		iscmagic.ISCSendOptions{},
	)
	// request must fail with an error, and receiver should not receive any funds
	require.Error(t, err)
	require.Regexp(t, "execution reverted", err.Error())
	iscReceipt := env.soloChain.LastReceipt()
	require.Error(t, iscReceipt.Error.AsGoError())
	require.EqualValues(t, 0, env.solo.L1BaseTokens(receiver))

	// retry the request above, but now leave some tokens to pay for the gas fees
	tokensToWithdraw -= 2*iscReceipt.GasFeeCharged + 1 // +1 is needed because of the way gas budget calc works
	metadata.GasBudget = iscReceipt.GasBudget
	_, err = env.MagicContract(ethKey).callFn(
		[]ethCallOptions{{sender: ethKey}},
		"send",
		iscmagic.WrapL1Address(receiver),
		iscmagic.WrapISCFungibleTokens(*isc.NewFungibleBaseTokens(tokensToWithdraw)),
		false,
		metadata,
		iscmagic.ISCSendOptions{},
	)
	require.NoError(t, err)
	iscReceipt = env.soloChain.LastReceipt()
	require.NoError(t, iscReceipt.Error.AsGoError())
	require.EqualValues(t, tokensToWithdraw, env.solo.L1BaseTokens(receiver))
}

func TestEVMNonZeroGasPriceRequest(t *testing.T) {
	env := initEVM(t)
	ethKey, senderAddress := env.soloChain.NewEthereumAccountWithL2Funds()

	// deploy solidity `storage` contract
	storage := env.deployStorageContract(ethKey)

	// call FuncCallView to call EVM contract's `retrieve` view, get 42
	require.EqualValues(t, 42, storage.retrieve())

	// issue a tx with non-0 gas price
	valueToStore := uint32(888)
	gasPrice := big.NewInt(1234) // non 0
	callArguments, err := storage.abi.Pack("store", valueToStore)
	require.NoError(t, err)
	nonce := storage.chain.getNonce(senderAddress)
	unsignedTx := types.NewTransaction(nonce, storage.address, util.Big0, gas.MaxGasPerRequest, gasPrice, callArguments)

	tx, err := types.SignTx(unsignedTx, storage.chain.signer(), ethKey)
	require.NoError(t, err)

	err = storage.chain.evmChain.SendTransaction(tx)
	require.NoError(t, err)

	rec := env.soloChain.LastReceipt()

	require.EqualValues(t, valueToStore, storage.retrieve())

	// assert the gas fee is the same as a normal request (with 0 gas price)
	res, err := storage.store(999)
	require.NoError(t, err)
	require.EqualValues(t, 999, storage.retrieve())
	require.Equal(t, res.iscReceipt.GasBurned, rec.GasBurned)
	require.Equal(t, res.iscReceipt.GasFeeCharged, rec.GasFeeCharged)
}

func TestEVMTransferBaseTokens(t *testing.T) {
	env := initEVM(t)
	ethKey, ethAddr := env.soloChain.NewEthereumAccountWithL2Funds()
	_, someEthereumAddr := solo.NewEthereumAccount()
	someAgentID := isc.NewEthereumAddressAgentID(someEthereumAddr)

	sendTx := func(amount *big.Int) {
		nonce := env.getNonce(ethAddr)
		unsignedTx := types.NewTransaction(nonce, someEthereumAddr, amount, gas.MaxGasPerRequest, util.Big0, []byte{})
		tx, err := types.SignTx(unsignedTx, evmutil.Signer(big.NewInt(int64(env.evmChainID))), ethKey)
		require.NoError(t, err)
		err = env.evmChain.SendTransaction(tx)
		require.NoError(t, err)
	}

	// try to transfer base tokens between 2 ethereum addresses

	// issue a tx with non-0 amount (try to send ETH/basetoken)
	// try sending 1 million base tokens (expressed in ethereum decimals)
	value := util.BaseTokensDecimalsToEthereumDecimals(
		new(big.Int).SetUint64(1*isc.Million),
		int64(parameters.L1ForTesting.BaseToken.Decimals),
	)
	sendTx(value)
	env.soloChain.AssertL2BaseTokens(someAgentID, 1*isc.Million)

	// by default iota/shimmer base token has 6 decimal cases, so anything past the 6th decimal case should be ignored
	valueWithExtraDecimals := big.NewInt(1_000_000_999_999_999_999) // all these 9's will be ignored and only 1 million tokens should be transferred
	sendTx(valueWithExtraDecimals)
	env.soloChain.AssertL2BaseTokens(someAgentID, 2*isc.Million)

	// issue a tx with a too low amount
	lowValue := big.NewInt(999_999_999_999) // all these 9's will be ignored and nothing should be transferred
	sendTx(lowValue)
	env.soloChain.AssertL2BaseTokens(someAgentID, 2*isc.Million)
}

func TestSolidityTransferBaseTokens(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	_, someEthereumAddr := solo.NewEthereumAccount()
	someEthereumAgentID := isc.NewEthereumAddressAgentID(someEthereumAddr)

	iscTest := env.deployISCTestContract(ethKey)

	// try sending funds to `someEthereumAddr` by sending a "value tx" to the isc test contract
	oneMillionInEthDecimals := util.BaseTokensDecimalsToEthereumDecimals(
		new(big.Int).SetUint64(1*isc.Million),
		int64(parameters.L1ForTesting.BaseToken.Decimals),
	)

	_, err := iscTest.callFn([]ethCallOptions{{
		sender: ethKey,
		value:  oneMillionInEthDecimals,
	}}, "sendTo", someEthereumAddr, oneMillionInEthDecimals)
	require.NoError(t, err)
	env.soloChain.AssertL2BaseTokens(someEthereumAgentID, 1*isc.Million)

	// attempt to send more than the contract will have available
	twoMillionInEthDecimals := util.BaseTokensDecimalsToEthereumDecimals(
		new(big.Int).SetUint64(2*isc.Million),
		int64(parameters.L1ForTesting.BaseToken.Decimals),
	)

	_, err = iscTest.callFn([]ethCallOptions{{
		sender: ethKey,
		value:  oneMillionInEthDecimals,
	}}, "sendTo", someEthereumAddr, twoMillionInEthDecimals)
	require.Error(t, err)
	env.soloChain.AssertL2BaseTokens(someEthereumAgentID, 1*isc.Million)

	{
		// try sending a value to too high precision (anything over the 6 decimals will be ignored)
		_, err = iscTest.callFn([]ethCallOptions{{
			sender: ethKey,
			value:  oneMillionInEthDecimals,
			// wei is expressed with 18 decimal precision, iota/smr is 6, so anything in the 12 last decimal cases will be ignored
		}}, "sendTo", someEthereumAddr, big.NewInt(1_000_000_999_999_999_999))
		require.Error(t, err)
		env.soloChain.AssertL2BaseTokens(someEthereumAgentID, 1*isc.Million)
		// this will fail if the (ignored) decimals are above the contract balance,
		// but if we provide enough funds, the call should succeed and the extra decimals should be correctly ignored
		_, err = iscTest.callFn([]ethCallOptions{{
			sender: ethKey,
			value:  twoMillionInEthDecimals,
			// wei is expressed with 18 decimal precision, iota/smr is 6, so anything in the 12 last decimal cases will be ignored
		}}, "sendTo", someEthereumAddr, big.NewInt(1_000_000_999_999_999_999))
		require.NoError(t, err)
		env.soloChain.AssertL2BaseTokens(someEthereumAgentID, 2*isc.Million)
	}

	// fund the contract via a L1 wallet ISC transfer, then call `sendTo` to use those funds
	l1Wallet, _ := env.soloChain.Env.NewKeyPairWithFunds()
	env.soloChain.TransferAllowanceTo(
		isc.NewAllowanceBaseTokens(10*isc.Million),
		isc.NewEthereumAddressAgentID(iscTest.address),
		true,
		l1Wallet,
	)

	tenMillionInEthDecimals := util.BaseTokensDecimalsToEthereumDecimals(
		new(big.Int).SetUint64(10*isc.Million),
		int64(parameters.L1ForTesting.BaseToken.Decimals),
	)

	_, err = iscTest.callFn([]ethCallOptions{{
		sender: ethKey,
	}}, "sendTo", someEthereumAddr, tenMillionInEthDecimals)
	require.NoError(t, err)
	env.soloChain.AssertL2BaseTokens(someEthereumAgentID, 12*isc.Million)

	// send more than the balance
	_, err = iscTest.callFn([]ethCallOptions{{
		sender:   ethKey,
		value:    tenMillionInEthDecimals.Mul(tenMillionInEthDecimals, big.NewInt(10000)),
		gasLimit: 100_000, // provide a gas limit value as the estimation will fail
	}}, "sendTo", someEthereumAddr, big.NewInt(0))
	require.Error(t, err)
	env.soloChain.AssertL2BaseTokens(someEthereumAgentID, 12*isc.Million)
}

func TestSendEntireBalance(t *testing.T) {
	env := initEVM(t)
	ethKey, ethAddr := env.soloChain.NewEthereumAccountWithL2Funds()
	_, someEthereumAddr := solo.NewEthereumAccount()
	someEthereumAgentID := isc.NewEthereumAddressAgentID(someEthereumAddr)

	// send all initial
	initial := env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr))
	// try sending funds to `someEthereumAddr` by sending a "value tx"
	initialBalanceInEthDecimals := util.BaseTokensDecimalsToEthereumDecimals(
		new(big.Int).SetUint64(initial),
		int64(parameters.L1ForTesting.BaseToken.Decimals),
	)

	unsignedTx := types.NewTransaction(0, someEthereumAddr, initialBalanceInEthDecimals, gas.MaxGasPerRequest, util.Big0, []byte{})
	tx, err := types.SignTx(unsignedTx, evmutil.Signer(big.NewInt(int64(env.evmChainID))), ethKey)
	require.NoError(t, err)
	err = env.evmChain.SendTransaction(tx)
	// this will produce an error because there won't be tokens left in the account to pay for gas
	require.Error(t, err)
	evmReceipt, err := env.evmChain.TransactionReceipt(tx.Hash())
	require.NoError(t, err)
	require.Equal(t, evmReceipt.Status, types.ReceiptStatusFailed)
	rec := env.soloChain.LastReceipt()
	require.EqualValues(t, rec.ResolvedError, vm.ErrNotEnoughTokensLeftForGas.Error())
	env.soloChain.AssertL2BaseTokens(someEthereumAgentID, 0)

	// now try sending all balance, minus the funds needed for gas
	currentBalance := env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr))

	currentBalanceInEthDecimals := util.BaseTokensDecimalsToEthereumDecimals(
		new(big.Int).SetUint64(currentBalance),
		int64(parameters.L1ForTesting.BaseToken.Decimals),
	)

	estimatedGas, err := env.evmChain.EstimateGas(ethereum.CallMsg{
		From:     ethAddr,
		To:       &someEthereumAddr,
		GasPrice: evm.GasPrice,
		Value:    currentBalanceInEthDecimals,
		Data:     []byte{},
	})
	require.NoError(t, err)

	gasPerToken := env.soloChain.GetGasFeePolicy().GasPerToken
	tokensForGasBudget := uint64(math.Ceil(float64(estimatedGas) / float64(gasPerToken)))

	gasLimit := env.soloChain.GetGasFeePolicy().GasPerToken * tokensForGasBudget

	valueToSendInEthDecimals := util.BaseTokensDecimalsToEthereumDecimals(
		new(big.Int).SetUint64(currentBalance-tokensForGasBudget),
		int64(parameters.L1ForTesting.BaseToken.Decimals),
	)
	unsignedTx = types.NewTransaction(1, someEthereumAddr, valueToSendInEthDecimals, gasLimit, util.Big0, []byte{})
	tx, err = types.SignTx(unsignedTx, evmutil.Signer(big.NewInt(int64(env.evmChainID))), ethKey)
	require.NoError(t, err)
	err = env.evmChain.SendTransaction(tx)
	require.NoError(t, err)
	env.soloChain.AssertL2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr), 0)
	env.soloChain.AssertL2BaseTokens(someEthereumAgentID, currentBalance-tokensForGasBudget)
}

func TestSolidityRevertMessage(t *testing.T) {
	env := initEVM(t)
	ethKey, ethAddr := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// test the revert reason is shown when invoking eth_call
	callData, err := iscTest.abi.Pack("testRevertReason")
	require.NoError(t, err)
	viewRes, err := env.soloChain.CallView(evm.Contract.Name, evm.FuncCallContract.Name, dict.Dict{
		evm.FieldCallMsg: evmtypes.EncodeCallMsg(ethereum.CallMsg{
			From: ethAddr,
			To:   &iscTest.address,
			Gas:  100_000,
			Data: callData,
		}),
	})
	require.Error(t, err)
	require.EqualValues(t, "execution reverted: foobar", err.Error())
	require.Nil(t, viewRes)

	res, err := iscTest.callFn([]ethCallOptions{{
		gasLimit: 100_000, // needed because gas estimation would fail
	}}, "testRevertReason")
	require.Error(t, err)
	require.EqualValues(t, "execution reverted: foobar", res.iscReceipt.ResolvedError)
}
