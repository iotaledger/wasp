// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtest

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/isccontract"
	"github.com/stretchr/testify/require"
)

func TestStorageContract(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	// deploy solidity `storage` contract
	storage := env.deployStorageContract(ethKey, 42)
	require.EqualValues(t, 1, env.getBlockNumber())

	// call FuncCallView to call EVM contract's `retrieve` view, get 42
	require.EqualValues(t, 42, storage.retrieve())

	// call FuncSendTransaction with EVM tx that calls `store(43)`
	res, err := storage.store(43)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, res.evmReceipt.Status)
	require.EqualValues(t, 2, env.getBlockNumber())

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
	storage := env.deployStorageContract(ethKey, 42)

	// call `store(999)` with enough gas
	res, err := storage.store(999)
	require.NoError(t, err)
	t.Log("evm gas used:", res.evmReceipt.GasUsed)
	t.Log("iscp gas used:", res.iscpReceipt.GasBurned)
	t.Log("iscp gas fee:", res.iscpReceipt.GasFeeCharged)
	require.Greater(t, res.evmReceipt.GasUsed, uint64(0))
	require.Greater(t, res.iscpReceipt.GasBurned, uint64(0))
	require.Greater(t, res.iscpReceipt.GasFeeCharged, uint64(0))
}

func TestGasRatio(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	storage := env.deployStorageContract(ethKey, 42)

	require.Equal(t, evmtypes.DefaultGasRatio, env.getGasRatio())

	res, err := storage.store(43)
	require.NoError(t, err)
	initialGasFee := res.iscpReceipt.GasFeeCharged

	// only the owner can call the setGasRatio endpoint
	newGasRatio := util.Ratio32{A: evmtypes.DefaultGasRatio.A * 10, B: evmtypes.DefaultGasRatio.B}
	newUserWallet, _ := env.solo.NewKeyPairWithFunds()
	err = env.setGasRatio(newGasRatio, iotaCallOptions{wallet: newUserWallet})
	require.True(t, iscp.VMErrorIs(err, vm.ErrUnauthorized))
	require.Equal(t, evmtypes.DefaultGasRatio, env.getGasRatio())

	// current owner is able to set a new gasRatio
	err = env.setGasRatio(newGasRatio, iotaCallOptions{wallet: env.soloChain.OriginatorPrivateKey})
	require.NoError(t, err)
	require.Equal(t, newGasRatio, env.getGasRatio())

	// run an equivalent request and compare the gas fees
	res, err = storage.store(44)
	require.NoError(t, err)
	require.Greater(t, res.iscpReceipt.GasFeeCharged, initialGasFee)
}

// tests that the gas limits are correctly enforced based on the iotas sent
func TestGasLimit(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	storage := env.deployStorageContract(ethKey, 42)

	// set a gas ratio such that evm gas cost in iotas is larger than dust cost
	err := env.setGasRatio(util.Ratio32{A: 10, B: 1}, iotaCallOptions{wallet: env.soloChain.OriginatorPrivateKey})
	require.NoError(t, err)

	// estimate gas by sending a valid tx
	result, err := storage.store(123)
	require.NoError(t, err)
	gas := result.iscpReceipt.GasBurned
	fee := result.iscpReceipt.GasFeeCharged
	t.Logf("gas: %d, fee: %d", gas, fee)

	// send again with same gas limit but not enough iotas
	notEnoughIotasForGas := fee * 9 / 10
	ethKey2, _ := env.soloChain.NewEthereumAccountWithL2Funds(notEnoughIotasForGas)
	_, err = storage.store(124, ethCallOptions{sender: ethKey2})
	require.Error(t, err)
	require.Regexp(t, `\bgas\b`, err.Error())
}

// ensure the amount of iotas sent impacts the amount of gas used
func TestLoop(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	loop := env.deployLoopContract(ethKey)

	gasRatio := env.getGasRatio()

	for _, gasLimit := range []uint64{200000, 400000} {
		iotasSent := evmtypes.EVMGasToISC(gasLimit, &gasRatio) + iscp.EVMGasBookkeeping
		ethKey2, ethAddr2 := env.soloChain.NewEthereumAccountWithL2Funds(iotasSent)
		require.EqualValues(t,
			env.soloChain.L2Iotas(iscp.NewEthereumAddressAgentID(ethAddr2)),
			iotasSent,
		)
		loop.loop(ethCallOptions{
			sender:   ethKey2,
			gasLimit: gasLimit,
		})
		// gas fee is charged regardless of result
		require.Less(t,
			env.soloChain.L2Iotas(iscp.NewEthereumAddressAgentID(ethAddr2)),
			iotasSent,
		)
	}
}

func TestISCContract(t *testing.T) {
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

	ret := new(isccontract.ISCAgentID)
	env.ISCContract(ethKey).callView("getChainOwnerID", nil, &ret)

	chainOwnerID := env.soloChain.OriginatorAgentID
	require.True(t, chainOwnerID.Equals(ret.MustUnwrap()))
}

func TestISCTimestamp(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	var ret int64
	env.ISCContract(ethKey).callView("getTimestampUnixSeconds", nil, &ret)

	require.EqualValues(t, env.soloChain.GetLatestBlockInfo().Timestamp.Unix(), ret)
}

func TestISCGetParam(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	key := string(evm.FieldCallMsg) // callView sends an ISC request including this parameter

	var has bool
	env.ISCContract(ethKey).callView("hasParam", []interface{}{key}, &has)
	require.True(t, has)

	var ret []byte
	env.ISCContract(ethKey).callView("getParam", []interface{}{key}, &ret)
	require.NotEmpty(t, ret)
}

func TestISCCallView(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	ret := new(isccontract.ISCDict)
	env.ISCContract(ethKey).callView("callView", []interface{}{
		accounts.Contract.Hname(),
		accounts.ViewBalance.Hname(),
		&isccontract.ISCDict{Items: []isccontract.ISCDictItem{{
			Key:   []byte(accounts.ParamAgentID),
			Value: env.soloChain.OriginatorAgentID.Bytes(),
		}}},
	}, &ret)

	require.NotEmpty(t, ret.Unwrap())
}

func TestISCLogPanic(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	_, err := env.ISCContract(ethKey).callFn([]ethCallOptions{{
		gasLimit: 100_000, // skip estimate gas (which will fail)
	}}, "logPanic", "Hi from EVM!")

	require.Error(t, err)
	require.Contains(t, err.Error(), "Hi from EVM!")
}

func TestISCNFTData(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	// mint an NFT and send it to the chain
	issuerWallet, issuerAddress := env.solo.NewKeyPairWithFunds()
	metadata := []byte("foobar")
	nftInfo, err := env.solo.MintNFTL1(issuerWallet, issuerAddress, []byte("foobar"))
	require.NoError(t, err)
	_, err = env.soloChain.PostRequestSync(
		solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).
			AddIotas(100000).
			WithNFT(&iscp.NFT{
				ID:       nftInfo.NFTID,
				Issuer:   issuerAddress,
				Metadata: metadata,
			}).
			WithMaxAffordableGasBudget().
			WithSender(nftInfo.NFTID.ToAddress()),
		issuerWallet,
	)
	require.NoError(t, err)

	// call getNFTData from EVM
	ret := new(isccontract.ISCNFT)
	env.ISCContract(ethKey).callView(
		"getNFTData",
		[]interface{}{isccontract.WrapIotaNFTID(nftInfo.NFTID)},
		&ret,
	)

	require.EqualValues(t, nftInfo.NFTID, ret.MustUnwrap().ID)
	require.True(t, issuerAddress.Equal(ret.MustUnwrap().Issuer))
	require.EqualValues(t, metadata, ret.MustUnwrap().Metadata)
}

func TestISCTriggerEvent(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// call ISCTest.triggerEvent(string) function of isc-test.sol which in turn:
	//  calls the ISC.iscpTriggerEvent(string) function of isc.sol at 0x1074, which:
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
	//  calls ISC.iscpEntropy() function of isc.sol at 0x1074, which:
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

	reqID := new(iscp.RequestID)
	iscTest.callFnExpectEvent(nil, "RequestIDEvent", &reqID, "emitRequestID")

	require.EqualValues(t, env.soloChain.LastReceipt().Request.ID(), *reqID)
}

func TestISCGetCaller(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	agentID := new(isccontract.ISCAgentID)
	iscTest.callFnExpectEvent(nil, "GetCallerEvent", &agentID, "emitGetCaller")

	senderAddress := crypto.PubkeyToAddress(iscTest.defaultSender.PublicKey)
	require.True(t, agentID.MustUnwrap().Equals(iscp.NewEthereumAddressAgentID(senderAddress)))
}

func TestISCGetSenderAccount(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	sender := new(isccontract.ISCAgentID)
	iscTest.callFnExpectEvent(nil, "SenderAccountEvent", &sender, "emitSenderAccount")

	require.EqualValues(t, isccontract.WrapISCAgentID(env.soloChain.LastReceipt().Request.SenderAccount()), *sender)
}

func TestISCGetAllowanceIotas(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	var iotas uint64
	iscTest.callFnExpectEvent([]ethCallOptions{{
		allowance: iscp.NewAllowanceIotas(42),
	}}, "AllowanceIotasEvent", &iotas, "emitAllowanceIotas")

	require.EqualValues(t, 42, iotas)
}

func TestISCGetAllowanceAvailableIotas(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	var iotasAvailable uint64
	iscTest.callFnExpectEvent([]ethCallOptions{{
		allowance: iscp.NewAllowanceIotas(42),
	}}, "AllowanceAvailableIotasEvent", &iotasAvailable, "emitAllowanceAvailableIotas")

	require.EqualValues(t, 42, iotasAvailable)
}

func TestRevert(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	err := iscTest.callFnExpectError([]ethCallOptions{{
		gasLimit: 100_000, // skip estimate gas (which will fail)
	}}, "emitRevertVMError")

	t.Log(err.Error())
	require.Error(t, err)
	require.Contains(t, err.Error(), "execution reverted")
}

func TestSend(t *testing.T) {
	env := initEVM(t, inccounter.Processor)
	err := env.soloChain.DeployContract(nil, inccounter.Contract.Name, inccounter.Contract.ProgramHash)
	require.NoError(t, err)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)
	_, receiver := env.solo.NewKeyPair()
	require.Zero(t, env.solo.L1Iotas(receiver))
	iscTest.callFn(nil, "send", isccontract.WrapIotaAddress(receiver))
	require.GreaterOrEqual(t, env.solo.L1Iotas(receiver), uint64(1024))
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
		env.soloChain.MustDepositIotasToL2(10*iscp.Mi, nil) // for gas
		issuerWallet, issuerAddress := env.solo.NewKeyPairWithFunds()
		metadata := []byte("foobar")
		nftInfo, err := env.solo.MintNFTL1(issuerWallet, issuerAddress, metadata)
		require.NoError(t, err)

		_, err = iscTest.callFn([]ethCallOptions{{
			iota: iotaCallOptions{
				wallet: issuerWallet,
				before: func(cp *solo.CallParams) {
					cp.AddIotas(100000).
						WithNFT(&iscp.NFT{
							ID:       nftInfo.NFTID,
							Issuer:   issuerAddress,
							Metadata: metadata,
						}).
						AddAllowanceNFTs(nftInfo.NFTID).
						WithMaxAffordableGasBudget()
				},
			},
		}}, "callSendAsNFT", isccontract.WrapIotaNFTID(nftInfo.NFTID))
		require.NoError(t, err)
	*/
}

func TestISCGetAllowanceAvailableNativeTokens(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	iscTest := env.deployISCTestContract(ethKey)

	// mint some native tokens
	env.soloChain.MustDepositIotasToL2(10*iscp.Mi, nil) // for gas
	sn, tokenID, err := env.soloChain.NewFoundryParams(10000).
		WithUser(env.soloChain.OriginatorPrivateKey).
		CreateFoundry()
	require.NoError(t, err)
	err = env.soloChain.MintTokens(sn, 10000, env.soloChain.OriginatorPrivateKey)
	require.NoError(t, err)

	nt := new(isccontract.IotaNativeToken)
	iscTest.callFnExpectEvent([]ethCallOptions{{
		allowance: iscp.NewAllowanceFungibleTokens(iscp.NewEmptyAssets().AddNativeTokens(tokenID, 42)),
	}}, "AllowanceAvailableNativeTokenEvent", &nt, "emitAllowanceAvailableNativeTokens")

	require.EqualValues(t, tokenID[:], nt.ID.Data)
	require.EqualValues(t, 42, nt.Amount.Uint64())
}

func TestISCGetAllowanceNFTs(t *testing.T) {
	// TODO: how to send an NFT to an ethereum address on L2?
	t.SkipNow()
	/*
		env := initEVM(t)
		ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
		iscTest := env.deployISCTestContract(ethKey)

		// mint an NFT and send to chain
		env.soloChain.MustDepositIotasToL2(10*iscp.Mi, nil) // for gas
		issuerWallet, issuerAddress := env.solo.NewKeyPairWithFunds()
		metadata := []byte("foobar")
		nftInfo, err := env.solo.MintNFTL1(issuerWallet, issuerAddress, metadata)
		require.NoError(t, err)

		nft := new(isccontract.ISCNFT)
		iscTest.callFnExpectEvent([]ethCallOptions{{
			iota: iotaCallOptions{
				wallet: issuerWallet,
				before: func(cp *solo.CallParams) {
					cp.AddIotas(1 * iscp.Mi).
						WithNFT(&iscp.NFT{
							ID:       nftInfo.NFTID,
							Issuer:   issuerAddress,
							Metadata: metadata,
						}).
						WithAllowance(&iscp.Allowance{
							Assets: &iscp.FungibleTokens{Iotas: 1000},
							NFTs:   []iotago.NFTID{nftInfo.NFTID},
						}).
						WithGasBudget(1000000)
				},
			},
		}}, "AllowanceNFTEvent", &nft, "emitAllowanceNFTs")

		require.EqualValues(t, nftInfo.NFTID, nft.MustUnwrap().ID)
		require.True(t, issuerAddress.Equal(nft.MustUnwrap().Issuer))
		require.EqualValues(t, metadata, nft.MustUnwrap().Metadata)
	*/
}

func TestISCGetAllowanceAvailableNFTs(t *testing.T) {
	// TODO: how to send an NFT to an ethereum address on L2?
	t.SkipNow()
	/*
		env := initEVM(t)
		ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
		iscTest := env.deployISCTestContract(ethKey)

		// mint an NFT and send to chain
		env.soloChain.MustDepositIotasToL2(10*iscp.Mi, nil) // for gas
		issuerWallet, issuerAddress := env.solo.NewKeyPairWithFunds()
		metadata := []byte("foobar")
		nftInfo, err := env.solo.MintNFTL1(issuerWallet, issuerAddress, metadata)
		require.NoError(t, err)

		nft := new(isccontract.ISCNFT)
		iscTest.callFnExpectEvent([]ethCallOptions{{
			iota: iotaCallOptions{
				wallet: issuerWallet,
				before: func(cp *solo.CallParams) {
					cp.AddIotas(1 * iscp.Mi).
						WithNFT(&iscp.NFT{
							ID:       nftInfo.NFTID,
							Issuer:   issuerAddress,
							Metadata: metadata,
						}).
						WithAllowance(&iscp.Allowance{
							Assets: &iscp.FungibleTokens{Iotas: 1000},
							NFTs:   []iotago.NFTID{nftInfo.NFTID},
						}).
						WithGasBudget(1000000)
				},
			},
		}}, "AllowanceAvailableNFTEvent", &nft, "emitAllowanceAvailableNFTs")

		require.EqualValues(t, nftInfo.NFTID, nft.MustUnwrap().ID)
		require.True(t, issuerAddress.Equal(nft.MustUnwrap().Issuer))
		require.EqualValues(t, metadata, nft.MustUnwrap().Metadata)
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
		inccounter.FuncGetCounter.Name,
	)
	require.NoError(env.solo.T, err)
	require.EqualValues(t, 42, codec.MustDecodeInt64(r.MustGet(inccounter.VarCounter)))
}

func TestFibonacciContract(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	fibo := env.deployFibonacciContract(ethKey)
	require.EqualValues(t, 1, env.getBlockNumber())

	res, err := fibo.fib(7)
	require.NoError(t, err)
	t.Log("evm gas used:", res.evmReceipt.GasUsed)
	t.Log("iscp gas used:", res.iscpReceipt.GasBurned)
	t.Log("iscpc gas fee:", res.iscpReceipt.GasFeeCharged)
}
