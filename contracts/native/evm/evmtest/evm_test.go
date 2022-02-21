// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtest

import (
	"bytes"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/contracts/native/evm/evmchain"
	"github.com/iotaledger/wasp/contracts/native/evm/evmlight"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/evm/evmflavors"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/solo/solobench"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/stretchr/testify/require"
)

func TestDeploy(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		initEVMChain(t, evmFlavor)
	})
}

func TestFaucetBalance(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		evmChain := initEVMChain(t, evmFlavor)
		bal := evmChain.getBalance(evmChain.faucetAddress())
		require.Zero(t, evmChain.faucetSupply.Cmp(bal))
	})
}

func TestStorageContract(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		evmChain := initEVMChain(t, evmFlavor)

		// deploy solidity `storage` contract
		storage := evmChain.deployStorageContract(evmChain.faucetKey, 42)
		require.EqualValues(t, 1, evmChain.getBlockNumber())

		// call FuncCallView to call EVM contract's `retrieve` view, get 42
		require.EqualValues(t, 42, storage.retrieve())

		// call FuncSendTransaction with EVM tx that calls `store(43)`
		res, err := storage.store(43)
		require.NoError(t, err)
		require.Equal(t, types.ReceiptStatusSuccessful, res.evmReceipt.Status)
		require.EqualValues(t, 2, evmChain.getBlockNumber())

		// call `retrieve` view, get 43
		require.EqualValues(t, 43, storage.retrieve())
	})
}

func TestERC20Contract(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		evmChain := initEVMChain(t, evmFlavor)

		// deploy solidity `erc20` contract
		erc20 := evmChain.deployERC20Contract(evmChain.faucetKey, "TestCoin", "TEST")

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
	})
}

func TestGetCode(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		evmChain := initEVMChain(t, evmFlavor)
		erc20 := evmChain.deployERC20Contract(evmChain.faucetKey, "TestCoin", "TEST")

		// get contract bytecode from EVM emulator
		retrievedBytecode := evmChain.getCode(erc20.address)

		// ensure returned bytecode matches the expected runtime bytecode
		require.True(t, bytes.Equal(retrievedBytecode, evmtest.ERC20ContractRuntimeBytecode), "bytecode retrieved from the chain must match the deployed bytecode")
	})
}

func TestGasCharged(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		evmChain := initEVMChain(t, evmFlavor)
		storage := evmChain.deployStorageContract(evmChain.faucetKey, 42)

		iotaWallet, _ := evmChain.solo.NewKeyPairWithFunds()

		// call `store(999)` with enough gas
		res, err := storage.store(999, ethCallOptions{iota: iotaCallOptions{wallet: iotaWallet}})
		require.NoError(t, err)
		t.Log("evm gas used:", res.evmReceipt.GasUsed)
		t.Log("iscp gas used:", res.iscpReceipt.GasBurned)
		t.Log("iscp gas fee:", res.iscpReceipt.GasFeeCharged)
		require.Greater(t, res.evmReceipt.GasUsed, uint64(0))
		require.Greater(t, res.iscpReceipt.GasBurned, uint64(0))
		require.Greater(t, res.iscpReceipt.GasFeeCharged, uint64(0))
	})
}

func TestOwner(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		evmChain := initEVMChain(t, evmFlavor)

		// the default owner is correct
		owner := evmChain.getOwner()
		require.True(t, owner.Equals(evmChain.soloChain.OriginatorAgentID))

		// only the owner can call the setOwner endpoint
		user1Wallet, user1Address := evmChain.solo.NewKeyPairWithFunds()
		user1AgentID := iscp.NewAgentID(user1Address, 0)
		_, err := evmChain.soloChain.PostRequestSync(
			solo.NewCallParams(evmFlavor.Name, evm.FuncSetNextOwner.Name, evm.FieldNextEVMOwner, user1AgentID).
				AddAssetsIotas(100000).
				WithMaxAffordableGasBudget(),
			user1Wallet,
		)
		require.Error(t, err)

		// ensure owner didn't change after a failed call
		owner = evmChain.getOwner()
		require.True(t, owner.Equals(evmChain.soloChain.OriginatorAgentID))

		// current owner is able to set a new "next owner"
		_, err = evmChain.soloChain.PostRequestSync(
			solo.NewCallParams(evmFlavor.Name, evm.FuncSetNextOwner.Name, evm.FieldNextEVMOwner, user1AgentID).
				AddAssetsIotas(100000).
				WithMaxAffordableGasBudget(),
			evmChain.soloChain.OriginatorPrivateKey,
		)
		require.NoError(t, err)

		// check that the owner didn't change yet (new owner needs to claim ownership)
		owner = evmChain.getOwner()
		require.True(t, owner.Equals(evmChain.soloChain.OriginatorAgentID))

		// check no other user can claim ownership
		user2Wallet, _ := evmChain.solo.NewKeyPairWithFunds()

		_, err = evmChain.soloChain.PostRequestSync(
			solo.NewCallParams(evmFlavor.Name, evm.FuncClaimOwnership.Name).
				AddAssetsIotas(100000).
				WithMaxAffordableGasBudget(),
			user2Wallet,
		)
		require.Error(t, err)

		// owner still the same
		owner = evmChain.getOwner()
		require.True(t, owner.Equals(evmChain.soloChain.OriginatorAgentID))

		// claim ownership successfully
		_, err = evmChain.soloChain.PostRequestSync(
			solo.NewCallParams(evmFlavor.Name, evm.FuncClaimOwnership.Name).
				AddAssetsIotas(100000).
				WithMaxAffordableGasBudget(),
			user1Wallet,
		)
		require.NoError(t, err)
		owner = evmChain.getOwner()
		require.True(t, owner.Equals(user1AgentID))
	})
}

func TestGasRatio(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		evmChain := initEVMChain(t, evmFlavor)
		storage := evmChain.deployStorageContract(evmChain.faucetKey, 42)

		require.Equal(t, evm.DefaultGasRatio, evmChain.getGasRatio())

		res, err := storage.store(43)
		require.NoError(t, err)
		initialGasFee := res.iscpReceipt.GasFeeCharged

		// only the owner can call the setGasRatio endpoint
		newGasRatio := util.Ratio32{A: evm.DefaultGasRatio.A * 10, B: evm.DefaultGasRatio.B}
		newUserWallet, _ := evmChain.solo.NewKeyPairWithFunds()
		err = evmChain.setGasRatio(newGasRatio, iotaCallOptions{wallet: newUserWallet})
		require.True(t, iscp.VMErrorIs(err, vm.ErrUnauthorized))
		require.Equal(t, evm.DefaultGasRatio, evmChain.getGasRatio())

		// current owner is able to set a new gasRatio
		err = evmChain.setGasRatio(newGasRatio, iotaCallOptions{wallet: evmChain.soloChain.OriginatorPrivateKey})
		require.NoError(t, err)
		require.Equal(t, newGasRatio, evmChain.getGasRatio())

		// run an equivalent request and compare the gas fees
		res, err = storage.store(44)
		require.NoError(t, err)
		require.Greater(t, res.iscpReceipt.GasFeeCharged, initialGasFee)
	})
}

// tests that the gas limits are correctly enforced based on the iotas sent
func TestGasLimit(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		evmChain := initEVMChain(t, evmFlavor)
		storage := evmChain.deployStorageContract(evmChain.faucetKey, 42)

		// set a gas ratio such that evm gas cost in iotas is larger than dust cost
		err := evmChain.setGasRatio(util.Ratio32{A: 10, B: 1}, iotaCallOptions{wallet: evmChain.soloChain.OriginatorPrivateKey})
		require.NoError(t, err)

		// estimate gas by sending a valid tx
		iotaWallet1, _ := evmChain.solo.NewKeyPairWithFunds()
		result, err := storage.store(123, ethCallOptions{iota: iotaCallOptions{wallet: iotaWallet1}})
		require.NoError(t, err)
		gas := result.iscpReceipt.GasBurned
		fee := result.iscpReceipt.GasFeeCharged
		t.Logf("gas: %d, fee: %d", gas, fee)

		// send again with same gas limit but not enough iotas
		iotaWallet2, _ := evmChain.solo.NewKeyPairWithFunds()
		_, err = storage.store(124, ethCallOptions{iota: iotaCallOptions{
			wallet: iotaWallet2,
			before: func(req *solo.CallParams) {
				req.WithGasBudget(gas).AddAssetsIotas(fee * 9 / 10)
			},
		}})
		require.Error(t, err)
		require.Regexp(t, `\bgas\b`, err.Error())

		// send again with gas limit not enough for transaction
		iotaWallet3, _ := evmChain.solo.NewKeyPairWithFunds()
		_, err = storage.store(125, ethCallOptions{iota: iotaCallOptions{
			wallet: iotaWallet3,
			before: func(req *solo.CallParams) {
				req.WithGasBudget(gas / 2).AddAssetsIotas(fee)
			},
		}})
		require.Error(t, err)
		require.Regexp(t, `\bgas\b`, err.Error())
	})
}

// ensure the amount of iotas sent impacts the amount of gas used
func TestLoop(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		evmChain := initEVMChain(t, evmFlavor)
		loop := evmChain.deployLoopContract(evmChain.faucetKey)

		gasPerToken := evmChain.soloChain.GetGasFeePolicy().GasPerToken

		for _, evmGasLimit := range []uint64{100000, 200000} {
			iscGasBudget := evmGasLimit * 2
			iotasSent := iscGasBudget / gasPerToken
			iotaWallet, iotaAddress := evmChain.solo.NewKeyPairWithFunds()
			loop.loop(ethCallOptions{
				gasLimit: &evmGasLimit,
				iota: iotaCallOptions{
					wallet: iotaWallet,
					before: func(req *solo.CallParams) {
						req.WithGasBudget(iscGasBudget).AddAssetsIotas(iotasSent)
					},
				},
			})
			// gas fee is charged regardless of result
			require.LessOrEqual(t,
				evmChain.soloChain.L2Iotas(iscp.NewAgentID(iotaAddress, 0)),
				iotasSent-evmGasLimit/gasPerToken,
			)
		}
	})
}

func TestPrePaidFees(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		evmChain := initEVMChain(t, evmFlavor)
		storage := evmChain.deployStorageContract(evmChain.faucetKey, 42)

		iotaWallet, iotaAddress := evmChain.solo.NewKeyPairWithFunds()

		// test sending off-ledger request without depositing funds first
		txdata, _, _ := storage.buildEthTxData(nil, "store", uint32(999))
		offledgerRequest := evmChain.buildSoloRequest(evm.FuncSendTransaction.Name, evm.FieldTransactionData, txdata).
			WithMaxAffordableGasBudget()
		_, err := evmChain.soloChain.PostRequestOffLedger(offledgerRequest, iotaWallet)
		require.Error(t, err)
		require.EqualValues(t, 42, storage.retrieve())

		// deposit funds
		initialBalance := evmChain.solo.L1Iotas(iotaAddress)
		err = evmChain.soloChain.DepositIotasToL2(initialBalance/2, iotaWallet)
		require.NoError(t, err)

		// send offledger request again and check that is works
		_, err = evmChain.soloChain.PostRequestOffLedger(offledgerRequest, iotaWallet)
		require.NoError(t, err)
		require.EqualValues(t, 999, int(storage.retrieve()))
	})
}

func TestISCPContract(t *testing.T) {
	// deploy the evmlight contract, which starts an EVM chain and automatically
	// deploys the iscp.sol EVM contract at address 0x1074
	evmChain := initEVMChain(t, evmlight.Contract)

	// deploy the iscp-test.sol EVM contract
	iscpTest := evmChain.deployISCPTestContract(evmChain.faucetKey)

	// call the getChainId() view function of iscp-test.sol which in turn:
	//  calls the getChainId() view function of iscp.sol, which:
	//   returns the ChainID of the underlying ISCP chain
	chainID := iscpTest.getChainID()

	require.True(t, evmChain.soloChain.ChainID.Equals(chainID))
}

func TestISCPTriggerEvent(t *testing.T) {
	evmChain := initEVMChain(t, evmlight.Contract)
	iscpTest := evmChain.deployISCPTestContract(evmChain.faucetKey)

	// call the triggerEvent(string) function of iscp-test.sol which in turn:
	//  calls the iscpTriggerEvent(string) function of iscp.sol, which:
	//   executes a custom opcode, which:
	//    gets intercepted by the evmlight contract, which:
	//     triggers an ISCP event with the given string parameter
	res, err := iscpTest.triggerEvent("Hi from EVM!")
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, res.evmReceipt.Status)
	ev, err := evmChain.soloChain.GetEventsForBlock(evmChain.soloChain.GetLatestBlockInfo().BlockIndex)
	require.NoError(t, err)
	require.Len(t, ev, 1)
	require.Contains(t, ev[0], "Hi from EVM!")
}

func TestISCPEntropy(t *testing.T) {
	evmChain := initEVMChain(t, evmlight.Contract)
	iscpTest := evmChain.deployISCPTestContract(evmChain.faucetKey)

	// call the emitEntropy() function of iscp-test.sol which in turn:
	//  calls the iscpEntropy() function of iscp.sol, which:
	//   executes a custom opcode, which:
	//    gets intercepted by the evmlight contract, which:
	//     returns the entropy value from the sandbox
	//  emits an EVM event (aka log) with the entropy value
	res, err := iscpTest.emitEntropy()
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, res.evmReceipt.Status)
	require.Len(t, res.evmReceipt.Logs, 1)
	entropy := res.evmReceipt.Logs[0].Data
	require.Len(t, entropy, 32)
	require.NotEqualValues(t, entropy, make([]byte, 32))
}

func TestBlockTime(t *testing.T) {
	evmChain := initEVMChain(t, evmlight.Contract)

	// deposit funds to cover for dust, gas, etc
	_, err := evmChain.soloChain.PostRequestSync(
		solo.NewCallParams(accounts.Contract.Name, accounts.FuncTransferAllowanceTo.Name,
			accounts.ParamAgentID, iscp.NewAgentID(evmChain.soloChain.ChainID.AsAddress(), evmChain.evmFlavor.Hname()),
		).
			AddAssetsIotas(200000).
			AddAllowanceIotas(100000).
			WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	evmChain.setBlockTime(60)

	storage := evmChain.deployStorageContract(evmChain.faucetKey, 42)
	require.EqualValues(t, 42, storage.retrieve())
	require.EqualValues(t, 0, evmChain.getBlockNumber())

	res, err := storage.store(43)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, res.evmReceipt.Status)

	require.EqualValues(t, 43, storage.retrieve())
	require.EqualValues(t, 0, evmChain.getBlockNumber())

	// there is 1 timelocked request
	mempoolInfo := evmChain.soloChain.MempoolInfo()
	require.EqualValues(t, 1, mempoolInfo.InBufCounter-mempoolInfo.OutPoolCounter)

	// first block gets minted
	evmChain.solo.AdvanceClockBy(61*time.Second, 1)
	evmChain.soloChain.WaitUntil(func(mstats mempool.MempoolInfo) bool {
		return mstats.OutPoolCounter == mempoolInfo.InBufCounter
	})
	require.EqualValues(t, 1, evmChain.getBlockNumber())
	block := evmChain.getBlockByNumber(1)
	require.EqualValues(t, 2, len(block.Transactions()))

	// there is 1 timelocked request
	mempoolInfo = evmChain.soloChain.MempoolInfo()
	require.EqualValues(t, 1, mempoolInfo.InBufCounter-mempoolInfo.OutPoolCounter)

	// second (empty) block gets minted
	evmChain.solo.AdvanceClockBy(61*time.Second, 1)
	evmChain.soloChain.WaitUntil(func(mstats mempool.MempoolInfo) bool {
		return mstats.OutPoolCounter == mempoolInfo.InBufCounter
	})
	require.EqualValues(t, 2, evmChain.getBlockNumber())
	block = evmChain.getBlockByNumber(2)
	require.EqualValues(t, 0, len(block.Transactions()))
}

func initBenchmark(b *testing.B, evmFlavor *coreutil.ContractInfo) (*solo.Chain, []*solo.CallParams) {
	// setup: deploy the EVM chain
	log := testlogger.NewSilentLogger(b.Name(), true)
	env := solo.New(b, &solo.InitOptions{Log: log}).
		WithNativeContract(evmflavors.Processors[evmFlavor.Name])
	evmChain := initEVMChainWithSolo(b, evmFlavor, env)
	// setup: deploy the `storage` EVM contract
	storage := evmChain.deployStorageContract(evmChain.faucetKey, 42)

	// setup: prepare N requests that call FuncSendTransaction with an EVM tx
	// that calls `storage.store()`
	reqs := make([]*solo.CallParams, b.N)
	for i := 0; i < b.N; i++ {
		sender, err := crypto.GenerateKey() // send from a new address so that nonce is always 0
		require.NoError(b, err)

		txdata, _, _ := storage.buildEthTxData([]ethCallOptions{{sender: sender}}, "store", uint32(i))
		reqs[i] = storage.chain.buildSoloRequest(evm.FuncSendTransaction.Name, evm.FieldTransactionData, txdata)
		reqs[i].WithMaxAffordableGasBudget()
	}

	return evmChain.soloChain, reqs
}

// run benchmarks with: go test -benchmem -cpu=1 -run=' ' -bench='Bench.*'

func doBenchmark(b *testing.B, evmFlavor *coreutil.ContractInfo, f solobench.Func) {
	ch, reqs := initBenchmark(b, evmFlavor)
	f(b, ch, reqs, nil)
}

func BenchmarkEVMChainSync(b *testing.B) {
	doBenchmark(b, evmchain.Contract, solobench.RunBenchmarkSync)
}

func BenchmarkEVMLightSync(b *testing.B) {
	doBenchmark(b, evmlight.Contract, solobench.RunBenchmarkSync)
}

func BenchmarkEVMChainAsync(b *testing.B) {
	doBenchmark(b, evmchain.Contract, solobench.RunBenchmarkAsync)
}

func BenchmarkEVMLightAsync(b *testing.B) {
	doBenchmark(b, evmlight.Contract, solobench.RunBenchmarkAsync)
}
