// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtest

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"strings"
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

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/evm/evmerrors"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/solo"
	testparameters "github.com/iotaledger/wasp/packages/testutil/parameters"
	"github.com/iotaledger/wasp/packages/testutil/testdbhash"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func TestStorageContract(t *testing.T) {
	env := InitEVM(t, false)
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
	require.EqualValues(t, 3, env.getBlockNumber())

	// call `retrieve` view, get 43
	require.EqualValues(t, 43, storage.retrieve())

	blockNumber := rpc.BlockNumber(env.getBlockNumber())

	// try the view call explicitly passing the EVM block
	{
		for _, v := range []uint32{44, 45, 46} {
			_, err = storage.store(v)
			require.NoError(t, err)
		}
		for _, i := range []uint32{0, 1, 2, 3} {
			var v uint32
			bn := blockNumber + rpc.BlockNumber(i)
			require.NoError(t, storage.callView("retrieve", nil, &v, rpc.BlockNumberOrHashWithNumber(bn)))
			require.EqualValuesf(t, 43+i, v, "blockNumber %d should have counter=%d, got=%d", bn, 43+i, v)
		}
	}
	// same but with blockNumber = -1 (latest block)
	{
		var v uint32
		require.NoError(t, storage.callView("retrieve", nil, &v, rpc.BlockNumberOrHashWithNumber(rpc.LatestBlockNumber)))
		require.EqualValues(t, 46, v)
	}

	testdbhash.VerifyContractStateHash(env.solo, evm.Contract, "", t.Name())
}

func TestLowLevelCallRevert(t *testing.T) {
	env := InitEVM(t, false)
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
	env := InitEVM(t, false)
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
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	erc20 := env.deployERC20Contract(ethKey, "TestCoin", "TEST")

	// get contract bytecode from EVM emulator
	retrievedBytecode := env.getCode(erc20.address)

	// ensure returned bytecode matches the expected runtime bytecode
	require.True(t, bytes.Equal(retrievedBytecode, evmtest.ERC20ContractRuntimeBytecode), "bytecode retrieved from the chain must match the deployed bytecode")
}

func TestGasCharged(t *testing.T) {
	env := InitEVM(t, false)
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
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	storage := env.deployStorageContract(ethKey)

	require.Equal(t, gas.DefaultEVMGasRatio, env.getEVMGasRatio())

	res, err := storage.store(43)
	require.NoError(t, err)
	initialGasFee := res.ISCReceipt.GasFeeCharged

	// only the owner can call the setEVMGasRatio endpoint
	newGasRatio := util.Ratio32{A: gas.DefaultEVMGasRatio.A * 10, B: gas.DefaultEVMGasRatio.B}
	newUserWallet, _ := env.solo.NewKeyPairWithFunds()
	err = env.setEVMGasRatio(newGasRatio, iscCallOptions{wallet: newUserWallet})
	require.True(t, isc.VMErrorIs(err, vm.ErrUnauthorized))
	require.Equal(t, gas.DefaultEVMGasRatio, env.getEVMGasRatio())

	// current owner is able to set a new gasRatio
	err = env.setEVMGasRatio(newGasRatio, iscCallOptions{wallet: env.Chain.OriginatorPrivateKey})
	require.NoError(t, err)
	require.Equal(t, newGasRatio, env.getEVMGasRatio())

	// run an equivalent request and compare the gas fees
	res, err = storage.store(44)
	require.NoError(t, err)
	require.Greater(t, res.ISCReceipt.GasFeeCharged, initialGasFee)
}

// tests that the gas limits are correctly enforced based on the base tokens sent
func TestGasLimit(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	storage := env.deployStorageContract(ethKey)

	// set a gas ratio such that evm gas cost in base tokens is larger than storage deposit cost
	err := env.setEVMGasRatio(util.Ratio32{A: 10, B: 1}, iscCallOptions{wallet: env.Chain.OriginatorPrivateKey})
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
	env := InitEVM(t, false)
	ethKey, ethAddress := env.Chain.NewEthereumAccountWithL2Funds()
	storage := env.deployStorageContract(ethKey)

	_, err := storage.store(43)
	require.NoError(t, err)

	// only the owner can call the setEVMGasRatio endpoint
	// set the ISC gas ratio VERY HIGH
	newGasRatio := util.Ratio32{A: gas.DefaultEVMGasRatio.A * 500, B: gas.DefaultEVMGasRatio.B}
	err = env.setEVMGasRatio(newGasRatio, iscCallOptions{wallet: env.Chain.OriginatorPrivateKey})
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
	iscNonce := env.Chain.Nonce(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddress))
	evmNonce := env.getNonce(ethAddress)
	require.EqualValues(t, iscNonce, evmNonce)
}

// ensure the amount of base tokens sent impacts the amount of gas used
func TestLoop(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	loop := env.deployLoopContract(ethKey)

	gasRatio := env.getEVMGasRatio()

	for _, gasLimit := range []uint64{200000, 400000} {
		baseTokensSent := gas.EVMGasToISC(gasLimit, &gasRatio)
		ethKey2, ethAddr2 := env.Chain.NewEthereumAccountWithL2Funds(baseTokensSent)
		require.EqualValues(t,
			env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr2)),
			baseTokensSent,
		)
		loop.loop(ethCallOptions{
			sender:   ethKey2,
			gasLimit: gasLimit,
		})
		// gas fee is charged regardless of result
		require.Less(t,
			env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr2)),
			baseTokensSent,
		)
	}
}

func TestLoopWithGasLeft(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	gasRatio := env.getEVMGasRatio()
	var usedGas []uint64
	for _, gasLimit := range []uint64{50000, 200000} {
		baseTokensSent := gas.EVMGasToISC(gasLimit, &gasRatio)
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
	env := InitEVM(t, false)
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
	env := InitEVM(t, false)
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
	baseTokensSent := gas.EVMGasToISC(estimatedGas, &gasRatio)
	ethKey2, _ := env.Chain.NewEthereumAccountWithL2Funds(baseTokensSent)
	res, err := iscTest.CallFn([]ethCallOptions{{
		sender:   ethKey2,
		gasLimit: estimatedGas,
	}}, "loopWithGasLeft")
	require.NoError(t, err)
	require.LessOrEqual(t, res.EVMReceipt.GasUsed, estimatedGas)
}

func TestEstimateContractGas(t *testing.T) {
	env := InitEVM(t, true)
	ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()
	contract := env.deployERC20Contract(ethKey, "TEST", "tst")

	base := env.ERC20BaseTokens(ethKey)
	initialBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr))
	_, err := base.CallFn(nil, "transfer", contract.address, big.NewInt(int64(1*isc.Million)))
	require.NoError(t, err)
	require.LessOrEqual(t,
		env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr)),
		initialBalance-1*isc.Million,
	)
	require.EqualValues(t,
		1*isc.Million,
		env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, contract.address)),
	)
	estimatedGas, err := env.evmChain.EstimateGas(ethereum.CallMsg{
		From: contract.address,
		To:   &ethAddr,
	}, nil)
	require.NoError(t, err)
	require.NotZero(t, estimatedGas)
}

func TestCallViewGasLimit(t *testing.T) {
	env := InitEVM(t, false)
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
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

	// deploy the isc-test.sol EVM contract
	iscTest := env.deployISCTestContract(ethKey)

	// call the ISCTest.getChainId() view function of isc-test.sol which in turn:
	//  calls the ISC.getChainId() view function of isc.sol at 0x1074..., which:
	//   returns the ChainID of the underlying ISC chain
	chainID := iscTest.getChainID()

	require.True(t, env.Chain.ChainID.Equals(chainID))
}

func TestISCChainOwnerID(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

	var ret struct {
		iscmagic.ISCAgentID
	}
	env.ISCMagicSandbox(ethKey).callView("getChainOwnerID", nil, &ret)

	chainOwnerID := env.Chain.OriginatorAgentID
	require.True(t, chainOwnerID.Equals(ret.MustUnwrap()))
}

func TestISCTimestamp(t *testing.T) {
	env := InitEVM(t, false)
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
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

	ret := new(iscmagic.ISCDict)
	env.ISCMagicSandbox(ethKey).callView("callView", []interface{}{
		accounts.Contract.Hname(),
		accounts.ViewBalance.Hname(),
		&iscmagic.ISCDict{Items: []iscmagic.ISCDictItem{{
			Key:   []byte(accounts.ParamAgentID),
			Value: env.Chain.OriginatorAgentID.Bytes(),
		}}},
	}, &ret)

	require.NotEmpty(t, ret.Unwrap())
}

func TestISCNFTData(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

	// mint an NFT and send it to the chain
	issuerWallet, issuerAddress := env.solo.NewKeyPairWithFunds()
	metadata := []byte("foobar")
	nft, _, err := env.solo.MintNFTL1(issuerWallet, issuerAddress, metadata)
	require.NoError(t, err)
	_, err = env.Chain.PostRequestSync(
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

func TestISCNFTMint(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	var mintID1 []byte
	iscTest.CallFnExpectEvent(
		[]ethCallOptions{{
			value: big.NewInt(int64(5000000000000 * isc.Million)),
		}},
		"nftMint",
		&mintID1,
		"mintNFT",
	)
	require.NotEmpty(t, mintID1)

	/// produce a block (so the minted nft gets accounted for)
	_, err := iscTest.triggerEvent("Hi from EVM!")
	require.NoError(t, err)
	///

	// assert the event for minting the L1 NFT is issued on the following block
	logs := env.LastBlockEVMLogs()
	require.Len(t, logs, 1)

	evmAccNFTs := env.Chain.L2NFTs(isc.NewEthereumAddressAgentID(env.Chain.ID(), ethAddr))
	require.Len(t, evmAccNFTs, 1)
	nftID1 := evmAccNFTs[0]

	checkTransferEventERC721(
		t,
		logs[0],
		iscmagic.ERC721NFTsAddress,
		common.Address{}, // zero address (mint)
		ethAddr,
		iscmagic.WrapNFTID(nftID1).TokenID(),
	)

	// assert collection contract is NOT created
	collectionAddr := iscmagic.ERC721NFTCollectionAddress(nftID1)
	require.Empty(t, env.getCode(collectionAddr))

	ret := new(iscmagic.ISCNFT)
	env.ISCMagicSandbox(ethKey).callView(
		"getNFTData",
		[]interface{}{iscmagic.WrapNFTID(nftID1)},
		&ret,
	)

	nftData := ret.MustUnwrap()
	var l1NFTData []byte
	chainNFTOuts := env.solo.L1NFTs(env.Chain.ID().AsAddress())
	require.Len(t, chainNFTOuts, 1)
	for _, o := range chainNFTOuts {
		l1NFTData = o.ImmutableFeatureSet().MetadataFeature().Data
	}
	// assert the correct metadata is persisted in L1 and the chain
	require.Equal(t, l1NFTData, nftData.Metadata)

	retIRC27 := new(iscmagic.IRC27NFT)

	err = env.ISCMagicSandbox(ethKey).callView(
		"getIRC27NFTData",
		[]interface{}{iscmagic.WrapNFTID(nftID1)},
		&retIRC27)
	require.NoError(t, err)

	irc27MetaData, err := isc.IRC27NFTMetadataFromBytes(ret.Metadata)
	require.NoError(t, err)

	require.Equal(t, irc27MetaData.Name, retIRC27.Metadata.Name)

	// mint a new NFT using the initial one as the collection, assert the collection contract is created

	// send the collection NFT to the "isctest contract", so it can mint as part of that collection
	{
		erc721 := env.ERC721NFTs(ethKey)
		_, err = erc721.CallFn(nil, "approve", iscTest.address, iscmagic.WrapNFTID(nftID1).TokenID())
		require.NoError(t, err)

		_, err = erc721.CallFn([]ethCallOptions{{
			sender: ethKey,
		}}, "transferFrom", ethAddr, iscTest.address, iscmagic.WrapNFTID(nftID1).TokenID())
		require.NoError(t, err)

		evmAccNFTs = env.Chain.L2NFTs(isc.NewEthereumAddressAgentID(env.Chain.ID(), ethAddr))
		require.Len(t, evmAccNFTs, 0)
	}

	var mintID2 []byte
	iscTest.CallFnExpectEvent(
		[]ethCallOptions{{
			value: big.NewInt(int64(5000000000000 * isc.Million)),
		}},
		"nftMint",
		&mintID2,
		"mintNFTForCollection",
		iscmagic.WrapNFTID(nftID1),
	)
	require.NotEmpty(t, mintID2)

	/// produce a block (so the minted nft gets accounted for)
	_, err = iscTest.triggerEvent("Hi from EVM 2 !")
	require.NoError(t, err)
	///

	evmAccNFTs = env.Chain.L2NFTs(isc.NewEthereumAddressAgentID(env.Chain.ID(), ethAddr))
	require.Len(t, evmAccNFTs, 1)
	nftID2 := evmAccNFTs[0]

	// assert collection contract is created (for the collection NFT only)
	require.NotEmpty(t, env.getCode(collectionAddr))
	require.Empty(t, env.getCode(iscmagic.ERC721NFTCollectionAddress(nftID2)))

	logs = env.LastBlockEVMLogs()
	require.Len(t, logs, 1)

	checkTransferEventERC721(
		t,
		logs[0],
		collectionAddr,
		common.Address{}, // zero address (mint)
		ethAddr,
		iscmagic.WrapNFTID(nftID2).TokenID(),
	)
}

func TestEVMMintNFTToL1(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	someL1Addr := tpkg.RandEd25519Address()

	_, err := iscTest.CallFn([]ethCallOptions{{
		value: big.NewInt(int64(5000000000000 * isc.Million)),
	}}, "mintNFTToL1", someL1Addr[:])

	require.NoError(t, err)

	require.Len(t, env.solo.L1NFTs(someL1Addr), 1)
}

func TestISCTriggerEvent(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// call ISCTest.triggerEvent(string) function of isc-test.sol which in turn:
	//  calls the ISC.iscTriggerEvent(string) function of isc.sol at 0x1074..., which:
	//   triggers an ISC event with the given string parameter
	res, err := iscTest.triggerEvent("Hi from EVM!")
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, res.EVMReceipt.Status)
	events, err := env.Chain.GetEventsForBlock(env.Chain.GetLatestBlockInfo().BlockIndex())
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, string(events[0].Payload), "Hi from EVM!")
}

func TestISCTriggerEventThenFail(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// test that triggerEvent() followed by revert() does not actually trigger the event
	_, err := iscTest.triggerEventFail("Hi from EVM!", ethCallOptions{
		gasLimit: 100_000, // skip estimate gas (which will fail)
	})
	require.Error(t, err)
	events, err := env.Chain.GetEventsForBlock(env.Chain.GetLatestBlockInfo().BlockIndex())
	require.NoError(t, err)
	require.Len(t, events, 0)
}

func TestISCEntropy(t *testing.T) {
	env := InitEVM(t, false)
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
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	wrappedReqID := new(iscmagic.ISCRequestID)
	res := iscTest.CallFnExpectEvent(nil, "RequestIDEvent", &wrappedReqID, "emitRequestID")

	reqid, err := wrappedReqID.Unwrap()
	require.NoError(t, err)

	// check evm log is as expected
	require.NotEqualValues(t, res.EVMReceipt.Logs[0].TxHash, common.Hash{})
	require.NotEqualValues(t, res.EVMReceipt.Logs[0].BlockHash, common.Hash{})

	require.EqualValues(t, env.Chain.LastReceipt().DeserializedRequest().ID(), reqid)
}

func TestReceiptOfFailedTxDoesNotContainEvents(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// set gas policy to a very high price (fails when charging ISC gas)
	{
		feePolicy := env.Chain.GetGasFeePolicy()
		feePolicy.GasPerToken.A = 1
		feePolicy.GasPerToken.B = 1000000000
		err := env.setFeePolicy(*feePolicy)
		require.NoError(t, err)
	}

	res, err := iscTest.CallFn(nil, "emitDummyEvent")
	require.Error(t, err)
	testmisc.RequireErrorToBe(t, err, "gas budget exceeded")
	require.Len(t, res.EVMReceipt.Logs, 0)
}

func TestISCGetSenderAccount(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	var sender struct {
		iscmagic.ISCAgentID
	}
	iscTest.CallFnExpectEvent(nil, "SenderAccountEvent", &sender, "emitSenderAccount")

	require.True(t, env.Chain.LastReceipt().DeserializedRequest().SenderAccount().Equals(sender.MustUnwrap()))
}

func TestSendNonPayableValueTX(t *testing.T) {
	env := InitEVM(t, false)

	ethKey, ethAddress := env.Chain.NewEthereumAccountWithL2Funds()

	// L2 balance of evm core contract is 0
	require.Zero(t, env.Chain.L2BaseTokens(isc.NewContractAgentID(env.Chain.ChainID, evm.Contract.Hname())))

	// L2 balance of ISC magic contract (0x1074...) is 0
	require.Zero(t, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, iscmagic.Address)))

	// initial L2 balance of sender
	senderInitialBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddress))

	// call any function including some value
	value := util.BaseTokensDecimalsToEthereumDecimals(1*isc.Million, parameters.L1().BaseToken.Decimals)

	sandbox := env.ISCMagicSandbox(ethKey)

	res, err := sandbox.CallFn(
		[]ethCallOptions{{sender: ethKey, value: value, gasLimit: 100_000}},
		"getSenderAccount",
	)
	require.Error(t, err, evmimpl.ErrPayingUnpayableMethod)

	// L2 balance of evm core contract is 0
	require.Zero(t, env.Chain.L2BaseTokens(isc.NewContractAgentID(env.Chain.ChainID, evm.Contract.Hname())))
	// L2 balance of ISC magic contract (0x1074...) is 0
	require.Zero(t, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, iscmagic.Address)))
	// L2 balance of common account is: 0
	require.Zero(t, env.Chain.L2BaseTokens(isc.NewContractAgentID(env.Chain.ChainID, 0)))
	// L2 balance of sender is: initial-gasFeeCharged
	require.EqualValues(t, senderInitialBalance-res.ISCReceipt.GasFeeCharged, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddress)))
}

func TestSendPayableValueTX(t *testing.T) {
	env := InitEVM(t, false)

	ethKey, senderEthAddress := env.Chain.NewEthereumAccountWithL2Funds()
	_, receiver := env.solo.NewKeyPair()

	require.Zero(t, env.solo.L1BaseTokens(receiver))
	senderInitialBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, senderEthAddress))

	value := util.BaseTokensDecimalsToEthereumDecimals(1*isc.Million, parameters.L1().BaseToken.Decimals)

	res, err := env.ISCMagicSandbox(ethKey).CallFn(
		[]ethCallOptions{{sender: ethKey, value: value, gasLimit: 100_000}},
		"send", iscmagic.WrapL1Address(receiver),
		iscmagic.WrapISCAssets(isc.NewEmptyAssets()),
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

	decimals := parameters.L1().BaseToken.Decimals
	valueInBaseTokens, bigRemainder := util.EthereumDecimalsToBaseTokenDecimals(
		value,
		decimals,
	)
	require.Zero(t, bigRemainder.BitLen())

	// L2 balance of evm core contract is 0
	require.Zero(t, env.Chain.L2BaseTokens(isc.NewContractAgentID(env.Chain.ChainID, evm.Contract.Hname())))
	// L2 balance of ISC magic contract (0x1074...) is 0 (!!important)
	require.Zero(t, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, iscmagic.Address)))
	// L2 balance of common account is: 0
	require.Zero(t, env.Chain.L2BaseTokens(isc.NewContractAgentID(env.Chain.ChainID, 0)))
	// L2 balance of sender is: initial - value sent in tx - gas fee
	require.EqualValues(t, senderInitialBalance-valueInBaseTokens-res.ISCReceipt.GasFeeCharged, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, senderEthAddress)))
	// L1 balance of receiver is `values sent in tx`
	require.EqualValues(t, valueInBaseTokens, env.solo.L1BaseTokens(receiver))
}

func TestSendTimelock(t *testing.T) {
	env := InitEVM(t, false)

	ethKey, senderEthAddress := env.Chain.NewEthereumAccountWithL2Funds()
	_, receiver := env.solo.NewKeyPair()

	require.Zero(t, env.solo.L1BaseTokens(receiver))
	senderInitialBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, senderEthAddress))

	value := util.BaseTokensDecimalsToEthereumDecimals(1*isc.Million, parameters.L1().BaseToken.Decimals)

	res, err := env.ISCMagicSandbox(ethKey).CallFn(
		[]ethCallOptions{{sender: ethKey, value: value, gasLimit: 100_000}},
		"send", iscmagic.WrapL1Address(receiver),
		iscmagic.WrapISCAssets(isc.NewEmptyAssets()),
		false, // auto adjust SD
		iscmagic.ISCSendMetadata{},
		iscmagic.ISCSendOptions{
			Timelock: 1,
			Expiration: iscmagic.ISCExpiration{
				Time:          0,
				ReturnAddress: iscmagic.L1Address{},
			},
		},
	)
	require.NoError(t, err)

	decimals := parameters.L1().BaseToken.Decimals
	valueInBaseTokens, bigRemainder := util.EthereumDecimalsToBaseTokenDecimals(
		value,
		decimals,
	)
	require.Zero(t, bigRemainder.BitLen())

	// L2 balance of sender is: initial - value sent in tx - gas fee
	require.EqualValues(t, senderInitialBalance-valueInBaseTokens-res.ISCReceipt.GasFeeCharged, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, senderEthAddress)))
}

func TestSendBaseTokens(t *testing.T) {
	env := InitEVM(t, true)

	ethKey, ethAddress := env.Chain.EthereumAccountByIndexWithL2Funds(0)
	_, receiver := env.solo.NewKeyPair(env.solo.NewSeedFromIndex(1))

	iscTest := env.deployISCTestContract(ethKey)

	require.Zero(t, env.solo.L1BaseTokens(receiver))
	senderInitialBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddress))

	// transfer 1 mil from ethAddress L2 to receiver L1
	transfer := 1 * isc.Million

	// attempt the operation without first calling `allow`
	_, err := iscTest.CallFn([]ethCallOptions{{
		gasLimit: 100_000, // skip estimate gas (which will fail)
	}}, "sendBaseTokens", iscmagic.WrapL1Address(receiver), transfer)
	require.Error(t, err)
	require.Contains(t, err.Error(), "remaining allowance insufficient")

	// allow ISCTest to take the tokens
	_, err = env.ISCMagicSandbox(ethKey).CallFn(
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

	testdbhash.VerifyContractStateHash(env.solo, evm.Contract, "", t.Name())

	// attempt again
	const allAllowed = uint64(0)
	_, err = iscTest.CallFn(nil, "sendBaseTokens", iscmagic.WrapL1Address(receiver), allAllowed)
	require.NoError(t, err)
	require.GreaterOrEqual(t, env.solo.L1BaseTokens(receiver), transfer-500) // 500 is the amount of tokens the contract will reserve to pay for the gas fees
	require.LessOrEqual(t, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddress)), senderInitialBalance-transfer)

	// allowance should be empty now
	require.True(t, getAllowanceTo(iscTest.address).IsEmpty())
}

func TestSendBaseTokensAnotherChain(t *testing.T) {
	env := InitEVM(t, false)

	ethKey, ethAddress := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)
	foreignChain := env.solo.NewChain()

	senderInitialBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddress))

	// transfer 1 mil from ethAddress L2 to another chain
	transfer := 1 * isc.Million

	// allow ISCTest to take the tokens
	_, err := env.ISCMagicSandbox(ethKey).CallFn(
		[]ethCallOptions{{sender: ethKey}},
		"allow",
		iscTest.address,
		iscmagic.WrapISCAssets(isc.NewAssetsBaseTokens(transfer)),
	)
	require.NoError(t, err)

	// status of foreign chain before sending:
	bi := foreignChain.GetLatestBlockInfo()
	balanceOnForeignChain := foreignChain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, iscTest.address))
	require.Zero(t, balanceOnForeignChain)

	const allAllowed = uint64(0)
	target := iscmagic.WrapL1Address(foreignChain.ChainID.AsAddress())
	_, err = iscTest.CallFn(nil, "sendBaseTokens", target, allAllowed)
	require.NoError(t, err)
	require.LessOrEqual(t, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddress)), senderInitialBalance-transfer)

	// wait until foreign chain processes the deposit
	foreignChain.WaitUntil(func() bool {
		return foreignChain.GetLatestBlockInfo().BlockIndex() > bi.BlockIndex()
	})

	// assert iscTest contract now has a balance on the foreign chain
	balanceOnForeignChain = foreignChain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, iscTest.address))
	require.Positive(t, balanceOnForeignChain)
}

func TestCannotDepleteAccount(t *testing.T) {
	env := InitEVM(t, false)

	ethKey, ethAddress := env.Chain.NewEthereumAccountWithL2Funds()
	_, receiver := env.solo.NewKeyPair()

	iscTest := env.deployISCTestContract(ethKey)

	require.Zero(t, env.solo.L1BaseTokens(receiver))
	senderInitialBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddress))

	// we eill attempt to transfer so much that we are left with no funds for gas
	transfer := senderInitialBalance - 300

	// allow ISCTest to take the tokens
	_, err := env.ISCMagicSandbox(ethKey).CallFn(
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

	const allAllowed = uint64(0)
	_, err = iscTest.CallFn([]ethCallOptions{{
		gasLimit: 100_000, // skip estimate gas (which will fail)
	}}, "sendBaseTokens", iscmagic.WrapL1Address(receiver), allAllowed)
	require.ErrorContains(t, err, vm.ErrNotEnoughTokensLeftForGas.Error())
}

func TestSendNFT(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()
	ethAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr)

	iscTest := env.deployISCTestContract(ethKey)

	nft, _, err := env.solo.MintNFTL1(env.Chain.OriginatorPrivateKey, env.Chain.OriginatorAddress, []byte("foobar"))
	require.NoError(t, err)
	env.Chain.MustDepositNFT(nft, ethAgentID, env.Chain.OriginatorPrivateKey)

	const storageDeposit uint64 = 10_000

	// allow ISCTest to take the NFT
	_, err = env.ISCMagicSandbox(ethKey).CallFn(
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
	_, err = iscTest.CallFn(nil, "sendNFT",
		iscmagic.WrapL1Address(receiver),
		iscmagic.WrapNFTID(nft.ID),
		storageDeposit,
	)
	require.NoError(t, err)
	require.Empty(t, env.Chain.L2NFTs(ethAgentID))
	require.Equal(t,
		[]iotago.NFTID{nft.ID},
		lo.Map(
			lo.Values(env.solo.L1NFTs(receiver)),
			func(v *iotago.NFTOutput, _ int) iotago.NFTID { return v.NFTID },
		),
	)
	// there must be 2 Transfer events emitted from the ERC721NFTs contract:
	// 1. Transfer NFT ethAddress -> ISCTest
	// 2. Transfer NFT ISCTest -> 0x0 (send to L1)
	{
		blockTxs := lo.Must(env.evmChain.BlockByNumber(nil)).Transactions()
		require.Len(t, blockTxs, 1)
		tx := blockTxs[0]
		receipt := env.evmChain.TransactionReceipt(tx.Hash())
		require.Len(t, receipt.Logs, 2)
		checkTransferEventERC721(
			t,
			receipt.Logs[0],
			iscmagic.ERC721NFTsAddress,
			ethAddr,
			iscTest.address,
			iscmagic.WrapNFTID(nft.ID).TokenID(),
		)
		checkTransferEventERC721(
			t,
			receipt.Logs[1],
			iscmagic.ERC721NFTsAddress,
			iscTest.address,
			common.Address{},
			iscmagic.WrapNFTID(nft.ID).TokenID(),
		)
	}
}

func TestERC721NFTs(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()
	ethAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr)

	erc721 := env.ERC721NFTs(ethKey)

	{
		var n *big.Int
		erc721.callView("balanceOf", []any{ethAddr}, &n)
		require.EqualValues(t, 0, n.Uint64())
	}

	nft, _, err := env.solo.MintNFTL1(env.Chain.OriginatorPrivateKey, env.Chain.OriginatorAddress, []byte("foobar"))
	require.NoError(t, err)
	env.Chain.MustDepositNFT(nft, ethAgentID, env.Chain.OriginatorPrivateKey)

	// there must be a Transfer event emitted from the ERC721NFTs contract
	{
		blockTxs := env.latestEVMTxs()
		require.Len(t, blockTxs, 1)
		tx := blockTxs[0]
		receipt := env.evmChain.TransactionReceipt(tx.Hash())
		require.Len(t, receipt.Logs, 1)
		checkTransferEventERC721(
			t,
			receipt.Logs[0],
			iscmagic.ERC721NFTsAddress,
			common.Address{},
			ethAddr,
			iscmagic.WrapNFTID(nft.ID).TokenID(),
		)
	}

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

	receiverKey, receiverAddr := env.Chain.NewEthereumAccountWithL2Funds()

	{
		_, err2 := erc721.CallFn([]ethCallOptions{{
			sender:   receiverKey,
			gasLimit: 100_000, // skip estimate gas (which will fail)
		}}, "transferFrom", ethAddr, receiverAddr, iscmagic.WrapNFTID(nft.ID).TokenID())
		require.Error(t, err2)
	}

	_, err = erc721.CallFn(nil, "approve", receiverAddr, iscmagic.WrapNFTID(nft.ID).TokenID())
	require.NoError(t, err)

	{
		var a common.Address
		erc721.callView("getApproved", []any{iscmagic.WrapNFTID(nft.ID).TokenID()}, &a)
		require.EqualValues(t, receiverAddr, a)
	}

	_, err = erc721.CallFn([]ethCallOptions{{
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
	env := InitEVM(t, false)

	collectionOwner, collectionOwnerAddr := env.solo.NewKeyPairWithFunds()
	err := env.Chain.DepositBaseTokensToL2(env.solo.L1BaseTokens(collectionOwnerAddr)/2, collectionOwner)
	require.NoError(t, err)

	ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()
	ethAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr)

	collectionMetadata := isc.NewIRC27NFTMetadata(
		"text/html",
		"https://my-awesome-nft-project.com",
		"a string that is longer than 32 bytes",
		[]interface{}{`{"trait_type": "collection", "value": "super"}`},
	)

	collection, collectionInfo, err := env.solo.MintNFTL1(collectionOwner, collectionOwnerAddr, collectionMetadata.Bytes())
	require.NoError(t, err)

	nftMetadatas := []*isc.IRC27NFTMetadata{
		isc.NewIRC27NFTMetadata(
			"application/json",
			"https://my-awesome-nft-project.com/1.json",
			"nft1",
			[]interface{}{`{"trait_type": "Foo", "value": "Bar"}`},
		),
		isc.NewIRC27NFTMetadata(
			"application/json",
			"https://my-awesome-nft-project.com/2.json",
			"nft2",
			[]interface{}{`{"trait_type": "Bar", "value": "Baz"}`},
		),
	}
	allNFTs, _, err := env.solo.MintNFTsL1(collectionOwner, collectionOwnerAddr, &collectionInfo.OutputID,
		lo.Map(nftMetadatas, func(item *isc.IRC27NFTMetadata, index int) []byte {
			return item.Bytes()
		}),
	)
	require.NoError(t, err)

	require.Len(t, allNFTs, 3)
	for _, nft := range allNFTs {
		require.True(t, env.solo.HasL1NFT(collectionOwnerAddr, &nft.ID))
	}

	// deposit the collection NFT in the owner's L2 account
	collectionNFT, _ := lo.Find(allNFTs, func(nft *isc.NFT) bool { return nft.ID == collection.ID })
	env.Chain.MustDepositNFT(collectionNFT, isc.NewAgentID(collectionOwnerAddr), collectionOwner)

	err = env.registerERC721NFTCollection(collectionOwner, collection.ID)
	require.NoError(t, err)

	// should not allow to register again
	err = env.registerERC721NFTCollection(collectionOwner, collection.ID)
	require.ErrorContains(t, err, "already exists")

	// deposit the two nfts of the collection on ethAddr's L2 account
	nfts := func() (nfts []*isc.NFT) {
		for _, nft := range allNFTs {
			if nft.ID == collection.ID {
				continue
			}
			env.Chain.MustDepositNFT(nft, ethAgentID, collectionOwner)
			nfts = append(nfts, nft)

			// there must be a Transfer event emitted from the ERC721NFTCollection contract
			{
				blockTxs := env.latestEVMTxs()
				require.Len(t, blockTxs, 1)
				tx := blockTxs[0]
				receipt := env.evmChain.TransactionReceipt(tx.Hash())
				require.Len(t, receipt.Logs, 1)
				checkTransferEventERC721(
					t,
					receipt.Logs[0],
					iscmagic.ERC721NFTCollectionAddress(collection.ID),
					common.Address{},
					ethAddr,
					iscmagic.WrapNFTID(nft.ID).TokenID(),
				)
			}
		}
		return nfts
	}()
	require.Len(t, nfts, 2)

	// minted NFTs are in random order; find the first one in nftMetadatas
	nft, ok := lo.Find(nfts, func(item *isc.NFT) bool {
		metadata, err2 := isc.IRC27NFTMetadataFromBytes(item.Metadata)
		require.NoError(t, err2)
		return metadata.URI == nftMetadatas[0].URI
	})
	require.True(t, ok)

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

	receiverKey, receiverAddr := env.Chain.NewEthereumAccountWithL2Funds()

	{
		_, err2 := erc721.CallFn([]ethCallOptions{{
			sender:   receiverKey,
			gasLimit: 100_000, // skip estimate gas (which will fail)
		}}, "transferFrom", ethAddr, receiverAddr, iscmagic.WrapNFTID(nft.ID).TokenID())
		require.Error(t, err2)
	}

	_, err = erc721.CallFn(nil, "approve", receiverAddr, iscmagic.WrapNFTID(nft.ID).TokenID())
	require.NoError(t, err)

	{
		var a common.Address
		erc721.callView("getApproved", []any{iscmagic.WrapNFTID(nft.ID).TokenID()}, &a)
		require.EqualValues(t, receiverAddr, a)
	}

	_, err = erc721.CallFn([]ethCallOptions{{
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
		p, err := evm.DecodePackedNFTURI(uri)
		require.NoError(t, err)
		require.EqualValues(t, nftMetadatas[0].URI, p.Image)
		require.EqualValues(t, nftMetadatas[0].Name, p.Name)
		require.EqualValues(t, nftMetadatas[0].Description, p.Description)
		require.EqualValues(t, nftMetadatas[0].Attributes, p.Attributes)
	}
}

func TestISCCall(t *testing.T) {
	env := InitEVM(t, false, inccounter.Processor)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	err := env.Chain.DeployContract(nil, inccounter.Contract.Name, inccounter.Contract.ProgramHash)
	require.NoError(t, err)
	iscTest := env.deployISCTestContract(ethKey)

	res, err := iscTest.CallFn(nil, "callInccounter")
	require.NoError(env.solo.T, err)
	require.Equal(env.solo.T, types.ReceiptStatusSuccessful, res.EVMReceipt.Status)

	r, err := env.Chain.CallView(
		inccounter.Contract.Name,
		inccounter.ViewGetCounter.Name,
	)
	require.NoError(env.solo.T, err)
	require.EqualValues(t, 42, codec.MustDecodeInt64(r.Get(inccounter.VarCounter)))
}

func TestFibonacciContract(t *testing.T) {
	env := InitEVM(t, false)
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
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// credit base tokens to the ISC test contract
	contractAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, iscTest.address)
	env.Chain.GetL2FundsFromFaucet(contractAgentID)
	initialContractBalance := env.Chain.L2BaseTokens(contractAgentID)

	randAgentID := isc.NewAgentID(tpkg.RandEd25519Address())

	nBaseTokens := uint64(100)
	allowance := isc.NewAssetsBaseTokens(nBaseTokens)

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
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

	iscTest := env.deployISCTestContract(ethKey)

	ret, err := iscTest.CallFn([]ethCallOptions{{
		gasLimit: 100_000, // skip estimate gas (which will fail)
	}}, "makeISCPanic")

	require.NotNil(t, ret.EVMReceipt) // evm receipt is produced

	require.Error(t, err)
	require.Equal(t, types.ReceiptStatusFailed, ret.EVMReceipt.Status)
	require.Contains(t, err.Error(), "not delegated to another chain owner")
}

func TestISCSendWithArgs(t *testing.T) {
	env := InitEVM(t, false, inccounter.Processor)
	err := env.Chain.DeployContract(nil, inccounter.Contract.Name, inccounter.Contract.ProgramHash)
	require.NoError(t, err)

	checkCounter := func(c int) {
		ret, err2 := env.Chain.CallView(inccounter.Contract.Name, inccounter.ViewGetCounter.Name)
		require.NoError(t, err2)
		counter := codec.MustDecodeUint64(ret.Get(inccounter.VarCounter))
		require.EqualValues(t, c, counter)
	}
	checkCounter(0)

	ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()
	senderInitialBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr))

	sendBaseTokens := 700 * isc.Million

	blockIndex := env.Chain.LatestBlockIndex()

	ret, err := env.ISCMagicSandbox(ethKey).CallFn(
		nil,
		"send",
		iscmagic.WrapL1Address(env.Chain.ChainID.AsAddress()),
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
	require.Nil(t, ret.ISCReceipt.Error)

	// wait a bit for the request going out of EVM to be processed by ISC
	env.Chain.WaitUntil(func() bool {
		return env.Chain.LatestBlockIndex() == blockIndex+2
	})

	// assert inc counter was incremented
	checkCounter(1)

	senderBalanceAfterSend := env.Chain.L2BaseTokensAtStateIndex(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr), blockIndex+1)
	require.Less(t, senderBalanceAfterSend, senderInitialBalance-sendBaseTokens)

	// the assets are deposited in sender account
	senderFinalBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr))
	require.Greater(t, senderFinalBalance, senderInitialBalance-sendBaseTokens)
}

func TestERC20BaseTokens(t *testing.T) {
	env := InitEVM(t, true)
	ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()

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
			env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr)),
			balance.Uint64(),
		)
	}
	{
		initialBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr))
		_, ethAddr2 := solo.NewEthereumAccount()
		_, err := erc20.CallFn(nil, "transfer", ethAddr2, big.NewInt(int64(1*isc.Million)))
		require.NoError(t, err)
		require.LessOrEqual(t,
			env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr)),
			initialBalance-1*isc.Million,
		)
		require.EqualValues(t,
			1*isc.Million,
			env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr2)),
		)
	}
	{
		initialBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr))
		ethKey2, ethAddr2 := env.Chain.NewEthereumAccountWithL2Funds()
		initialBalance2 := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr2))
		{
			_, err := erc20.CallFn(nil, "approve", ethAddr2, big.NewInt(int64(1*isc.Million)))
			require.NoError(t, err)
			require.Greater(t,
				env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr)),
				initialBalance-1*isc.Million,
			)
			require.EqualValues(t,
				initialBalance2,
				env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr2)),
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
			_, err := erc20.CallFn([]ethCallOptions{{sender: ethKey2}}, "transferFrom", ethAddr, ethAddr3, big.NewInt(int64(amount)))
			require.NoError(t, err)
			require.Less(t,
				initialBalance-1*isc.Million,
				env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr)),
			)
			require.EqualValues(t,
				amount,
				env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr3)),
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

func checkTransferEventERC721(
	t *testing.T,
	log *types.Log,
	contractAddress, from, to common.Address,
	tokenID *big.Int,
) {
	require.Equal(t, contractAddress, log.Address)
	require.Len(t, log.Topics, 4)
	require.Equal(t, crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)")), log.Topics[0])
	require.Equal(t, evmutil.AddressToIndexedTopic(from), log.Topics[1])
	require.Equal(t, evmutil.AddressToIndexedTopic(to), log.Topics[2])
	require.Equal(t, evmutil.ERC721TokenIDToIndexedTopic(tokenID), log.Topics[3])
	require.Empty(t, log.Data)
}

func TestERC20NativeTokens(t *testing.T) {
	env := InitEVM(t, false)

	const (
		tokenName         = "ERC20 Native Token Test"
		tokenTickerSymbol = "ERC20NT"
		tokenDecimals     = 8
	)

	foundryOwner, foundryOwnerAddr := env.solo.NewKeyPairWithFunds()
	err := env.Chain.DepositBaseTokensToL2(env.solo.L1BaseTokens(foundryOwnerAddr)/2, foundryOwner)
	require.NoError(t, err)

	supply := big.NewInt(int64(10 * isc.Million))
	foundrySN, nativeTokenID, err := env.Chain.NewNativeTokenParams(supply).
		WithUser(foundryOwner).
		WithTokenName(tokenName).
		WithTokenSymbol(tokenTickerSymbol).
		WithTokenDecimals(tokenDecimals).
		CreateFoundry()
	require.NoError(t, err)
	err = env.Chain.MintTokens(foundrySN, supply, foundryOwner)
	require.NoError(t, err)

	// should not allow to register again
	err = env.registerERC20NativeToken(foundryOwner, foundrySN, tokenName, tokenTickerSymbol, tokenDecimals)
	require.ErrorContains(t, err, "already exists")

	ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()
	ethAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr)

	err = env.Chain.SendFromL2ToL2Account(isc.NewAssets(0, iotago.NativeTokens{
		&iotago.NativeToken{ID: nativeTokenID, Amount: supply},
	}), ethAgentID, foundryOwner)
	require.NoError(t, err)

	// there must be a Transfer event emitted from the ERC20NativeTokens contract
	{
		blockTxs := env.latestEVMTxs()
		require.Len(t, blockTxs, 1)
		tx := blockTxs[0]
		receipt := env.evmChain.TransactionReceipt(tx.Hash())
		require.Len(t, receipt.Logs, 1)
		checkTransferEventERC20(
			t,
			receipt.Logs[0],
			iscmagic.ERC20NativeTokensAddress(foundrySN),
			common.Address{},
			ethAddr,
			supply,
		)
	}

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

func checkTransferEventERC20(
	t *testing.T,
	log *types.Log,
	contractAddress, from, to common.Address,
	amount *big.Int,
) {
	require.Equal(t, contractAddress, log.Address)

	require.Len(t, log.Topics, 3)
	require.Equal(t, crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)")), log.Topics[0])
	require.Equal(t, evmutil.AddressToIndexedTopic(from), log.Topics[1])
	require.Equal(t, evmutil.AddressToIndexedTopic(to), log.Topics[2])
	require.Equal(t, evmutil.PackUint256(amount), log.Data)
}

// helper to make sandbox calls via EVM in a more readable way
func sandboxCall(t *testing.T, wallet *ecdsa.PrivateKey, sandboxContract *IscContractInstance, contract isc.Hname, entrypoint isc.Hname, params dict.Dict, allowance uint64) {
	evmParams := &iscmagic.ISCDict{}
	for k, v := range params {
		evmParams.Items = append(evmParams.Items, iscmagic.ISCDictItem{Key: []byte(k), Value: v})
	}
	_, err := sandboxContract.CallFn(
		[]ethCallOptions{{sender: wallet}},
		"call",
		contract,
		entrypoint,
		evmParams,
		&iscmagic.ISCAssets{
			BaseTokens: allowance,
		},
	)
	require.NoError(t, err)
}

func TestERC20NativeTokensWithExternalFoundry(t *testing.T) {
	env := InitEVM(t, true)

	const (
		tokenName         = "ERC20 Native Token Test"
		tokenTickerSymbol = "ERC20NT"
		tokenDecimals     = 8
	)

	foundryOwner, foundryOwnerAddr := env.solo.NewKeyPairWithFunds(env.solo.NewSeedFromIndex(1))
	err := env.Chain.DepositBaseTokensToL2(env.solo.L1BaseTokens(foundryOwnerAddr)/2, foundryOwner)
	require.NoError(t, err)

	// need an alias to create a foundry; the easiest way is to create a "disposable" ISC chain
	foundryChain, _ := env.solo.NewChainExt(foundryOwner, 0, "foundryChain")
	// use an ethereum address to create the foundry
	ethKey, ethAddr := foundryChain.EthereumAccountByIndexWithL2Funds(1)

	// create a fake "env" to create a sandbox contractInstance for the foundry chain (I think we should change these creations to be more "functional" and less "OOP")
	// TODO could be improved, but we cannot just do env.ISCMagicSandbox to create a sandbox of the foundry chain. Will keep it this way to minimize conflicts with the 2.0 branch
	parsedABI, err := abi.JSON(strings.NewReader(iscmagic.SandboxABI))
	require.NoError(t, err)
	foundryChainISCMagic := &IscContractInstance{
		EVMContractInstance: &EVMContractInstance{
			chain: &SoloChainEnv{
				t:          t,
				solo:       env.solo,
				Chain:      foundryChain,
				evmChainID: evm.DefaultChainID,
				evmChain:   foundryChain.EVM(),
			},
			defaultSender: nil,
			address:       iscmagic.Address,
			abi:           parsedABI,
		},
	}

	supply := big.NewInt(int64(10 * isc.Million))
	sandboxCall(t, ethKey, foundryChainISCMagic,
		accounts.Contract.Hname(),
		accounts.FuncNativeTokenCreate.Hname(),
		dict.Dict{
			accounts.ParamTokenScheme: codec.EncodeTokenScheme(&iotago.SimpleTokenScheme{
				MaximumSupply: supply,
				MeltedTokens:  big.NewInt(0),
				MintedTokens:  big.NewInt(0),
			}),
			accounts.ParamTokenName:         codec.EncodeString(tokenName),
			accounts.ParamTokenTickerSymbol: codec.EncodeString(tokenTickerSymbol),
			accounts.ParamTokenDecimals:     codec.EncodeUint8(tokenDecimals),
		},
		1*isc.Million, // allowance necessary to cover the foundry creation SD
	)
	// NOTE here we know that the SN must be 1. An ethereum contract calling "FuncFoundryCreateNew" would have to save the return value of that function call and persis the obtained foundrySN into it's state
	foundrySN := uint32(1)
	nativeTokenID, err := foundryChain.GetNativeTokenIDByFoundrySN(foundrySN)
	require.NoError(t, err)

	// use the foundry owner ethereum account to mint tokens
	sandboxCall(t, ethKey, foundryChainISCMagic,
		accounts.Contract.Hname(),
		accounts.FuncNativeTokenModifySupply.Hname(),
		dict.Dict{
			accounts.ParamFoundrySN:      codec.Encode(foundrySN),
			accounts.ParamSupplyDeltaAbs: codec.Encode(supply), // mint the entire supply
		},
		1*isc.Million, // allowance necessary to cover the accounting UTXO created for a first time a new kind of NT is minted
	)

	// foundryChain itself will create a request targeting the test chain
	// this request must be done by the foundry owner (the foundry creator in this case)
	sandboxCall(t, ethKey, foundryChainISCMagic,
		evm.Contract.Hname(),
		evm.FuncRegisterERC20NativeTokenOnRemoteChain.Hname(),
		dict.Dict{
			evm.FieldFoundrySN:         codec.EncodeUint32(foundrySN),
			evm.FieldTokenName:         codec.EncodeString(tokenName),
			evm.FieldTokenTickerSymbol: codec.EncodeString(tokenTickerSymbol),
			evm.FieldTokenDecimals:     codec.EncodeUint8(tokenDecimals),
			evm.FieldTargetAddress:     codec.EncodeAddress(env.Chain.ChainID.AsAddress()), // the target chain is the test chain
		},
		1*isc.Million, // provide funds for cross-chain request SD
	)

	// wait until the test chain handles the request, get the erc20 contract address on the test chain
	var erc20addr common.Address
	if !env.Chain.WaitUntil(func() bool {
		res, err2 := env.Chain.CallView(evm.Contract.Name, evm.FuncGetERC20ExternalNativeTokenAddress.Name,
			evm.FieldNativeTokenID, nativeTokenID[:],
		)
		require.NoError(t, err2)
		if len(res[evm.FieldResult]) == 0 {
			return false
		}
		copy(erc20addr[:], res[evm.FieldResult])
		return true
	}) {
		require.FailNow(t, "could not get ERC20 address on target chain")
	}

	// m = keyISCMagic, e = prefixERC20ExternalNativeTokens
	testdbhash.VerifyContractStateHash(env.solo, evm.Contract, "me", t.Name())

	// send base tokens and the minted native tokens from the foundry chain EVM address to the test chain (same EVM address)

	// save test chain current block, so we can know when it processes the transfer req
	blockIndex := env.Chain.GetLatestBlockInfo().BlockIndex()

	ethAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr)
	baseTokensToTransferOnTestChain := 10 * isc.Million
	metadata := iscmagic.WrapISCSendMetadata(
		isc.SendMetadata{
			TargetContract: accounts.Contract.Hname(),
			EntryPoint:     accounts.FuncTransferAllowanceTo.Hname(),
			Params: dict.Dict{
				accounts.ParamAgentID: codec.Encode(ethAgentID),
			},
			Allowance: &isc.Assets{
				BaseTokens: baseTokensToTransferOnTestChain,
				NativeTokens: []*iotago.NativeToken{
					{ID: nativeTokenID, Amount: supply}, // specify the token to be transferred here
				},
			},
			GasBudget: math.MaxUint64, // allow all gas that can be used
		},
	)

	_, err = foundryChainISCMagic.CallFn(
		[]ethCallOptions{{sender: ethKey}},
		"send",
		iscmagic.WrapL1Address(env.Chain.ChainID.AsAddress()), // target of the "send" call is the test chain
		iscmagic.WrapISCAssets(
			&isc.Assets{
				BaseTokens: baseTokensToTransferOnTestChain + 1*isc.Million, // must add some base tokens in order to pay for the gas on the target chain
				NativeTokens: []*iotago.NativeToken{
					{ID: nativeTokenID, Amount: supply}, // specify the token to be transferred here
				},
			},
		),
		false,
		metadata,
		iscmagic.ISCSendOptions{},
	)
	require.NoError(t, err)

	// there must be a Transfer event emitted from the foundry chain's ERC20NativeTokens contract
	{
		blockTxs := lo.Must(foundryChain.EVM().BlockByNumber(nil)).Transactions()
		require.Len(t, blockTxs, 1)
		tx := blockTxs[0]
		receipt := foundryChain.EVM().TransactionReceipt(tx.Hash())
		require.Len(t, receipt.Logs, 1)
		checkTransferEventERC20(
			t,
			receipt.Logs[0],
			iscmagic.ERC20NativeTokensAddress(foundrySN),
			ethAddr,
			common.Address{},
			supply,
		)
	}

	// wait until chainB handles the request, assert it was processed successfully
	env.Chain.WaitUntil(func() bool {
		return env.Chain.GetLatestBlockInfo().BlockIndex() > blockIndex
	})
	lastBlockReceipts := env.Chain.GetRequestReceiptsForBlock()
	require.Len(t, lastBlockReceipts, 1)
	require.Nil(t, lastBlockReceipts[0].Error)

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
	env *SoloChainEnv,
	erc20 *IscContractInstance,
	nativeTokenID iotago.NativeTokenID,
	tokenName, tokenTickerSymbol string,
	tokenDecimals uint8,
	supply *big.Int,
	ethAgentID isc.AgentID,
) {
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
		eth2AgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr2)
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
		eth2AgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr2)
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
			eth3AgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr3)
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
}

func TestERC20NativeTokensLongName(t *testing.T) {
	env := InitEVM(t, false)

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
}

// test withdrawing ALL EVM balance to a L1 address via the magic contract
func TestEVMWithdrawAll(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, ethAddress := env.Chain.NewEthereumAccountWithL2Funds()
	_, receiver := env.solo.NewKeyPair()

	tokensToWithdraw := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddress))

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
	_, err := env.ISCMagicSandbox(ethKey).CallFn(
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
	require.Regexp(t, vm.ErrNotEnoughTokensLeftForGas.Error(), err.Error())
	iscReceipt := env.Chain.LastReceipt()
	require.Error(t, iscReceipt.Error.AsGoError())
	require.EqualValues(t, 0, env.solo.L1BaseTokens(receiver))

	// retry the request above, but now leave some tokens to pay for the gas fees
	tokensToWithdraw -= 2*iscReceipt.GasFeeCharged + 1 // +1 is needed because of the way gas budget calc works
	metadata.GasBudget = iscReceipt.GasBudget
	_, err = env.ISCMagicSandbox(ethKey).CallFn(
		[]ethCallOptions{{sender: ethKey}},
		"send",
		iscmagic.WrapL1Address(receiver),
		iscmagic.WrapISCAssets(isc.NewAssetsBaseTokens(tokensToWithdraw)),
		false,
		metadata,
		iscmagic.ISCSendOptions{},
	)
	require.NoError(t, err)
	iscReceipt = env.Chain.LastReceipt()
	require.NoError(t, iscReceipt.Error.AsGoError())
	require.EqualValues(t, tokensToWithdraw, env.solo.L1BaseTokens(receiver))
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
			expectedError: "insufficient gas price: got 0, minimum is 10000000000",
			gasBurned:     168098,
			feeCharged:    1681,
		},
		{
			name:          "default policy, gas price too low",
			gasPerToken:   util.Ratio32{A: 100, B: 1}, // default: 1 base token = 100 gas units
			evmGasRatio:   util.Ratio32{A: 1, B: 1},   // default
			txGasPrice:    big.NewInt(9999999999),
			expectedError: "insufficient gas price: got 9999999999, minimum is 10000000000",
			gasBurned:     168098,
			feeCharged:    1681,
		},
		{
			name:        "default policy, gas price just enough",
			gasPerToken: util.Ratio32{A: 100, B: 1}, // default: 1 base token = 100 gas units
			evmGasRatio: util.Ratio32{A: 1, B: 1},   // default
			txGasPrice:  big.NewInt(10000000000),
			gasBurned:   25883,
			feeCharged:  259,
		},
		{
			name:        "default policy, gas price 2x",
			gasPerToken: util.Ratio32{A: 100, B: 1}, // default: 1 base token = 100 gas units
			evmGasRatio: util.Ratio32{A: 1, B: 1},   // default
			txGasPrice:  big.NewInt(2 * 10000000000),
			gasBurned:   25883,
			feeCharged:  2 * 259,
		},
		{
			name:        "default policy, gas price 2x, evmGasRatio cheaper",
			gasPerToken: util.Ratio32{A: 100, B: 1}, // default: 1 base token = 100 gas units
			evmGasRatio: util.Ratio32{A: 1, B: 2},
			txGasPrice:  big.NewInt(2 * 10000000000),
			gasBurned:   (25883 + 1) / 2,
			feeCharged:  2 * 259 / 2,
		},
		{
			name:          "gas more expensive, gas price too low",
			gasPerToken:   util.Ratio32{A: 50, B: 1}, // 1 base token = 50 gas units
			evmGasRatio:   util.Ratio32{A: 1, B: 1},  // default
			txGasPrice:    big.NewInt(19999999999),
			expectedError: "insufficient gas price: got 19999999999, minimum is 20000000000",
			gasBurned:     168098,
			feeCharged:    2 * 1681,
		},
		{
			name:        "gas more expensive, gas price just enough",
			gasPerToken: util.Ratio32{A: 50, B: 1}, // 1 base token = 50 gas units
			evmGasRatio: util.Ratio32{A: 1, B: 1},  // default
			txGasPrice:  big.NewInt(2 * 10000000000),
			gasBurned:   25883,
			feeCharged:  2 * 259,
		},
		{
			name:        "gas more expensive, gas price 2x",
			gasPerToken: util.Ratio32{A: 50, B: 1}, // 1 base token = 50 gas units
			evmGasRatio: util.Ratio32{A: 1, B: 1},  // default
			txGasPrice:  big.NewInt(2 * 2 * 10000000000),
			gasBurned:   25883,
			feeCharged:  2 * 2 * 259,
		},
	} {
		t.Run(v.name, func(t *testing.T) {
			env := InitEVM(t, false)
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
				require.Equal(t, err.Error(), v.expectedError)
			} else {
				require.NoError(t, err)
			}

			iscReceipt := env.Chain.LastReceipt()
			require.EqualValues(t, v.gasBurned, iscReceipt.GasBurned)
			require.EqualValues(t, v.feeCharged, iscReceipt.GasFeeCharged)
		})
	}
}

func TestEVMIntrinsicGas(t *testing.T) {
	env := InitEVM(t, false)
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
	env := InitEVM(t, false)
	ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()
	_, someEthereumAddr := solo.NewEthereumAccount()
	someAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, someEthereumAddr)

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
		testparameters.GetL1ParamsForTesting().BaseToken.Decimals,
	)
	sendTx(value)
	env.Chain.AssertL2BaseTokens(someAgentID, 1*isc.Million)
}

func TestSolidityTransferBaseTokens(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	_, someEthereumAddr := solo.NewEthereumAccount()
	someEthereumAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, someEthereumAddr)

	iscTest := env.deployISCTestContract(ethKey)

	// try sending funds to `someEthereumAddr` by sending a "value tx" to the isc test contract
	oneMillionInEthDecimals := util.BaseTokensDecimalsToEthereumDecimals(
		1*isc.Million,
		testparameters.GetL1ParamsForTesting().BaseToken.Decimals,
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
		testparameters.GetL1ParamsForTesting().BaseToken.Decimals,
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
		isc.NewAssetsBaseTokens(10*isc.Million),
		isc.NewEthereumAddressAgentID(env.Chain.ChainID, iscTest.address),
		l1Wallet,
	)

	tenMillionInEthDecimals := util.BaseTokensDecimalsToEthereumDecimals(
		10*isc.Million,
		testparameters.GetL1ParamsForTesting().BaseToken.Decimals,
	)

	_, err = iscTest.CallFn([]ethCallOptions{{
		sender: ethKey,
	}}, "sendTo", someEthereumAddr, tenMillionInEthDecimals)
	require.NoError(t, err)
	env.Chain.AssertL2BaseTokens(someEthereumAgentID, 11*isc.Million)

	// send more than the balance
	_, err = iscTest.CallFn([]ethCallOptions{{
		sender:   ethKey,
		value:    tenMillionInEthDecimals.Mul(tenMillionInEthDecimals, big.NewInt(10000)),
		gasLimit: 100_000, // provide a gas limit value as the estimation will fail
	}}, "sendTo", someEthereumAddr, big.NewInt(0))
	require.Error(t, err)
	env.Chain.AssertL2BaseTokens(someEthereumAgentID, 11*isc.Million)
}

func TestSendEntireBalance(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()
	_, someEthereumAddr := solo.NewEthereumAccount()
	someEthereumAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, someEthereumAddr)

	// send all initial
	initial := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr))
	// try sending funds to `someEthereumAddr` by sending a "value tx"
	initialBalanceInEthDecimals := util.BaseTokensDecimalsToEthereumDecimals(
		initial,
		testparameters.GetL1ParamsForTesting().BaseToken.Decimals,
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
	currentBalance := env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr))

	currentBalanceInEthDecimals := util.BaseTokensDecimalsToEthereumDecimals(
		currentBalance,
		testparameters.GetL1ParamsForTesting().BaseToken.Decimals,
	)

	estimatedGas, err := env.evmChain.EstimateGas(ethereum.CallMsg{
		From:  ethAddr,
		To:    &someEthereumAddr,
		Value: currentBalanceInEthDecimals,
		Data:  []byte{},
	}, nil)
	require.NoError(t, err)

	feePolicy := env.Chain.GetGasFeePolicy()
	gasPrice := feePolicy.DefaultGasPriceFullDecimals(testparameters.GetL1ParamsForTesting().BaseToken.Decimals)
	tokensForGasBudget := feePolicy.FeeFromGas(estimatedGas, gasPrice, testparameters.GetL1ParamsForTesting().BaseToken.Decimals)

	gasLimit := feePolicy.GasBudgetFromTokens(tokensForGasBudget, gasPrice, testparameters.GetL1ParamsForTesting().BaseToken.Decimals)

	valueToSendInEthDecimals := util.BaseTokensDecimalsToEthereumDecimals(
		currentBalance-tokensForGasBudget,
		testparameters.GetL1ParamsForTesting().BaseToken.Decimals,
	)
	unsignedTx = types.NewTransaction(1, someEthereumAddr, valueToSendInEthDecimals, gasLimit, env.evmChain.GasPrice(), []byte{})
	tx, err = types.SignTx(unsignedTx, evmutil.Signer(big.NewInt(int64(env.evmChainID))), ethKey)
	require.NoError(t, err)
	err = env.evmChain.SendTransaction(tx)
	require.NoError(t, err)
	env.Chain.AssertL2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr), 0)
	env.Chain.AssertL2BaseTokens(someEthereumAgentID, currentBalance-tokensForGasBudget)
}

func TestSolidityRevertMessage(t *testing.T) {
	env := InitEVM(t, false)
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
	env := InitEVM(t, false)
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
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	res, err := iscTest.CallFn([]ethCallOptions{{
		sender: ethKey,
	}}, "testStaticCall")
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, res.EVMReceipt.Status)
	events, err := env.Chain.GetEventsForBlock(env.Chain.GetLatestBlockInfo().BlockIndex())
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, string(events[0].Payload), "non-static")
}

func TestSelfDestruct(t *testing.T) {
	// NOTE: since EIP-6780 self-destruct was deprecated
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.EthereumAccountByIndexWithL2Funds(0)

	iscTest := env.deployISCTestContract(ethKey)
	iscTestAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, iscTest.address)

	// send some tokens to the ISCTest contract
	{
		const baseTokensDepositFee = 500
		k, _ := env.solo.NewKeyPairWithFunds(env.solo.NewSeedFromIndex(1))
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
	require.EqualValues(t, 1*isc.Million, env.Chain.L2BaseTokens(isc.NewEthereumAddressAgentID(env.Chain.ChainID, beneficiary)))

	testdbhash.VerifyContractStateHash(env.solo, evm.Contract, "", t.Name())
}

func TestSelfDestruct6780(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.EthereumAccountByIndexWithL2Funds(0)
	iscTest := env.deployISCTestContract(ethKey)

	var createContractAddr common.Address
	iscTest.CallFnExpectEvent(nil, "TestSelfDestruct6780ContractCreated", &createContractAddr, "testSelfDestruct6780")
	require.Empty(t, env.getCode(createContractAddr))
}

func TestChangeGasLimit(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	storage := env.deployStorageContract(ethKey)

	var blockHashes []common.Hash
	for i := 0; i < 10; i++ {
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
	env := InitEVM(t, false)

	var fee uint64
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

	var fee2 uint64
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
	env := InitEVM(t, false)

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
	env := InitEVM(t, false)

	// issue a view call from an account with no funds
	ethKey, _ := solo.NewEthereumAccount()

	var ret struct {
		iscmagic.ISCAgentID
	}
	err := env.ISCMagicSandbox(ethKey).callView("getChainOwnerID", nil, &ret)
	require.NoError(t, err)
}

func TestGasPrice(t *testing.T) {
	env := InitEVM(t, false)

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
	env := InitEVM(t, false)
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
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

	contract := env.deployERC20ExampleContract(ethKey)

	contractAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, contract.address)
	env.Chain.GetL2FundsFromFaucet(contractAgentID)

	_, err := contract.CallFn(nil, "createFoundry", big.NewInt(1000000), uint64(10_000))
	require.NoError(t, err)

	_, err = contract.CallFn(nil, "registerToken", "TESTCOIN", "TEST", uint8(18), uint64(10_000))
	require.NoError(t, err)

	_, err = contract.CallFn(nil, "mint", big.NewInt(1000), uint64(10_000))
	require.NoError(t, err)

	ethKey2, _ := env.Chain.NewEthereumAccountWithL2Funds()
	isTestContract := env.deployISCTestContract(ethKey2)
	iscTestAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, isTestContract.address)
	env.Chain.GetL2FundsFromFaucet(iscTestAgentID)

	_, err = isTestContract.CallFn(nil, "mint", uint32(1), big.NewInt(1000), uint64(10_000))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unauthorized")
}

func TestMagicContractExamplesWithNativeToken(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

	contract := env.deployERC20ExampleContract(ethKey)

	contractAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, contract.address)
	env.Chain.GetL2FundsFromFaucet(contractAgentID)

	_, err := contract.CallFn(nil, "createNativeTokenFoundry", "TESTCOIN", "TEST", uint8(18), big.NewInt(1000000), uint64(10_000))
	require.NoError(t, err)

	_, err = contract.CallFn(nil, "mint", big.NewInt(1000), uint64(10_000))
	require.NoError(t, err)

	ethKey2, _ := env.Chain.NewEthereumAccountWithL2Funds()
	isTestContract := env.deployISCTestContract(ethKey2)
	iscTestAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, isTestContract.address)
	env.Chain.GetL2FundsFromFaucet(iscTestAgentID)

	_, err = isTestContract.CallFn(nil, "mint", uint32(1), big.NewInt(1000), uint64(10_000))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unauthorized")
}

func TestCaller(t *testing.T) {
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)
	err := env.Chain.TransferAllowanceTo(
		isc.NewAssetsBaseTokens(42),
		isc.NewEthereumAddressAgentID(env.Chain.ChainID, iscTest.address),
		env.Chain.OriginatorPrivateKey,
	)
	require.NoError(t, err)

	_, err = iscTest.CallFn(nil, "testCallViewCaller")
	require.NoError(t, err)
	var r []byte
	err = iscTest.callView("testCallViewCaller", nil, &r)
	require.NoError(t, err)
	require.EqualValues(t, 42, big.NewInt(0).SetBytes(r).Uint64())
}

func TestCustomError(t *testing.T) {
	env := InitEVM(t, false)
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
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)
	res, err := iscTest.CallFn([]ethCallOptions{{
		gasLimit: 100000,
	}}, "emitEventAndRevert")
	require.ErrorContains(t, err, "execution reverted")
	require.Empty(t, res.EVMReceipt.Logs)
}

func TestL1DepositEVM(t *testing.T) {
	env := InitEVM(t, false)
	// ensure that after a deposit to an EVM account, there is a tx/receipt for it to be auditable on the EVM side
	wallet, l1Addr := env.solo.NewKeyPairWithFunds()
	_, ethAddr := solo.NewEthereumAccount()
	amount := 1 * isc.Million
	err := env.Chain.TransferAllowanceTo(
		isc.NewAssetsBaseTokens(amount),
		isc.NewEthereumAddressAgentID(env.Chain.ID(), ethAddr),
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
	a := isc.AgentIDFromReader(rr)
	require.True(t, a.Equals(isc.NewAddressAgentID(l1Addr)))
	var assets isc.Assets
	assets.Read(buf)

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
		util.MustEthereumDecimalsToBaseTokenDecimalsExact(bal, parameters.L1().BaseToken.Decimals),
		assets.BaseTokens)

	evmRec := env.Chain.EVM().TransactionReceipt(tx.Hash())
	require.NotNil(t, evmRec)
	require.Equal(t, types.ReceiptStatusSuccessful, evmRec.Status)
	iscRec := env.Chain.LastReceipt()
	feePolicy := env.Chain.GetGasFeePolicy()
	expectedGas := gas.ISCGasBudgetToEVM(iscRec.GasBurned, &feePolicy.EVMGasRatio)
	require.EqualValues(t, expectedGas, evmRec.GasUsed)

	// issue the same deposit again, assert txHashes do not collide

	err = env.Chain.TransferAllowanceTo(
		isc.NewAssetsBaseTokens(amount),
		isc.NewEthereumAddressAgentID(env.Chain.ID(), ethAddr),
		wallet,
	)
	require.NoError(t, err)

	blockTxs2 := env.latestEVMTxs()
	require.Len(t, blockTxs2, 1)
	tx2 := blockTxs2[0]
	require.NotEqual(t, tx.Hash(), tx2.Hash())
}

func TestDecimalsConversion(t *testing.T) {
	parameters.InitL1(parameters.L1ForTesting)
	env := InitEVM(t, false)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// call any function including 999999999999 wei as value (which is just 1 wei short of 1 base token)
	lessThanOneGlow := new(big.Int).SetUint64(999999999999)
	valueInBaseTokens, remainder := util.EthereumDecimalsToBaseTokenDecimals(
		lessThanOneGlow,
		parameters.L1().BaseToken.Decimals,
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
	env := InitEVM(t, false)
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

func TestDisableMagicWrap(t *testing.T) {
	envWithoutMagicWrap := InitEVM(t, false)
	require.Nil(t, envWithoutMagicWrap.getCode(envWithoutMagicWrap.ERC20BaseTokens(nil).address))

	envWithMagicWrap := InitEVM(t, true)
	require.NotNil(t, envWithMagicWrap.getCode(envWithMagicWrap.ERC20BaseTokens(nil).address))
}

func TestEVMEventOnFailedL1Deposit(t *testing.T) {
	env := InitEVM(t, false)
	_, ethAddr := env.Chain.NewEthereumAccountWithL2Funds()

	// set gas policy to a higher price (so that it can fails when charging ISC gas)
	{
		feePolicy := env.Chain.GetGasFeePolicy()
		feePolicy.GasPerToken.A = 1
		feePolicy.GasPerToken.B = 10
		err := env.setFeePolicy(*feePolicy)
		require.NoError(t, err)
	}
	// mint an NFT and send it to the chain
	issuerWallet, issuerAddress := env.solo.NewKeyPairWithFunds()
	metadata := []byte("foobar")
	nft, _, err := env.solo.MintNFTL1(issuerWallet, issuerAddress, metadata)
	require.NoError(t, err)
	ethAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, ethAddr)

	callParams := solo.NewCallParams(accounts.Contract.Name, accounts.FuncTransferAllowanceTo.Name, accounts.ParamAgentID, codec.Encode(ethAgentID)).
		AddBaseTokens(1_000_000).
		WithNFT(nft).
		WithAllowance(isc.NewEmptyAssets().AddNFTs(nft.ID)).
		WithMaxAffordableGasBudget()

	// do not include enough gas budget (but just enough to execute until the end)
	_, estimatedReceipt, err := env.Chain.EstimateGasOnLedger(callParams, issuerWallet)
	require.NoError(t, err)
	callParams.WithGasBudget(estimatedReceipt.GasBurned - 1)

	_, err = env.Chain.PostRequestSync(callParams, issuerWallet)
	require.Error(t, err)
	require.Contains(t, err.Error(), "gas budget exceeded")

	// assert NO event is issued
	logs := env.LastBlockEVMLogs()
	require.Len(t, logs, 0)
}
