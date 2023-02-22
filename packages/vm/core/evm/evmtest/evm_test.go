// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtest

import (
	"bytes"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
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
	testparameters "github.com/iotaledger/wasp/packages/testutil/parameters"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/transaction"
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

	require.Equal(t, gas.DefaultEVMGasRatio, env.getGasRatio())

	res, err := storage.store(43)
	require.NoError(t, err)
	initialGasFee := res.iscReceipt.GasFeeCharged

	// only the owner can call the setGasRatio endpoint
	newGasRatio := util.Ratio32{A: gas.DefaultEVMGasRatio.A * 10, B: gas.DefaultEVMGasRatio.B}
	newUserWallet, _ := env.solo.NewKeyPairWithFunds()
	err = env.setGasRatio(newGasRatio, iscCallOptions{wallet: newUserWallet})
	require.True(t, isc.VMErrorIs(err, vm.ErrUnauthorized))
	require.Equal(t, gas.DefaultEVMGasRatio, env.getGasRatio())

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
	newGasRatio := util.Ratio32{A: gas.DefaultEVMGasRatio.A * 500, B: gas.DefaultEVMGasRatio.B}
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
		baseTokensSent := gas.EVMGasToISC(gasLimit, &gasRatio)
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

func TestLoopWithGasLeft(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	gasRatio := env.getGasRatio()
	var usedGas []uint64
	for _, gasLimit := range []uint64{50000, 200000} {
		baseTokensSent := gas.EVMGasToISC(gasLimit, &gasRatio)
		ethKey2, _ := env.soloChain.NewEthereumAccountWithL2Funds(baseTokensSent)
		res, err := iscTest.callFn([]ethCallOptions{{
			sender:   ethKey2,
			gasLimit: gasLimit,
		}}, "loopWithGasLeft")
		require.NoError(t, err)
		require.NotEmpty(t, res.evmReceipt.Logs)
		usedGas = append(usedGas, res.evmReceipt.GasUsed)
	}
	require.Greater(t, usedGas[1], usedGas[0])
}

func TestLoopWithGasLeftEstimateGas(t *testing.T) {
	env := initEVM(t)
	ethKey, ethAddr := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	callData, err := iscTest.abi.Pack("loopWithGasLeft")
	require.NoError(t, err)
	estimatedGas, err := env.evmChain.EstimateGas(ethereum.CallMsg{
		From: ethAddr,
		To:   &iscTest.address,
		Data: callData,
	})
	require.NoError(t, err)
	require.NotZero(t, estimatedGas)
	t.Log(estimatedGas)

	gasRatio := env.getGasRatio()
	baseTokensSent := gas.EVMGasToISC(estimatedGas, &gasRatio)
	ethKey2, _ := env.soloChain.NewEthereumAccountWithL2Funds(baseTokensSent)
	res, err := iscTest.callFn([]ethCallOptions{{
		sender:   ethKey2,
		gasLimit: estimatedGas,
	}}, "loopWithGasLeft")
	require.NoError(t, err)
	require.LessOrEqual(t, res.evmReceipt.GasUsed, estimatedGas)
}

func TestCallViewGasLimit(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	loop := env.deployLoopContract(ethKey)

	callArguments, err := loop.abi.Pack("loop")
	require.NoError(t, err)
	senderAddress := crypto.PubkeyToAddress(loop.defaultSender.PublicKey)
	callMsg := loop.callMsg(ethereum.CallMsg{
		From:     senderAddress,
		Gas:      math.MaxUint64,
		GasPrice: evm.GasPrice,
		Data:     callArguments,
	})
	_, err = loop.chain.evmChain.CallContract(callMsg, latestBlock)
	require.Contains(t, err.Error(), "out of gas")
}

func TestMagicContract(t *testing.T) {
	// deploy the evm contract, which starts an EVM chain and automatically
	// deploys the isc.sol EVM contract at address 0x10740000...
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	// deploy the isc-test.sol EVM contract
	iscTest := env.deployISCTestContract(ethKey)

	// call the ISCTest.getChainId() view function of isc-test.sol which in turn:
	//  calls the ISC.getChainId() view function of isc.sol at 0x1074..., which:
	//   returns the ChainID of the underlying ISC chain
	chainID := iscTest.getChainID()

	require.True(t, env.soloChain.ChainID.Equals(chainID))
}

func TestISCChainOwnerID(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	var ret struct {
		iscmagic.ISCAgentID
	}
	env.ISCMagicSandbox(ethKey).callView("getChainOwnerID", nil, &ret)

	chainOwnerID := env.soloChain.OriginatorAgentID
	require.True(t, chainOwnerID.Equals(ret.MustUnwrap()))
}

func TestISCTimestamp(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	var ret int64
	env.ISCMagicSandbox(ethKey).callView("getTimestampUnixSeconds", nil, &ret)

	require.EqualValues(t, env.soloChain.GetLatestBlockInfo().Timestamp.Unix(), ret)
}

func TestISCCallView(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	ret := new(iscmagic.ISCDict)
	env.ISCMagicSandbox(ethKey).callView("callView", []interface{}{
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
	env.ISCMagicSandbox(ethKey).callView(
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
	//  calls the ISC.iscTriggerEvent(string) function of isc.sol at 0x1074..., which:
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
	//  calls ISC.iscEntropy() function of isc.sol at 0x1074..., which:
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

	wrappedReqID := new(iscmagic.ISCRequestID)
	res := iscTest.callFnExpectEvent(nil, "RequestIDEvent", &wrappedReqID, "emitRequestID")

	reqid, err := wrappedReqID.Unwrap()
	require.NoError(t, err)

	// check evm log is as expected
	require.NotEqualValues(t, res.evmReceipt.Logs[0].TxHash, common.Hash{})
	require.NotEqualValues(t, res.evmReceipt.Logs[0].BlockHash, common.Hash{})

	require.EqualValues(t, env.soloChain.LastReceipt().DeserializedRequest().ID(), reqid)
}

func TestISCGetSenderAccount(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	var sender struct {
		iscmagic.ISCAgentID
	}
	iscTest.callFnExpectEvent(nil, "SenderAccountEvent", &sender, "emitSenderAccount")

	require.True(t, env.soloChain.LastReceipt().DeserializedRequest().SenderAccount().Equals(sender.MustUnwrap()))
}

func TestSendBaseTokens(t *testing.T) {
	env := initEVM(t)

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
	_, err = env.ISCMagicSandbox(ethKey).callFn(
		[]ethCallOptions{{sender: ethKey}},
		"allow",
		iscTest.address,
		iscmagic.WrapISCAssets(isc.NewAssetsBaseTokens(transfer)),
	)
	require.NoError(t, err)

	getAllowanceTo := func(target common.Address) *isc.Assets {
		var ret struct{ Allowance iscmagic.ISCAssets }
		env.ISCMagicSandbox(ethKey).callView("getAllowanceTo", []interface{}{target}, &ret)
		return ret.Allowance.Unwrap()
	}

	// stored allowance should be == transfer
	require.Equal(t, transfer, getAllowanceTo(iscTest.address).BaseTokens)

	// attempt again
	const allAllowed = uint64(0)
	_, err = iscTest.callFn(nil, "sendBaseTokens", iscmagic.WrapL1Address(receiver), allAllowed)
	require.NoError(t, err)
	require.GreaterOrEqual(t, env.solo.L1BaseTokens(receiver), transfer-500) // 500 is the amount of tokens the contract will reserve to pay for the gas fees
	require.LessOrEqual(t, env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddress)), senderInitialBalance-transfer)

	// allowance should be empty now
	require.True(t, getAllowanceTo(iscTest.address).IsEmpty())
}

func TestSendNFT(t *testing.T) {
	env := initEVM(t)
	ethKey, ethAddr := env.soloChain.NewEthereumAccountWithL2Funds()
	ethAgentID := isc.NewEthereumAddressAgentID(ethAddr)

	iscTest := env.deployISCTestContract(ethKey)

	nft, _, err := env.solo.MintNFTL1(env.soloChain.OriginatorPrivateKey, env.soloChain.OriginatorAddress, []byte("foobar"))
	require.NoError(t, err)
	env.soloChain.MustDepositNFT(nft, ethAgentID, env.soloChain.OriginatorPrivateKey)

	const storageDeposit uint64 = 10_000

	// allow ISCTest to take the NFT
	_, err = env.ISCMagicSandbox(ethKey).callFn(
		[]ethCallOptions{{sender: ethKey}},
		"allow",
		iscTest.address,
		iscmagic.WrapISCAssets(isc.NewAssets(
			storageDeposit,
			nil,
			nft.ID,
		)),
	)
	require.NoError(t, err)

	// send to receiver on L1
	_, receiver := env.solo.NewKeyPair()
	_, err = iscTest.callFn(nil, "sendNFT",
		iscmagic.WrapL1Address(receiver),
		iscmagic.WrapNFTID(nft.ID),
		storageDeposit,
	)
	require.NoError(t, err)
	require.Empty(t, env.soloChain.L2NFTs(ethAgentID))
	require.Equal(t,
		[]iotago.NFTID{nft.ID},
		lo.Map(
			lo.Values(env.solo.L1NFTs(receiver)),
			func(v *iotago.NFTOutput, _ int) iotago.NFTID { return v.NFTID },
		),
	)
}

func TestERC721NFTs(t *testing.T) {
	env := initEVM(t)
	ethKey, ethAddr := env.soloChain.NewEthereumAccountWithL2Funds()
	ethAgentID := isc.NewEthereumAddressAgentID(ethAddr)

	erc721 := env.ERC721NFTs(ethKey)

	{
		var n *big.Int
		erc721.callView("balanceOf", []any{ethAddr}, &n)
		require.EqualValues(t, 0, n.Uint64())
	}

	nft, _, err := env.solo.MintNFTL1(env.soloChain.OriginatorPrivateKey, env.soloChain.OriginatorAddress, []byte("foobar"))
	require.NoError(t, err)
	env.soloChain.MustDepositNFT(nft, ethAgentID, env.soloChain.OriginatorPrivateKey)

	{
		var n *big.Int
		erc721.callView("balanceOf", []any{ethAddr}, &n)
		require.EqualValues(t, 1, n.Uint64())
	}

	{
		var a common.Address
		erc721.callView("ownerOf", []any{iscmagic.WrapNFTID(nft.ID).TokenID()}, &a)
		require.EqualValues(t, ethAddr, a)
	}

	receiverKey, receiverAddr := env.soloChain.NewEthereumAccountWithL2Funds()

	{
		_, err2 := erc721.callFn([]ethCallOptions{{
			sender:   receiverKey,
			gasLimit: 100_000, // skip estimate gas (which will fail)
		}}, "transferFrom", ethAddr, receiverAddr, iscmagic.WrapNFTID(nft.ID).TokenID())
		require.Error(t, err2)
	}

	_, err = erc721.callFn(nil, "approve", receiverAddr, iscmagic.WrapNFTID(nft.ID).TokenID())
	require.NoError(t, err)

	{
		var a common.Address
		erc721.callView("getApproved", []any{iscmagic.WrapNFTID(nft.ID).TokenID()}, &a)
		require.EqualValues(t, receiverAddr, a)
	}

	_, err = erc721.callFn([]ethCallOptions{{
		sender: receiverKey,
	}}, "transferFrom", ethAddr, receiverAddr, iscmagic.WrapNFTID(nft.ID).TokenID())
	require.NoError(t, err)

	{
		var a common.Address
		erc721.callView("getApproved", []any{iscmagic.WrapNFTID(nft.ID).TokenID()}, &a)
		var zero common.Address
		require.EqualValues(t, zero, a)
	}

	{
		var a common.Address
		erc721.callView("ownerOf", []any{iscmagic.WrapNFTID(nft.ID).TokenID()}, &a)
		require.EqualValues(t, receiverAddr, a)
	}
}

func TestERC721NFTCollection(t *testing.T) {
	env := initEVM(t)

	collectionOwner, collectionOwnerAddr := env.solo.NewKeyPairWithFunds()
	err := env.soloChain.DepositBaseTokensToL2(env.solo.L1BaseTokens(collectionOwnerAddr)/2, collectionOwner)
	require.NoError(t, err)

	ethKey, ethAddr := env.soloChain.NewEthereumAccountWithL2Funds()
	ethAgentID := isc.NewEthereumAddressAgentID(ethAddr)

	collectionMetadata := transaction.NewIRC27NFTMetadata(
		"text/html",
		"https://my-awesome-nft-project.com",
		"a string that is longer than 32 bytes",
	)

	collection, collectionInfo, err := env.solo.MintNFTL1(collectionOwner, collectionOwnerAddr, collectionMetadata.MustBytes())
	require.NoError(t, err)

	nftMetadatas := []*transaction.IRC27NFTMetadata{
		transaction.NewIRC27NFTMetadata(
			"application/json",
			"https://my-awesome-nft-project.com/1.json",
			"nft1",
		),
		transaction.NewIRC27NFTMetadata(
			"application/json",
			"https://my-awesome-nft-project.com/2.json",
			"nft2",
		),
	}
	allNFTs, _, err := env.solo.MintNFTsL1(collectionOwner, collectionOwnerAddr, &collectionInfo.OutputID,
		lo.Map(nftMetadatas, func(item *transaction.IRC27NFTMetadata, index int) []byte {
			return item.MustBytes()
		}),
	)
	require.NoError(t, err)

	require.Len(t, allNFTs, 3)
	for _, nft := range allNFTs {
		require.True(t, env.solo.HasL1NFT(collectionOwnerAddr, &nft.ID))
	}

	// deposit all nfts on L2
	nfts := func() []*isc.NFT {
		var nfts []*isc.NFT
		for _, nft := range allNFTs {
			if nft.ID == collection.ID {
				// the collection NFT in the owner's account
				env.soloChain.MustDepositNFT(nft, isc.NewAgentID(collectionOwnerAddr), collectionOwner)
			} else {
				// others in ethAgentID's account
				env.soloChain.MustDepositNFT(nft, ethAgentID, collectionOwner)
				nfts = append(nfts, nft)
			}
		}
		return nfts
	}()
	require.Len(t, nfts, 2)

	// minted NFTs are in random order; find the first one in nftMetadatas
	nft, ok := lo.Find(nfts, func(item *isc.NFT) bool {
		metadata, err2 := transaction.IRC27NFTMetadataFromBytes(item.Metadata)
		require.NoError(t, err2)
		return metadata.URI == nftMetadatas[0].URI
	})
	require.True(t, ok)

	err = env.registerERC721NFTCollection(collectionOwner, collection.ID)
	require.NoError(t, err)

	// should not allow to register again
	err = env.registerERC721NFTCollection(collectionOwner, collection.ID)
	require.ErrorContains(t, err, "already exists")

	erc721 := env.ERC721NFTCollection(ethKey, collection.ID)

	{
		var n *big.Int
		erc721.callView("balanceOf", []any{ethAddr}, &n)
		require.EqualValues(t, 2, n.Uint64())
	}

	{
		var a common.Address
		erc721.callView("ownerOf", []any{iscmagic.WrapNFTID(nft.ID).TokenID()}, &a)
		require.EqualValues(t, ethAddr, a)
	}

	receiverKey, receiverAddr := env.soloChain.NewEthereumAccountWithL2Funds()

	{
		_, err2 := erc721.callFn([]ethCallOptions{{
			sender:   receiverKey,
			gasLimit: 100_000, // skip estimate gas (which will fail)
		}}, "transferFrom", ethAddr, receiverAddr, iscmagic.WrapNFTID(nft.ID).TokenID())
		require.Error(t, err2)
	}

	_, err = erc721.callFn(nil, "approve", receiverAddr, iscmagic.WrapNFTID(nft.ID).TokenID())
	require.NoError(t, err)

	{
		var a common.Address
		erc721.callView("getApproved", []any{iscmagic.WrapNFTID(nft.ID).TokenID()}, &a)
		require.EqualValues(t, receiverAddr, a)
	}

	_, err = erc721.callFn([]ethCallOptions{{
		sender: receiverKey,
	}}, "transferFrom", ethAddr, receiverAddr, iscmagic.WrapNFTID(nft.ID).TokenID())
	require.NoError(t, err)

	{
		var a common.Address
		erc721.callView("getApproved", []any{iscmagic.WrapNFTID(nft.ID).TokenID()}, &a)
		var zero common.Address
		require.EqualValues(t, zero, a)
	}

	{
		var a common.Address
		erc721.callView("ownerOf", []any{iscmagic.WrapNFTID(nft.ID).TokenID()}, &a)
		require.EqualValues(t, receiverAddr, a)
	}

	{
		var name string
		erc721.callView("name", nil, &name)
		require.EqualValues(t, collectionMetadata.Name, name)
	}

	{
		var uri string
		erc721.callView("tokenURI", []any{iscmagic.WrapNFTID(nft.ID).TokenID()}, &uri)
		require.EqualValues(t, nftMetadatas[0].URI, uri)
	}
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
	allowance := isc.NewAssetsBaseTokens(nBaseTokens)

	_, err := iscTest.callFn(
		nil,
		"moveToAccount",
		iscmagic.WrapISCAgentID(randAgentID),
		iscmagic.WrapISCAssets(allowance),
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
		ret, err2 := env.soloChain.CallView(inccounter.Contract.Name, inccounter.ViewGetCounter.Name)
		require.NoError(t, err2)
		counter := codec.MustDecodeUint64(ret.MustGet(inccounter.VarCounter))
		require.EqualValues(t, c, counter)
	}
	checkCounter(0)

	ethKey, ethAddr := env.soloChain.NewEthereumAccountWithL2Funds()
	senderInitialBalance := env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr))

	sendBaseTokens := 700 * isc.Million

	blockIndex := env.soloChain.GetLatestBlockInfo().BlockIndex

	ret, err := env.ISCMagicSandbox(ethKey).callFn(
		nil,
		"send",
		iscmagic.WrapL1Address(env.soloChain.ChainID.AsAddress()),
		iscmagic.WrapISCAssets(isc.NewAssetsBaseTokens(sendBaseTokens)),
		false, // auto adjust SD
		iscmagic.WrapISCSendMetadata(isc.SendMetadata{
			TargetContract: inccounter.Contract.Hname(),
			EntryPoint:     inccounter.FuncIncCounter.Hname(),
			Params:         dict.Dict{},
			Allowance:      isc.NewEmptyAssets(),
			GasBudget:      math.MaxUint64,
		}),
		iscmagic.ISCSendOptions{},
	)
	require.NoError(t, err)
	require.Nil(t, ret.iscReceipt.Error)

	senderFinalBalance := env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(ethAddr))
	require.Less(t, senderFinalBalance, senderInitialBalance-sendBaseTokens)

	// wait a bit for the request going out of EVM to be processed by ISC
	env.soloChain.WaitUntil(func(solo.MempoolInfo) bool {
		return env.soloChain.GetLatestBlockInfo().BlockIndex == blockIndex+2
	})

	// assert inc counter was incremented
	checkCounter(1)
}

func TestERC20BaseTokens(t *testing.T) {
	env := initEVM(t)
	ethKey, ethAddr := env.soloChain.NewEthereumAccountWithL2Funds()

	erc20 := env.ERC20BaseTokens(ethKey)

	{
		var name string
		require.NoError(t, erc20.callView("name", nil, &name))
		require.Equal(t, parameters.L1().BaseToken.Name, name)
	}
	{
		var sym string
		require.NoError(t, erc20.callView("symbol", nil, &sym))
		require.Equal(t, parameters.L1().BaseToken.TickerSymbol, sym)
	}
	{
		var dec uint8
		require.NoError(t, erc20.callView("decimals", nil, &dec))
		require.EqualValues(t, parameters.L1().BaseToken.Decimals, dec)
	}
	{
		var supply *big.Int
		require.NoError(t, erc20.callView("totalSupply", nil, &supply))
		require.Equal(t, parameters.L1().Protocol.TokenSupply, supply.Uint64())
	}
	{
		var balance *big.Int
		require.NoError(t, erc20.callView("balanceOf", []interface{}{ethAddr}, &balance))
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
			require.NoError(t, erc20.callView("allowance", []interface{}{ethAddr, ethAddr2}, &allowance))
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
				require.NoError(t, erc20.callView("allowance", []interface{}{ethAddr, ethAddr2}, &allowance))
				require.EqualValues(t,
					1*isc.Million-amount,
					allowance.Uint64(),
				)
			}
		}
	}
}

func TestERC20NativeTokens(t *testing.T) {
	env := initEVM(t)

	const (
		tokenName         = "ERC20 Native Token Test"
		tokenTickerSymbol = "ERC20NT"
		tokenDecimals     = 8
	)

	foundryOwner, foundryOwnerAddr := env.solo.NewKeyPairWithFunds()
	err := env.soloChain.DepositBaseTokensToL2(env.solo.L1BaseTokens(foundryOwnerAddr)/2, foundryOwner)
	require.NoError(t, err)

	supply := big.NewInt(int64(10 * isc.Million))
	foundrySN, nativeTokenID, err := env.soloChain.NewFoundryParams(supply).WithUser(foundryOwner).CreateFoundry()
	require.NoError(t, err)
	err = env.soloChain.MintTokens(foundrySN, supply, foundryOwner)
	require.NoError(t, err)

	err = env.registerERC20NativeToken(foundryOwner, foundrySN, tokenName, tokenTickerSymbol, tokenDecimals)
	require.NoError(t, err)

	// should not allow to register again
	err = env.registerERC20NativeToken(foundryOwner, foundrySN, tokenName, tokenTickerSymbol, tokenDecimals)
	require.ErrorContains(t, err, "already exists")

	ethKey, ethAddr := env.soloChain.NewEthereumAccountWithL2Funds()
	ethAgentID := isc.NewEthereumAddressAgentID(ethAddr)

	err = env.soloChain.SendFromL2ToL2Account(isc.NewAssets(0, iotago.NativeTokens{
		&iotago.NativeToken{ID: nativeTokenID, Amount: supply},
	}), ethAgentID, foundryOwner)
	require.NoError(t, err)

	{
		sandbox := env.ISCMagicSandbox(ethKey)
		var addr common.Address
		sandbox.callView("erc20NativeTokensAddress", []any{foundrySN}, &addr)
		require.Equal(t, iscmagic.ERC20NativeTokensAddress(foundrySN), addr)
	}

	erc20 := env.ERC20NativeTokens(ethKey, foundrySN)

	testERC20NativeTokens(
		env,
		erc20,
		nativeTokenID,
		tokenName, tokenTickerSymbol,
		tokenDecimals,
		supply,
		ethAgentID,
	)
}

func TestERC20NativeTokensWithExternalFoundry(t *testing.T) {
	env := initEVM(t)

	const (
		tokenName         = "ERC20 Native Token Test"
		tokenTickerSymbol = "ERC20NT"
		tokenDecimals     = 8
	)

	foundryOwner, foundryOwnerAddr := env.solo.NewKeyPairWithFunds()
	err := env.soloChain.DepositBaseTokensToL2(env.solo.L1BaseTokens(foundryOwnerAddr)/2, foundryOwner)
	require.NoError(t, err)

	// need an alias to create a foundry; the easiest way is to create a "disposable" ISC chain
	foundryChain, _, _ := env.solo.NewChainExt(foundryOwner, 0, "foundryChain")
	err = foundryChain.DepositBaseTokensToL2(env.solo.L1BaseTokens(foundryOwnerAddr)/2, foundryOwner)
	require.NoError(t, err)
	supply := big.NewInt(int64(10 * isc.Million))
	foundrySN, nativeTokenID, err := foundryChain.NewFoundryParams(supply).WithUser(foundryOwner).CreateFoundry()
	require.NoError(t, err)
	err = foundryChain.MintTokens(foundrySN, supply, foundryOwner)
	require.NoError(t, err)

	erc20addr, err := env.registerERC20ExternalNativeToken(foundryChain, foundrySN, tokenName, tokenTickerSymbol, tokenDecimals)
	require.NoError(t, err)

	ethKey, ethAddr := env.soloChain.NewEthereumAccountWithL2Funds()
	ethAgentID := isc.NewEthereumAddressAgentID(ethAddr)

	{
		assets := isc.NewAssets(0, iotago.NativeTokens{
			&iotago.NativeToken{ID: nativeTokenID, Amount: supply},
		})
		err = foundryChain.Withdraw(assets, foundryOwner)
		require.NoError(t, err)
		err = env.soloChain.SendFromL1ToL2Account(0, assets, ethAgentID, foundryOwner)
		require.NoError(t, err)
	}

	erc20 := env.ERC20ExternalNativeTokens(ethKey, erc20addr)

	testERC20NativeTokens(
		env,
		erc20,
		nativeTokenID,
		tokenName, tokenTickerSymbol,
		tokenDecimals,
		supply,
		ethAgentID,
	)
}

func testERC20NativeTokens(
	env *soloChainEnv,
	erc20 *iscContractInstance,
	nativeTokenID iotago.NativeTokenID,
	tokenName, tokenTickerSymbol string,
	tokenDecimals uint8,
	supply *big.Int,
	ethAgentID isc.AgentID,
) {
	t := env.t
	ethAddr := ethAgentID.(*isc.EthereumAddressAgentID).EthAddress()

	l2Balance := func(agentID isc.AgentID) uint64 {
		return env.soloChain.L2NativeTokens(agentID, nativeTokenID).Uint64()
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
		eth2AgentID := isc.NewEthereumAddressAgentID(ethAddr2)
		_, err := erc20.callFn(nil, "transfer", ethAddr2, big.NewInt(int64(1*isc.Million)))
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
		ethKey2, ethAddr2 := env.soloChain.NewEthereumAccountWithL2Funds()
		eth2AgentID := isc.NewEthereumAddressAgentID(ethAddr2)
		initialBalance2 := l2Balance(eth2AgentID)
		{
			_, err := erc20.callFn(nil, "approve", ethAddr2, big.NewInt(int64(1*isc.Million)))
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
			eth3AgentID := isc.NewEthereumAddressAgentID(ethAddr3)
			_, err := erc20.callFn([]ethCallOptions{{sender: ethKey2}}, "transferFrom", ethAddr, ethAddr3, big.NewInt(int64(amount)))
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
}

func TestERC20NativeTokensLongName(t *testing.T) {
	env := initEVM(t)

	var (
		tokenName         = strings.Repeat("A", 10_000)
		tokenTickerSymbol = "ERC20NT"
		tokenDecimals     = uint8(8)
	)

	foundryOwner, foundryOwnerAddr := env.solo.NewKeyPairWithFunds()
	err := env.soloChain.DepositBaseTokensToL2(env.solo.L1BaseTokens(foundryOwnerAddr)/2, foundryOwner)
	require.NoError(t, err)

	supply := big.NewInt(int64(10 * isc.Million))

	foundrySN, _, err := env.soloChain.NewFoundryParams(supply).WithUser(foundryOwner).CreateFoundry()
	require.NoError(t, err)

	err = env.registerERC20NativeToken(foundryOwner, foundrySN, tokenName, tokenTickerSymbol, tokenDecimals)
	require.ErrorContains(t, err, "too long")
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
			Allowance:      isc.NewEmptyAssets(),
			GasBudget:      math.MaxUint64,
		},
	)
	_, err := env.ISCMagicSandbox(ethKey).callFn(
		[]ethCallOptions{{
			sender:   ethKey,
			gasLimit: 100_000, // provide a gas limit value as the estimation will fail
		}},
		"send",
		iscmagic.WrapL1Address(receiver),
		iscmagic.WrapISCAssets(isc.NewAssetsBaseTokens(tokensToWithdraw)),
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
	_, err = env.ISCMagicSandbox(ethKey).callFn(
		[]ethCallOptions{{sender: ethKey}},
		"send",
		iscmagic.WrapL1Address(receiver),
		iscmagic.WrapISCAssets(isc.NewAssetsBaseTokens(tokensToWithdraw)),
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
	value := util.CustomTokensDecimalsToEthereumDecimals(
		new(big.Int).SetUint64(1*isc.Million),
		testparameters.GetL1ParamsForTesting().BaseToken.Decimals,
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
	oneMillionInEthDecimals := util.CustomTokensDecimalsToEthereumDecimals(
		new(big.Int).SetUint64(1*isc.Million),
		testparameters.GetL1ParamsForTesting().BaseToken.Decimals,
	)

	_, err := iscTest.callFn([]ethCallOptions{{
		sender: ethKey,
		value:  oneMillionInEthDecimals,
	}}, "sendTo", someEthereumAddr, oneMillionInEthDecimals)
	require.NoError(t, err)
	env.soloChain.AssertL2BaseTokens(someEthereumAgentID, 1*isc.Million)

	// attempt to send more than the contract will have available
	twoMillionInEthDecimals := util.CustomTokensDecimalsToEthereumDecimals(
		new(big.Int).SetUint64(2*isc.Million),
		testparameters.GetL1ParamsForTesting().BaseToken.Decimals,
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
		isc.NewAssetsBaseTokens(10*isc.Million),
		isc.NewEthereumAddressAgentID(iscTest.address),
		l1Wallet,
	)

	tenMillionInEthDecimals := util.CustomTokensDecimalsToEthereumDecimals(
		new(big.Int).SetUint64(10*isc.Million),
		testparameters.GetL1ParamsForTesting().BaseToken.Decimals,
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
	initialBalanceInEthDecimals := util.CustomTokensDecimalsToEthereumDecimals(
		new(big.Int).SetUint64(initial),
		testparameters.GetL1ParamsForTesting().BaseToken.Decimals,
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

	currentBalanceInEthDecimals := util.CustomTokensDecimalsToEthereumDecimals(
		new(big.Int).SetUint64(currentBalance),
		testparameters.GetL1ParamsForTesting().BaseToken.Decimals,
	)

	estimatedGas, err := env.evmChain.EstimateGas(ethereum.CallMsg{
		From:     ethAddr,
		To:       &someEthereumAddr,
		GasPrice: evm.GasPrice,
		Value:    currentBalanceInEthDecimals,
		Data:     []byte{},
	})
	require.NoError(t, err)

	feePolicy := env.soloChain.GetGasFeePolicy()
	tokensForGasBudget := feePolicy.FeeFromGas(estimatedGas)

	gasLimit := feePolicy.GasBudgetFromTokens(tokensForGasBudget)

	valueToSendInEthDecimals := util.CustomTokensDecimalsToEthereumDecimals(
		new(big.Int).SetUint64(currentBalance-tokensForGasBudget),
		testparameters.GetL1ParamsForTesting().BaseToken.Decimals,
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

func TestSolidityTransferCustomBaseTokens(t *testing.T) {
	env := initEVM(t)
	ethKey, ethAddr := env.soloChain.NewEthereumAccountWithL2Funds()
	ethAgentID := isc.NewEthereumAddressAgentID(ethAddr)
	iscTest := env.deployISCTestContract(ethKey)

	// create some custom token, and set it as the chain gas token
	customTokenDecimals := uint32(20) // 2 more decimal cases than ethereum

	foundryOwner, foundryOwnerAddr := env.solo.NewKeyPairWithFunds()
	err := env.soloChain.DepositBaseTokensToL2(env.solo.L1BaseTokens(foundryOwnerAddr)/2, foundryOwner)
	require.NoError(t, err)

	supply := big.NewInt(999999999)
	foundrySN, nativeTokenID, err := env.soloChain.NewFoundryParams(supply).WithUser(foundryOwner).CreateFoundry()
	require.NoError(t, err)

	err = env.soloChain.MintTokens(foundrySN, big.NewInt(1_000_000), foundryOwner)
	require.NoError(t, err)
	env.soloChain.AssertL2NativeTokens(isc.NewAgentID(foundryOwnerAddr), nativeTokenID, big.NewInt(1_000_000))

	env.setFeePolicy(gas.GasFeePolicy{
		GasFeeTokenID:       nativeTokenID,
		GasFeeTokenDecimals: customTokenDecimals,
		GasPerToken:         gas.DefaultGasPerToken,
		ValidatorFeeShare:   0,
		EVMGasRatio:         gas.DefaultEVMGasRatio,
	}, iscCallOptions{
		wallet: env.soloChain.OriginatorPrivateKey,
	})

	// move some of these custom tokens into an ethereum account
	tokensToMoveToEvmAccount := int64(500_000)
	err = env.soloChain.SendFromL2ToL2Account(
		isc.NewAssets(0, iotago.NativeTokens{{
			ID:     nativeTokenID,
			Amount: big.NewInt(tokensToMoveToEvmAccount),
		}}),
		ethAgentID,
		foundryOwner,
	)
	require.NoError(t, err)
	env.soloChain.AssertL2NativeTokens(ethAgentID, nativeTokenID, big.NewInt(tokensToMoveToEvmAccount))

	// try sending funds to `someEthereumAddr` by sending a "value tx" to the isc test contract
	_, someEthereumAddr := solo.NewEthereumAccount()
	someEthereumAgentID := isc.NewEthereumAddressAgentID(someEthereumAddr)

	amountOfTokensToMoveInEVMRequest := uint64(123456)
	amountInEthDecimals := util.CustomTokensDecimalsToEthereumDecimals(
		new(big.Int).SetUint64(amountOfTokensToMoveInEVMRequest),
		customTokenDecimals,
	)
	result, err := iscTest.callFn([]ethCallOptions{{
		sender: ethKey,
		value:  amountInEthDecimals,
	}}, "sendTo", someEthereumAddr, amountInEthDecimals)
	require.NoError(t, err)
	actualTokensMovedInEVMRequest := uint64(123400) // the last 2 decimal cases will be ignored
	env.soloChain.AssertL2NativeTokens(someEthereumAgentID, nativeTokenID, actualTokensMovedInEVMRequest)
	// ensure the gas fees and the tokens moved are the correct ones
	require.EqualValues(t,
		uint64(tokensToMoveToEvmAccount)-result.iscReceipt.GasFeeCharged-actualTokensMovedInEVMRequest,
		env.soloChain.L2Assets(ethAgentID).NativeTokens[0].Amount.Uint64(),
	)
}

func TestSandboxStackOverflow(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	iscTest := env.deployISCTestContract(ethKey)

	// tx contract call
	ret, err := iscTest.callFn([]ethCallOptions{{
		gasLimit: 100_000, // skip estimate gas (which will fail)
	}}, "testStackOverflow")

	testmisc.RequireErrorToBe(t, err, vm.ErrIllegalCall)
	require.NotNil(t, ret.evmReceipt) // evm receipt is produced

	// view call
	err = iscTest.callView("testStackOverflow", nil, nil)
	require.Error(t, err)
	testmisc.RequireErrorToBe(t, err, vm.ErrIllegalCall)
}

func TestStaticCall(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	res, err := iscTest.callFn([]ethCallOptions{{
		sender: ethKey,
	}}, "testStaticCall")
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, res.evmReceipt.Status)
	ev, err := env.soloChain.GetEventsForBlock(env.soloChain.GetLatestBlockInfo().BlockIndex)
	require.NoError(t, err)
	require.Len(t, ev, 1)
	require.Contains(t, ev[0], "non-static")
}

func TestSelfDestruct(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	iscTest := env.deployISCTestContract(ethKey)
	iscTestAgentID := isc.NewEthereumAddressAgentID(iscTest.address)

	// send some tokens to the ISCTest contract
	{
		const baseTokensDepositFee = 100
		k, _ := env.solo.NewKeyPairWithFunds()
		err := env.soloChain.SendFromL1ToL2AccountBaseTokens(baseTokensDepositFee, 1*isc.Million, iscTestAgentID, k)
		require.NoError(t, err)
		require.EqualValues(t, 1*isc.Million, env.soloChain.L2BaseTokens(iscTestAgentID))
	}

	_, beneficiary := solo.NewEthereumAccount()

	require.NotEmpty(t, env.getCode(iscTest.address))

	_, err := iscTest.callFn([]ethCallOptions{{sender: ethKey}}, "testSelfDestruct", beneficiary)
	require.NoError(t, err)

	require.Empty(t, env.getCode(iscTest.address))
	require.Zero(t, env.soloChain.L2BaseTokens(iscTestAgentID))
	require.EqualValues(t, 1*isc.Million, env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(beneficiary)))
}
