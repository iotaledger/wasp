// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpctest

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"slices"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/evm/evmerrors"
	"github.com/iotaledger/wasp/v2/packages/evm/evmtest"
	"github.com/iotaledger/wasp/v2/packages/evm/evmutil"
	"github.com/iotaledger/wasp/v2/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/metrics"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm/iscmagic"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
)

const nativeToEthDigitsConversionRate = 1e9

type soloTestEnv struct {
	Env
	solo      *solo.Solo
	soloChain *solo.Chain
}

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func newSoloTestEnv(t testing.TB) *soloTestEnv {
	var log log.Logger
	if _, ok := t.(*testing.B); ok {
		log = testlogger.NewSilentLogger(t.Name(), true)
	}

	s := solo.New(t, &solo.InitOptions{
		Debug:           true,
		PrintStackTrace: true,
		Log:             log,
	})
	chainAdmin, _ := s.NewKeyPairWithFunds()
	chain, _ := s.NewChainExt(chainAdmin, 0, "chain1", evm.DefaultChainID, governance.DefaultBlockKeepAmount)

	accounts := jsonrpc.NewAccountManager(nil)
	rpcsrv, err := jsonrpc.NewServer(
		chain.EVM(),
		accounts,
		metrics.NewChainWebAPIMetricsProvider().CreateForChain(chain.ChainID),
		jsonrpc.ParametersDefault(),
	)
	require.NoError(t, err)
	t.Cleanup(rpcsrv.Stop)

	rawClient := rpc.DialInProc(rpcsrv)
	client := ethclient.NewClient(rawClient)
	t.Cleanup(client.Close)

	return &soloTestEnv{
		Env: Env{
			T:                     t,
			Client:                client,
			RawClient:             rawClient,
			ChainID:               evm.DefaultChainID,
			accountManager:        accounts,
			NewAccountWithL2Funds: chain.NewEthereumAccountWithL2Funds,
		},
		solo:      s,
		soloChain: chain,
	}
}

func TestRPCGetBalance(t *testing.T) {
	env := newSoloTestEnv(t)
	_, emptyAddress := solo.NewEthereumAccount()

	{
		_, err := env.Client.BalanceAtHash(context.Background(), emptyAddress, common.Hash{})
		require.ErrorContains(env.T, err, "not found")
	}

	require.Zero(t, env.Balance(emptyAddress).Uint64())

	initialBalance := coin.Value(1_666_666_666) // enought for transfer + gas, but also fits single coin object allocated from faucet
	wallet, nonEmptyAddress := env.soloChain.NewEthereumAccountWithL2Funds(initialBalance)
	initialBalanceEth := env.Balance(nonEmptyAddress)
	initialBalanceNative := uint64(env.soloChain.L2BaseTokens(isc.NewEthereumAddressAgentID(nonEmptyAddress)))
	require.Equal(t, initialBalance.Uint64()*nativeToEthDigitsConversionRate, initialBalanceEth.Uint64())
	require.Equal(t, initialBalanceNative*nativeToEthDigitsConversionRate, initialBalanceEth.Uint64())

	// 18 decimals
	toSend := new(big.Int).SetUint64(1_111_111_111_111_111_111) // use all 18 decimals
	tx := types.NewTransaction(0, emptyAddress, toSend, uint64(100_000), env.MustGetGasPrice(), []byte{})
	signedTx, err := types.SignTx(tx, env.Signer(), wallet)
	require.NoError(t, err)

	receipt := env.mustSendTransactionAndWait(signedTx)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)

	fee := getGasFeeCharged(receipt)
	exptectedBalance := new(big.Int).Sub(initialBalanceEth, toSend)
	exptectedBalance = new(big.Int).Sub(exptectedBalance, fee)
	require.Equal(t, exptectedBalance, env.Balance(nonEmptyAddress))
	require.Equal(t, toSend, env.Balance(emptyAddress))
}

func getGasFeeCharged(receipt *types.Receipt) *big.Int {
	fee := new(big.Int).Mul(receipt.EffectiveGasPrice, new(big.Int).SetUint64(receipt.GasUsed))

	fee, leftover := new(big.Int).QuoRem(fee, big.NewInt(nativeToEthDigitsConversionRate), fee)
	if leftover.Sign() != 0 {
		fee.Add(fee, big.NewInt(1))
	}

	fee = new(big.Int).Mul(fee, big.NewInt(nativeToEthDigitsConversionRate))

	return fee
}

func TestRPCGetCode(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, creatorAddress := env.soloChain.NewEthereumAccountWithL2Funds()

	// account address
	require.Empty(t, env.Code(creatorAddress))

	// contract address
	{
		_, contractAddress, _ := env.deployStorageContract(creator)
		require.NotEmpty(t, env.Code(contractAddress))
	}
}

func TestRPCGetStorage(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	_, contractAddress, contractABI := env.deployStorageContract(creator)

	// first static variable in contract (uint32 n) has slot 0. See:
	// https://docs.soliditylang.org/en/v0.6.6/miscellaneous.html#layout-of-state-variables-in-storage
	slot := common.Hash{}
	ret := env.Storage(contractAddress, slot)

	var v uint32
	err := contractABI.UnpackIntoInterface(&v, "retrieve", ret)
	require.NoError(t, err)
	require.Equal(t, uint32(42), v)
}

func TestRPCBlockNumber(t *testing.T) {
	env := newSoloTestEnv(t)
	require.EqualValues(t, 0, env.BlockNumber())
	creator, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	require.EqualValues(t, 1, env.BlockNumber()) // EVM block number is incremented along with ISC block index
	env.deployStorageContract(creator)
	require.EqualValues(t, 2, env.BlockNumber())
}

func TestRPCGetTransactionCount(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, creatorAddress := env.soloChain.NewEthereumAccountWithL2Funds()
	require.EqualValues(t, 0, env.NonceAt(creatorAddress))
	env.deployStorageContract(creator)
	require.EqualValues(t, 1, env.NonceAt(creatorAddress))
}

func TestRPCGetBlockByNumber(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	env.deployStorageContract(creator)
	require.EqualValues(t, 0, env.BlockByNumber(big.NewInt(0)).Number().Uint64())
	require.EqualValues(t, 1, env.BlockByNumber(big.NewInt(1)).Number().Uint64())
	require.EqualValues(t, 2, env.BlockByNumber(big.NewInt(2)).Number().Uint64())
	// latest:
	require.EqualValues(t, 2, env.BlockByNumber(nil).Number().Uint64())
}

func TestRPCGetBlockByHash(t *testing.T) {
	env := newSoloTestEnv(t)
	require.Nil(t, env.BlockByHash(common.Hash{}))
	creator, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	require.EqualValues(t, 0, env.BlockByHash(env.BlockByNumber(big.NewInt(0)).Hash()).Number().Uint64())
	env.deployStorageContract(creator)
	require.EqualValues(t, 1, env.BlockByHash(env.BlockByNumber(big.NewInt(1)).Hash()).Number().Uint64())
}

func TestRPCGetTransactionByHash(t *testing.T) {
	env := newSoloTestEnv(t)
	require.Nil(t, env.TransactionByHash(common.Hash{}))
	creator, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	env.deployStorageContract(creator)
	block := env.BlockByNumber(new(big.Int).SetUint64(env.BlockNumber()))
	tx := env.TransactionByHash(block.Transactions()[0].Hash())
	require.Equal(t, block.Transactions()[0].Hash(), tx.Hash())
}

func TestRPCGetTransactionByBlockHashAndIndex(t *testing.T) {
	env := newSoloTestEnv(t)
	require.Nil(t, env.TransactionByBlockHashAndIndex(common.Hash{}, 0))
	creator, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	env.deployStorageContract(creator)
	block := env.BlockByNumber(new(big.Int).SetUint64(env.BlockNumber()))
	tx := env.TransactionByBlockHashAndIndex(block.Hash(), 0)
	require.Equal(t, block.Transactions()[0].Hash(), tx.Hash())
}

func TestRPCGetUncleByBlockHashAndIndex(t *testing.T) {
	env := newSoloTestEnv(t)
	require.Nil(t, env.UncleByBlockHashAndIndex(common.Hash{}, 0))
	creator, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	env.deployStorageContract(creator)
	block := env.BlockByNumber(new(big.Int).SetUint64(env.BlockNumber()))
	require.Nil(t, env.UncleByBlockHashAndIndex(block.Hash(), 0))
}

func TestRPCGetTransactionByBlockNumberAndIndex(t *testing.T) {
	env := newSoloTestEnv(t)
	ret, err := env.TransactionByBlockNumberAndIndex(big.NewInt(3), 0)
	require.NoError(t, err)
	require.Nil(t, ret)
	creator, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	env.deployStorageContract(creator)
	block := env.BlockByNumber(new(big.Int).SetUint64(env.BlockNumber()))
	tx, err := env.TransactionByBlockNumberAndIndex(block.Number(), 0)
	require.NoError(t, err)
	require.EqualValues(t, block.Hash(), *tx.BlockHash)
	require.EqualValues(t, 0, *tx.TransactionIndex)
}

func TestRPCGetUncleByBlockNumberAndIndex(t *testing.T) {
	env := newSoloTestEnv(t)
	require.Nil(t, env.UncleByBlockNumberAndIndex(big.NewInt(3), 0))
	creator, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	env.deployStorageContract(creator)
	block := env.BlockByNumber(new(big.Int).SetUint64(env.BlockNumber()))
	require.Nil(t, env.UncleByBlockNumberAndIndex(block.Number(), 0))
}

func TestRPCGetTransactionCountByHash(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	env.deployStorageContract(creator)
	block := env.BlockByNumber(new(big.Int).SetUint64(env.BlockNumber()))
	require.Positive(t, len(block.Transactions()))
	require.EqualValues(t, len(block.Transactions()), env.BlockTransactionCountByHash(block.Hash()))
	require.EqualValues(t, 0, env.BlockTransactionCountByHash(common.Hash{}))
}

func TestRPCGetUncleCountByBlockHash(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	env.deployStorageContract(creator)
	block := env.BlockByNumber(new(big.Int).SetUint64(env.BlockNumber()))
	require.Zero(t, len(block.Uncles()))
	require.EqualValues(t, len(block.Uncles()), env.UncleCountByBlockHash(block.Hash()))
	require.EqualValues(t, 0, env.UncleCountByBlockHash(common.Hash{}))
}

func TestRPCGetTransactionCountByNumber(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	env.deployStorageContract(creator)
	block := env.BlockByNumber(new(big.Int).SetUint64(env.BlockNumber()))
	require.Positive(t, len(block.Transactions()))
	require.EqualValues(t, len(block.Transactions()), env.BlockTransactionCountByNumber())
}

func TestRPCGetUncleCountByBlockNumber(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	env.deployStorageContract(creator)
	block := env.BlockByNumber(new(big.Int).SetUint64(env.BlockNumber()))
	require.Zero(t, len(block.Uncles()))
	require.EqualValues(t, len(block.Uncles()), env.UncleCountByBlockNumber(big.NewInt(1)))
}

func TestRPCAccounts(t *testing.T) {
	env := newSoloTestEnv(t)
	k1, _ := solo.NewEthereumAccount()
	k2, _ := solo.NewEthereumAccount()
	env.accountManager.Add(k1)
	env.accountManager.Add(k2)
	accounts := env.Accounts()
	require.EqualValues(t, 2, len(accounts))
}

func TestRPCSign(t *testing.T) {
	k1, a1 := solo.NewEthereumAccount()
	env := newSoloTestEnv(t)
	env.accountManager.Add(k1)
	signed := env.Sign(a1, []byte("hello"))
	require.NotEmpty(t, signed)
}

const additionalGasBurnedByVM = 10

func TestRPCSignTransaction(t *testing.T) {
	_, to := solo.NewEthereumAccount()
	env := newSoloTestEnv(t)
	ethKey, ethAddr := env.soloChain.NewEthereumAccountWithL2Funds()
	env.accountManager.Add(ethKey)

	gas := hexutil.Uint64(params.TxGas) + additionalGasBurnedByVM
	nonce := hexutil.Uint64(env.NonceAt(ethAddr))
	signed := env.SignTransaction(&jsonrpc.SendTxArgs{
		From:     ethAddr,
		To:       &to,
		Gas:      &gas,
		GasPrice: (*hexutil.Big)(big.NewInt(1000)),
		Value:    (*hexutil.Big)(big.NewInt(42)),
		Nonce:    &nonce,
	})
	require.NotEmpty(t, signed)

	// assert that the tx is correctly signed
	{
		var decodedTx types.Transaction
		err := decodedTx.UnmarshalBinary(signed)
		require.NoError(t, err)
		sender, err := evmutil.GetSender(&decodedTx)
		require.NoError(t, err)
		require.Equal(t, ethAddr, sender)
	}
}

func TestRPCSendTransaction(t *testing.T) {
	env := newSoloTestEnv(t)
	ethKey, ethAddr := env.soloChain.NewEthereumAccountWithL2Funds()
	env.accountManager.Add(ethKey)

	gas := hexutil.Uint64(100_000)
	nonce := hexutil.Uint64(env.NonceAt(ethAddr))
	data := common.Hex2Bytes("600180600b6000396000f3") // some contract bytecode
	txHash := env.MustSendTransaction(&jsonrpc.SendTxArgs{
		From:     ethAddr,
		Gas:      &gas,
		GasPrice: (*hexutil.Big)(env.MustGetGasPrice()),
		Nonce:    &nonce,
		Data:     (*hexutil.Bytes)(&data),
	})
	require.NotEqualValues(t, common.Hash{}, txHash)
}

func TestRPCGetTxReceipt(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, _ := env.soloChain.NewEthereumAccountWithL2Funds()

	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)
	tx, receipt, contractAddress := env.DeployEVMContract(creator, contractABI, evmtest.StorageContractBytecode, uint32(42))

	require.EqualValues(t, types.LegacyTxType, receipt.Type)
	require.EqualValues(t, types.ReceiptStatusSuccessful, receipt.Status)
	require.NotZero(t, receipt.CumulativeGasUsed)
	require.EqualValues(t, types.Bloom{}, receipt.Bloom)
	require.EqualValues(t, 0, len(receipt.Logs))

	require.EqualValues(t, tx.Hash(), receipt.TxHash)
	require.EqualValues(t, contractAddress, receipt.ContractAddress)
	require.NotZero(t, receipt.GasUsed)

	require.EqualValues(t, big.NewInt(2), receipt.BlockNumber)
	require.EqualValues(t, env.BlockByNumber(big.NewInt(2)).Hash(), receipt.BlockHash)
	require.EqualValues(t, 0, receipt.TransactionIndex)

	expectedGasPrice := env.soloChain.GetGasFeePolicy().DefaultGasPriceFullDecimals(parameters.BaseTokenDecimals)
	require.EqualValues(t, expectedGasPrice, receipt.EffectiveGasPrice)
}

func TestRPCGetTxReceiptMissing(t *testing.T) {
	env := newSoloTestEnv(t)

	_, err := env.TxReceipt(common.Hash{})
	require.Error(t, err)
	require.Equal(t, "not found", err.Error())
}

func TestRPCCall(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, creatorAddress := env.soloChain.NewEthereumAccountWithL2Funds()
	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)
	_, _, contractAddress := env.DeployEVMContract(creator, contractABI, evmtest.StorageContractBytecode, uint32(1))

	callStore := func(n uint32) {
		receipt := env.mustSendTransactionAndWait(types.MustSignNewTx(
			creator,
			types.NewEIP155Signer(big.NewInt(int64(env.ChainID))),
			&types.LegacyTx{
				Nonce:    env.NonceAt(creatorAddress),
				To:       &contractAddress,
				Gas:      100_000,
				GasPrice: big.NewInt(10000000000),
				Data:     lo.Must(contractABI.Pack("store", n)),
			},
		))
		require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	}

	callRetrieve := func(bn *big.Int, expectError string) uint32 {
		ret, err := env.Client.CallContract(context.Background(), ethereum.CallMsg{
			From: creatorAddress,
			To:   &contractAddress,
			Data: lo.Must(contractABI.Pack("retrieve")),
			Gas:  100_000,
		}, bn)
		if expectError != "" {
			require.ErrorContains(t, err, expectError)
			return 0
		}
		require.NoError(t, err)
		var v uint32
		err = contractABI.UnpackIntoInterface(&v, "retrieve", ret)
		require.NoError(t, err)
		return v
	}

	blockNumber := env.soloChain.EVM().BlockNumber().Int64()

	v := callRetrieve(nil, "")
	require.Equal(t, uint32(1), v)

	v = callRetrieve(big.NewInt(blockNumber), "")
	require.Equal(t, uint32(1), v)

	_ = callRetrieve(big.NewInt(blockNumber+1), "not found")

	callStore(2)

	v = callRetrieve(nil, "")
	require.Equal(t, uint32(2), v)

	v = callRetrieve(big.NewInt(blockNumber), "")
	require.Equal(t, uint32(1), v)

	v = callRetrieve(big.NewInt(blockNumber+1), "")
	require.Equal(t, uint32(2), v)

	_ = callRetrieve(big.NewInt(blockNumber+2), "not found")

	callStore(3)

	v = callRetrieve(nil, "")
	require.Equal(t, uint32(3), v)

	v = callRetrieve(big.NewInt(blockNumber), "")
	require.Equal(t, uint32(1), v)

	v = callRetrieve(big.NewInt(blockNumber+1), "")
	require.Equal(t, uint32(2), v)

	v = callRetrieve(big.NewInt(blockNumber+2), "")
	require.Equal(t, uint32(3), v)

	_ = callRetrieve(big.NewInt(blockNumber+3), "not found")
}

func TestRPCCallNonView(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, creatorAddress := env.soloChain.NewEthereumAccountWithL2Funds()
	contractABI, err := abi.JSON(strings.NewReader(evmtest.ISCTestContractABI))
	require.NoError(t, err)
	_, _, contractAddress := env.DeployEVMContract(creator, contractABI, evmtest.ISCTestContractBytecode)

	callArguments, err := contractABI.Pack("triggerEvent", "hello")
	require.NoError(t, err)

	_, err = env.Client.CallContract(context.Background(), ethereum.CallMsg{
		From: creatorAddress,
		To:   &contractAddress,
		Data: callArguments,
		Gas:  100_000,
	}, nil)
	require.NoError(t, err)
}

func TestRPCAccessHistoricalState(t *testing.T) {
	env := newSoloTestEnv(t)
	env.TestRPCAccessHistoricalState()
}

func TestRPCGetLogs(t *testing.T) {
	env := newSoloTestEnv(t)
	env.TestRPCGetLogs()
}

func TestRPCLogIndex(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, creatorAddress := env.soloChain.NewEthereumAccountWithL2Funds()
	contractABI, err := abi.JSON(strings.NewReader(evmtest.ISCTestContractABI))
	require.NoError(t, err)
	_, _, contractAddress := env.DeployEVMContract(creator, contractABI, evmtest.ISCTestContractBytecode)

	callArguments, err := contractABI.Pack("loopWithGasLeft")
	require.NoError(t, err)

	value := big.NewInt(0)
	gas := uint64(100_000)
	tx, err := types.SignTx(
		types.NewTransaction(env.NonceAt(creatorAddress), contractAddress, value, gas, env.MustGetGasPrice(), callArguments),
		env.Signer(),
		creator,
	)
	require.NoError(t, err)
	receipt := env.mustSendTransactionAndWait(tx)

	logs, err := env.Client.FilterLogs(context.Background(), ethereum.FilterQuery{
		FromBlock: receipt.BlockNumber,
		ToBlock:   receipt.BlockNumber,
	})
	require.NoError(t, err)
	require.Greater(t, len(logs), 2)
	require.EqualValues(t, 1, logs[1].Index)
}

func TestRPCEthChainID(t *testing.T) {
	env := newSoloTestEnv(t)
	var chainID hexutil.Uint
	err := env.RawClient.Call(&chainID, "eth_chainId")
	require.NoError(t, err)
	require.EqualValues(t, evm.DefaultChainID)
}

func TestRPCTxRejectedIfNotEnoughFunds(t *testing.T) {
	creator, creatorAddress := solo.NewEthereumAccount()

	env := newSoloTestEnv(t)
	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)
	nonce := env.NonceAt(creatorAddress)
	constructorArguments, err := contractABI.Pack("", uint32(42))
	require.NoError(t, err)
	data := concatenate(evmtest.StorageContractBytecode, constructorArguments)
	value := big.NewInt(0)
	gasLimit := uint64(10_000)
	tx, err := types.SignTx(
		types.NewContractCreation(nonce, value, gasLimit, env.MustGetGasPrice(), data),
		env.Signer(),
		creator,
	)
	require.NoError(t, err)

	// the tx is rejected before posting to the wasp node
	err = env.Client.SendTransaction(context.Background(), tx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "sender doesn't have enough L2 funds to cover tx gas budget")
}

func TestRPCCustomError(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, creatorAddress := env.soloChain.NewEthereumAccountWithL2Funds()
	contractABI, err := abi.JSON(strings.NewReader(evmtest.ISCTestContractABI))
	require.NoError(t, err)
	_, _, contractAddress := env.DeployEVMContract(creator, contractABI, evmtest.ISCTestContractBytecode)

	callArguments, err := contractABI.Pack("revertWithCustomError")
	require.NoError(t, err)

	_, err = env.Client.CallContract(context.Background(), ethereum.CallMsg{
		From: creatorAddress,
		To:   &contractAddress,
		Data: callArguments,
	}, nil)
	require.ErrorContains(t, err, "execution reverted")

	dataErr, ok := err.(rpc.DataError)
	require.True(t, ok)

	revertData, err := hexutil.Decode(dataErr.ErrorData().(string))
	require.NoError(t, err)

	args, err := evmerrors.UnpackCustomError(revertData, contractABI.Errors["CustomError"])
	require.NoError(t, err)

	require.Len(t, args, 1)
	require.EqualValues(t, 42, args[0])
}

func TestRPCTraceTx(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, creatorAddress := env.soloChain.NewEthereumAccountWithL2Funds()
	contractABI, err := abi.JSON(strings.NewReader(evmtest.ISCTestContractABI))
	require.NoError(t, err)
	_, _, contractAddress := env.DeployEVMContract(creator, contractABI, evmtest.ISCTestContractBytecode)

	// make it so that 2 requests are included in the same block

	tx1 := types.MustSignNewTx(creator, types.NewEIP155Signer(big.NewInt(int64(env.ChainID))),
		&types.LegacyTx{
			Nonce:    env.NonceAt(creatorAddress),
			To:       &contractAddress,
			Value:    big.NewInt(123),
			Gas:      100000,
			GasPrice: big.NewInt(10000000000),
			Data:     lo.Must(contractABI.Pack("sendTo", common.Address{0x1}, big.NewInt(1))),
		})

	tx2 := types.MustSignNewTx(creator, types.NewEIP155Signer(big.NewInt(int64(env.ChainID))),
		&types.LegacyTx{
			Nonce:    env.NonceAt(creatorAddress) + 1,
			To:       &contractAddress,
			Value:    big.NewInt(123),
			Gas:      100000,
			GasPrice: big.NewInt(10000000000),
			Data:     lo.Must(contractABI.Pack("sendTo", common.Address{0x2}, big.NewInt(1))),
		})

	req1 := lo.Must(isc.NewEVMOffLedgerTxRequest(env.soloChain.ChainID, tx1))
	req2 := lo.Must(isc.NewEVMOffLedgerTxRequest(env.soloChain.ChainID, tx2))

	env.soloChain.RunRequestsSync([]isc.Request{req1, req2})

	bi := env.soloChain.GetLatestBlockInfo()
	require.EqualValues(t, 2, bi.NumSuccessfulRequests)

	t.Run("callTracer", func(t *testing.T) {
		// assert each tx can be individually traced
		trace1, err := env.traceTransactionWithCallTracer(tx1.Hash())
		require.NoError(t, err)
		_, err = env.traceTransactionWithCallTracer(tx2.Hash())
		require.NoError(t, err)

		require.Equal(t, creatorAddress, trace1.From)
		require.Equal(t, contractAddress, *trace1.To)
		require.Equal(t, "0x7b", trace1.Value.String())
		expectedInput, err := contractABI.Pack("sendTo", common.Address{0x1}, big.NewInt(1))
		require.NoError(t, err)
		require.Equal(t, expectedInput, []byte(trace1.Input))
		require.Empty(t, trace1.Error)
		require.Empty(t, trace1.RevertReason)
		require.Contains(t, trace1.Gas.String(), "0x")
		require.Contains(t, trace1.GasUsed.String(), "0x")

		require.Len(t, trace1.Calls, 1)
		trace2 := trace1.Calls[0]
		require.Equal(t, contractAddress, trace2.From)
		require.Equal(t, common.Address{0x1}, *trace2.To)
		require.Equal(t, "0x1", trace2.Value.String())
		require.Empty(t, trace2.Input)
		require.Empty(t, trace2.Error)
		require.Empty(t, trace2.RevertReason)
		require.Contains(t, trace2.Gas.String(), "0x")
		require.Contains(t, trace2.GasUsed.String(), "0x")
	})

	t.Run("prestate", func(t *testing.T) {
		accountMap, err := env.traceTransactionWithPrestate(tx1.Hash())
		// t.Logf("%s", lo.Must(json.MarshalIndent(accountMap, "", "  ")))
		require.NoError(t, err)
		require.NotEmpty(t, accountMap)

		diff, err := env.traceTransactionWithPrestateDiff(tx1.Hash())
		// t.Logf("%s", lo.Must(json.MarshalIndent(diff, "", "  ")))
		require.NoError(t, err)
		require.NotEmpty(t, diff.Pre)
		require.NotEmpty(t, diff.Post)
	})
}

func TestRPCTraceFailedTx(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, creatorAddress := env.soloChain.NewEthereumAccountWithL2Funds()
	creatorL2Balance := env.Balance(creatorAddress)
	contractABI, err := abi.JSON(strings.NewReader(evmtest.ISCTestContractABI))
	require.NoError(t, err)
	_, _, contractAddress := env.DeployEVMContract(creator, contractABI, evmtest.ISCTestContractBytecode)

	tx := types.MustSignNewTx(creator, types.NewEIP155Signer(big.NewInt(int64(env.ChainID))),
		&types.LegacyTx{
			Nonce:    env.NonceAt(creatorAddress),
			To:       &contractAddress,
			Value:    creatorL2Balance,
			Gas:      100000000000000,
			GasPrice: big.NewInt(10000000000),
			Data:     lo.Must(contractABI.Pack("sendTo", common.Address{0x1}, big.NewInt(1))),
		})

	_, err = env.SendTransactionAndWait(tx)
	require.ErrorContains(t, err, "insufficient funds for gas * price + value")

	bi := env.soloChain.GetLatestBlockInfo()
	require.EqualValues(t, 0, bi.NumSuccessfulRequests)

	t.Run("callTracer", func(t *testing.T) {
		_, err := env.traceTransactionWithCallTracer(tx.Hash())
		require.ErrorContains(t, err, "expected exactly one top-level call")
	})

	t.Run("prestate", func(t *testing.T) {
		accountMap, err := env.traceTransactionWithPrestate(tx.Hash())
		// t.Logf("%s", lo.Must(json.MarshalIndent(accountMap, "", "  ")))
		require.NoError(t, err)
		require.NotEmpty(t, accountMap)

		diff, err := env.traceTransactionWithPrestateDiff(tx.Hash())
		// t.Logf("%s", lo.Must(json.MarshalIndent(diff, "", "  ")))
		require.NoError(t, err)
		require.NotEmpty(t, diff.Pre)
		require.Empty(t, diff.Post)
	})
}

// Transfer calls produce "fake" Transactions to simulate EVM behavior.
// They are not real in the sense of being persisted to the blockchain, therefore requires additional checks.
func TestRPCTraceEVMDeposit(t *testing.T) {
	env := newSoloTestEnv(t)
	wallet, _ := env.solo.NewKeyPairWithFunds()
	_, evmAddr := env.soloChain.NewEthereumAccountWithL2Funds()

	err := env.soloChain.TransferAllowanceTo(
		isc.NewAssets(1000),
		isc.NewEthereumAddressAgentID(evmAddr),
		wallet)

	block := env.BlockByNumber(nil)
	require.NoError(t, err)
	txs := block.Transactions()
	tx := txs[0]

	require.Equal(t, evmAddr, *tx.To())

	rc, err := env.TxReceipt(txs[0].Hash())
	require.NoError(t, err)
	require.EqualValues(t, types.ReceiptStatusSuccessful, rc.Status)

	t.Run("callTracer_tx", func(t *testing.T) {
		var trace jsonrpc.CallFrame
		trace, err = env.traceTransactionWithCallTracer(tx.Hash())
		require.NoError(t, err)
		require.Equal(t, evmAddr.String(), trace.To.String())
		require.Equal(t, hexutil.EncodeUint64(isc.NewAssets(1000).BaseTokens().Uint64()*nativeToEthDigitsConversionRate), trace.Value.String())
	})

	t.Run("prestateTracer_tx", func(t *testing.T) {
		var prestate jsonrpc.PrestateAccountMap
		prestate, err = env.traceTransactionWithPrestate(tx.Hash())
		require.NoError(t, err)
		require.Empty(t, prestate)
	})

	t.Run("prestateTracerDiff_tx", func(t *testing.T) {
		var prestateDiff jsonrpc.PrestateDiffResult
		prestateDiff, err = env.traceTransactionWithPrestateDiff(tx.Hash())
		require.NoError(t, err)
		require.Empty(t, prestateDiff.Pre)
		require.Empty(t, prestateDiff.Post)
	})

	t.Run("callTracer_block", func(t *testing.T) {
		callTracer := "callTracer"
		var res1 json.RawMessage
		// we have to use the raw client, because the normal client does not support debug methods
		err = env.RawClient.CallContext(
			context.Background(),
			&res1,
			"debug_traceBlockByNumber",
			hexutil.Uint64(env.BlockNumber()).String(),
			tracers.TraceConfig{Tracer: &callTracer},
		)
		require.NoError(t, err)

		traces := make([]jsonrpc.TxTraceResult, 0)
		err = json.Unmarshal(res1, &traces)
		require.NoError(t, err)
		require.Len(t, traces, 1)
		require.Equal(t, tx.Hash(), traces[0].TxHash)

		cs := jsonrpc.CallFrame{}
		err = json.Unmarshal(traces[0].Result, &cs)
		require.NoError(t, err)
		require.Equal(t, evmAddr.String(), cs.To.String())
		require.Equal(t, hexutil.EncodeUint64(isc.NewAssets(1000).BaseTokens().Uint64()*nativeToEthDigitsConversionRate), cs.Value.String())
	})

	t.Run("prestateTracer_block", func(t *testing.T) {
		tracer := "prestateTracer"
		var res1 json.RawMessage
		// we have to use the raw client, because the normal client does not support debug methods
		err = env.RawClient.CallContext(
			context.Background(),
			&res1,
			"debug_traceBlockByNumber",
			hexutil.Uint64(env.BlockNumber()).String(),
			tracers.TraceConfig{Tracer: &tracer},
		)
		require.NoError(t, err)

		traces := make([]jsonrpc.TxTraceResult, 0)
		err = json.Unmarshal(res1, &traces)
		require.NoError(t, err)
		require.Len(t, traces, 1)
		require.Equal(t, tx.Hash(), traces[0].TxHash)

		prestate := jsonrpc.PrestateAccountMap{}
		err = json.Unmarshal(traces[0].Result, &prestate)
		require.NoError(t, err)
		require.Empty(t, prestate)
	})
}

func addNRequests(n int, env *soloTestEnv, creator *ecdsa.PrivateKey, creatorAddress common.Address, contractABI abi.ABI, contractAddress common.Address) {
	rqs := make([]isc.Request, 0, n)
	for i := range n {
		tx1 := types.MustSignNewTx(creator, types.NewEIP155Signer(big.NewInt(int64(env.ChainID))),
			&types.LegacyTx{
				Nonce:    env.NonceAt(creatorAddress) + uint64(i),
				To:       &contractAddress,
				Value:    big.NewInt(123),
				Gas:      100000,
				GasPrice: big.NewInt(10000000000),
				Data:     lo.Must(contractABI.Pack("sendTo", common.Address{0x1}, big.NewInt(2))),
			})

		req1 := lo.Must(isc.NewEVMOffLedgerTxRequest(env.soloChain.ChainID, tx1))
		rqs = append(rqs, req1)
	}

	env.soloChain.RunRequestsSync(rqs)
}

func TestRPCTraceBlockForLargeN(t *testing.T) {
	n := 400
	env := newSoloTestEnv(t)
	creator, creatorAddress := env.soloChain.NewEthereumAccountWithL2Funds()
	contractABI, err := abi.JSON(strings.NewReader(evmtest.ISCTestContractABI))
	require.NoError(t, err)
	_, _, contractAddress := env.DeployEVMContract(creator, contractABI, evmtest.ISCTestContractBytecode)

	addNRequests(n, env, creator, creatorAddress, contractABI, contractAddress)

	bi := env.soloChain.GetLatestBlockInfo()
	require.EqualValues(t, n, bi.NumSuccessfulRequests)

	var res1 json.RawMessage
	err = env.RawClient.CallContext(
		context.Background(),
		&res1,
		"trace_block",
		rpc.BlockNumber(env.BlockNumber()),
	)
	require.NoError(t, err)

	traces := make([]map[string]any, 0)
	err = json.Unmarshal(res1, &traces)
	require.NoError(t, err)

	// 400 requests, 2 traces per request
	require.Len(t, traces, 800)
}

func TestRPCDebugTraceBlockForLargeN(t *testing.T) {
	n := 400
	env := newSoloTestEnv(t)
	creator, creatorAddress := env.soloChain.NewEthereumAccountWithL2Funds()
	contractABI, err := abi.JSON(strings.NewReader(evmtest.ISCTestContractABI))
	require.NoError(t, err)
	_, _, contractAddress := env.DeployEVMContract(creator, contractABI, evmtest.ISCTestContractBytecode)

	addNRequests(n, env, creator, creatorAddress, contractABI, contractAddress)

	bi := env.soloChain.GetLatestBlockInfo()
	require.EqualValues(t, n, bi.NumSuccessfulRequests)

	callTracer := "callTracer"
	var res1 json.RawMessage
	// we have to use the raw client, because the normal client does not support debug methods
	err = env.RawClient.CallContext(
		context.Background(),
		&res1,
		"debug_traceBlockByNumber",
		hexutil.Uint64(env.BlockNumber()).String(),
		tracers.TraceConfig{Tracer: &callTracer},
	)
	require.NoError(t, err)

	traces := make([]jsonrpc.TxTraceResult, 0)
	err = json.Unmarshal(res1, &traces)
	require.NoError(t, err)

	// 400 requests, 1 trace per request
	require.Len(t, traces, 400)
}

func TestRPCTraceBlock(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, creatorAddress := env.soloChain.NewEthereumAccountWithL2Funds()
	creator2, creatorAddress2 := env.soloChain.NewEthereumAccountWithL2Funds()
	contractABI, err := abi.JSON(strings.NewReader(evmtest.ISCTestContractABI))
	require.NoError(t, err)
	_, _, contractAddress := env.DeployEVMContract(creator, contractABI, evmtest.ISCTestContractBytecode)

	// make it so that 2 requests are included in the same block
	tx1 := types.MustSignNewTx(creator, types.NewEIP155Signer(big.NewInt(int64(env.ChainID))),
		&types.LegacyTx{
			Nonce:    env.NonceAt(creatorAddress),
			To:       &contractAddress,
			Value:    big.NewInt(123),
			Gas:      100000,
			GasPrice: big.NewInt(10000000000),
			Data:     lo.Must(contractABI.Pack("sendTo", common.Address{0x1}, big.NewInt(2))),
		})

	tx2 := types.MustSignNewTx(creator2, types.NewEIP155Signer(big.NewInt(int64(env.ChainID))),
		&types.LegacyTx{
			Nonce:    env.NonceAt(creatorAddress2),
			To:       &contractAddress,
			Value:    big.NewInt(321),
			Gas:      100000,
			GasPrice: big.NewInt(10000000000),
			Data:     lo.Must(contractABI.Pack("sendTo", common.Address{0x2}, big.NewInt(3))),
		})

	req1 := lo.Must(isc.NewEVMOffLedgerTxRequest(env.soloChain.ChainID, tx1))
	req2 := lo.Must(isc.NewEVMOffLedgerTxRequest(env.soloChain.ChainID, tx2))
	env.soloChain.RunRequestsSync([]isc.Request{req1, req2})

	bi := env.soloChain.GetLatestBlockInfo()
	require.EqualValues(t, 2, bi.NumSuccessfulRequests)

	t.Run("callTracer", func(t *testing.T) {
		callTracer := "callTracer"
		var res1 json.RawMessage
		// we have to use the raw client, because the normal client does not support debug methods
		err = env.RawClient.CallContext(
			context.Background(),
			&res1,
			"debug_traceBlockByNumber",
			hexutil.Uint64(env.BlockNumber()).String(),
			tracers.TraceConfig{Tracer: &callTracer},
		)
		require.NoError(t, err)

		var res2 json.RawMessage
		// we have to use the raw client, because the normal client does not support debug methods
		err = env.RawClient.CallContext(
			context.Background(),
			&res2,
			"debug_traceBlockByHash",
			env.BlockByNumber(big.NewInt(int64(env.BlockNumber()))).Hash(),
			tracers.TraceConfig{Tracer: &callTracer},
		)
		require.NoError(t, err)

		require.Equal(t, res1, res2, "debug_traceBlockByNumber and debug_traceBlockByNumber should produce equal results")

		traceBlock := make([]jsonrpc.TxTraceResult, 0)
		err = json.Unmarshal(res1, &traceBlock)
		require.NoError(t, err)

		require.Len(t, traceBlock, 2)

		var trace1 jsonrpc.CallFrame
		err = json.Unmarshal(traceBlock[slices.IndexFunc(traceBlock, func(v jsonrpc.TxTraceResult) bool {
			return v.TxHash == tx1.Hash()
		})].Result, &trace1)
		require.NoError(t, err)

		var trace2 jsonrpc.CallFrame
		err = json.Unmarshal(traceBlock[slices.IndexFunc(traceBlock, func(v jsonrpc.TxTraceResult) bool {
			return v.TxHash == tx2.Hash()
		})].Result, &trace2)
		require.NoError(t, err)

		require.Equal(t, creatorAddress, trace1.From)
		require.Equal(t, contractAddress, *trace1.To)
		require.Equal(t, "0x7b", trace1.Value.String())
		expectedInput, err := contractABI.Pack("sendTo", common.Address{0x1}, big.NewInt(2)) //nolint:govet
		require.NoError(t, err)
		require.Equal(t, expectedInput, []byte(trace1.Input))
		require.Empty(t, trace1.Error)
		require.Empty(t, trace1.RevertReason)
		require.Contains(t, trace1.Gas.String(), "0x")
		require.Contains(t, trace1.GasUsed.String(), "0x")

		require.Len(t, trace1.Calls, 1)
		innerCall1 := trace1.Calls[0]
		require.Equal(t, contractAddress, innerCall1.From)
		require.Equal(t, common.Address{0x1}, *innerCall1.To)
		require.Equal(t, "0x2", innerCall1.Value.String())
		require.Empty(t, innerCall1.Input)
		require.Empty(t, innerCall1.Error)
		require.Empty(t, innerCall1.RevertReason)
		require.Contains(t, innerCall1.Gas.String(), "0x")
		require.Contains(t, innerCall1.GasUsed.String(), "0x")

		require.Equal(t, creatorAddress2, trace2.From)
		require.Equal(t, contractAddress, *trace2.To)
		require.Equal(t, "0x141", trace2.Value.String())
		expectedInput, err = contractABI.Pack("sendTo", common.Address{0x2}, big.NewInt(3))
		require.NoError(t, err)
		require.Equal(t, expectedInput, []byte(trace2.Input))
		require.Empty(t, trace2.Error)
		require.Empty(t, trace2.RevertReason)
		require.Contains(t, trace2.Gas.String(), "0x")
		require.Contains(t, trace2.GasUsed.String(), "0x")

		require.Len(t, trace2.Calls, 1)
		innerCall2 := trace2.Calls[0]
		require.Equal(t, contractAddress, innerCall2.From)
		require.Equal(t, common.Address{0x2}, *innerCall2.To)
		require.Equal(t, "0x3", innerCall2.Value.String())
		require.Empty(t, innerCall2.Input)
		require.Empty(t, innerCall2.Error)
		require.Empty(t, innerCall2.RevertReason)
		require.Contains(t, innerCall2.Gas.String(), "0x")
		require.Contains(t, innerCall2.GasUsed.String(), "0x")
	})
	t.Run("prestate", func(t *testing.T) {
		prestateTracer := "prestateTracer"
		var res2 json.RawMessage
		// we have to use the raw client, because the normal client does not support debug methods
		err = env.RawClient.CallContext(
			context.Background(),
			&res2,
			"debug_traceBlockByHash",
			env.BlockByNumber(big.NewInt(int64(env.BlockNumber()))).Hash(),
			tracers.TraceConfig{
				Tracer:       &prestateTracer,
				TracerConfig: []byte(`{"diffMode": false}`),
			},
		)
		require.NoError(t, err)
		var results []jsonrpc.TxTraceResult
		err = json.Unmarshal(res2, &results)
		require.NoError(t, err)
		require.Len(t, results, 2)
		for _, r := range results {
			var p jsonrpc.PrestateAccountMap
			err = json.Unmarshal(r.Result, &p)
			require.NoError(t, err)
			require.NotEmpty(t, p)
		}
	})
	t.Run("trace_block", func(t *testing.T) {
		var res json.RawMessage
		err = env.RawClient.CallContext(
			context.Background(),
			&res,
			"trace_block",
			rpc.BlockNumber(env.BlockNumber()),
		)
		require.NoError(t, err)

		var traces []*jsonrpc.Trace
		err = json.Unmarshal(res, &traces)
		require.NoError(t, err)
		require.Len(t, traces, 4)

		traceTx1Index := slices.IndexFunc(traces, func(v *jsonrpc.Trace) bool {
			return *v.TransactionHash == tx1.Hash()
		})
		traceTx2Index := slices.IndexFunc(traces, func(v *jsonrpc.Trace) bool {
			return *v.TransactionHash == tx2.Hash()
		})

		call11 := traces[traceTx1Index]
		call12 := traces[traceTx1Index+1]
		call21 := traces[traceTx2Index]
		call22 := traces[traceTx2Index+1]

		call11Action := call11.Action.(map[string]any)
		call12Action := call12.Action.(map[string]any)
		call21Action := call21.Action.(map[string]any)
		call22Action := call22.Action.(map[string]any)

		require.Equal(t, strings.ToLower(creatorAddress.String()), strings.ToLower(call11Action["from"].(string)))
		require.Equal(t, strings.ToLower(contractAddress.String()), strings.ToLower(call11Action["to"].(string)))
		require.Equal(t, "0x7b", call11Action["value"].(string))
		expectedInput, err := contractABI.Pack("sendTo", common.Address{0x1}, big.NewInt(2))
		require.NoError(t, err)
		require.Equal(t, hex.EncodeToString(expectedInput), call11Action["input"].(string)[2:])
		require.Empty(t, call11.Error)
		require.Equal(t, 1, call11.Subtraces)
		require.Equal(t, []int{}, call11.TraceAddress)

		require.Equal(t, strings.ToLower(contractAddress.String()), strings.ToLower(call12Action["from"].(string)))
		require.Equal(t, common.Address{0x1}.String(), strings.ToLower(call12Action["to"].(string)))
		require.Equal(t, "0x2", call12Action["value"].(string))
		require.Equal(t, "0x", call12Action["input"])
		require.Empty(t, call12.Error)
		require.Equal(t, 0, call12.Subtraces)
		require.Equal(t, []int{0}, call12.TraceAddress)

		require.Equal(t, strings.ToLower(creatorAddress2.String()), strings.ToLower(call21Action["from"].(string)))
		require.Equal(t, strings.ToLower(contractAddress.String()), strings.ToLower(call21Action["to"].(string)))
		require.Equal(t, "0x141", call21Action["value"].(string))
		expectedInput, err = contractABI.Pack("sendTo", common.Address{0x2}, big.NewInt(3))
		require.NoError(t, err)
		require.Equal(t, hex.EncodeToString(expectedInput), call21Action["input"].(string)[2:])
		require.Empty(t, call21.Error)
		require.Equal(t, 1, call21.Subtraces)
		require.Equal(t, []int{}, call21.TraceAddress)

		require.Equal(t, strings.ToLower(contractAddress.String()), strings.ToLower(call22Action["from"].(string)))
		require.Equal(t, common.Address{0x2}.String(), strings.ToLower(call22Action["to"].(string)))
		require.Equal(t, "0x3", call22Action["value"].(string))
		require.Equal(t, "0x", call22Action["input"])
		require.Empty(t, call22.Error)
		require.Equal(t, 0, call22.Subtraces)
		require.Equal(t, []int{0}, call22.TraceAddress)
	})
}

func TestRPCTraceBlockSingleCall(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, creatorAddress := env.soloChain.NewEthereumAccountWithL2Funds()
	contractABI, err := abi.JSON(strings.NewReader(evmtest.ISCTestContractABI))
	require.NoError(t, err)
	_, _, contractAddress := env.DeployEVMContract(creator, contractABI, evmtest.ISCTestContractBytecode)

	// make it so that 2 requests are included in the same block
	tx1 := types.MustSignNewTx(creator, types.NewEIP155Signer(big.NewInt(int64(env.ChainID))),
		&types.LegacyTx{
			Nonce:    env.NonceAt(creatorAddress),
			To:       &contractAddress,
			Value:    big.NewInt(123),
			Gas:      100000,
			GasPrice: big.NewInt(10000000000),
			Data:     lo.Must(contractABI.Pack("sendTo", common.Address{0x1}, big.NewInt(2))),
		})

	req1 := lo.Must(isc.NewEVMOffLedgerTxRequest(env.soloChain.ChainID, tx1))
	env.soloChain.RunRequestsSync([]isc.Request{req1})

	bi := env.soloChain.GetLatestBlockInfo()
	require.EqualValues(t, 1, bi.NumSuccessfulRequests)

	callTracer := "callTracer"
	var res1 json.RawMessage
	// we have to use the raw client, because the normal client does not support debug methods
	err = env.RawClient.CallContext(
		context.Background(),
		&res1,
		"debug_traceBlockByNumber",
		hexutil.Uint64(env.BlockNumber()).String(),
		tracers.TraceConfig{Tracer: &callTracer},
	)
	require.NoError(t, err)

	var res2 json.RawMessage
	// we have to use the raw client, because the normal client does not support debug methods
	err = env.RawClient.CallContext(
		context.Background(),
		&res2,
		"debug_traceBlockByHash",
		env.BlockByNumber(big.NewInt(int64(env.BlockNumber()))).Hash(),
		tracers.TraceConfig{Tracer: &callTracer},
	)
	require.NoError(t, err)

	require.Equal(t, res1, res2, "debug_traceBlockByNumber and debug_traceBlockByHash should produce equal results")

	traceBlock := make([]jsonrpc.TxTraceResult, 0)
	err = json.Unmarshal(res1, &traceBlock)
	require.NoError(t, err)

	require.Len(t, traceBlock, 1)

	var trace1 jsonrpc.CallFrame
	err = json.Unmarshal(traceBlock[slices.IndexFunc(traceBlock, func(v jsonrpc.TxTraceResult) bool {
		return v.TxHash == tx1.Hash()
	})].Result, &trace1)
	require.NoError(t, err)

	require.Equal(t, creatorAddress, trace1.From)
	require.Equal(t, contractAddress, *trace1.To)
	require.Equal(t, "0x7b", trace1.Value.String())
	expectedInput, err := contractABI.Pack("sendTo", common.Address{0x1}, big.NewInt(2))
	require.NoError(t, err)
	require.Equal(t, expectedInput, []byte(trace1.Input))
	require.Empty(t, trace1.Error)
	require.Empty(t, trace1.RevertReason)
	require.Contains(t, trace1.Gas.String(), "0x")
	require.Contains(t, trace1.GasUsed.String(), "0x")

	require.Len(t, trace1.Calls, 1)
	innerCall1 := trace1.Calls[0]
	require.Equal(t, contractAddress, innerCall1.From)
	require.Equal(t, common.Address{0x1}, *innerCall1.To)
	require.Equal(t, "0x2", innerCall1.Value.String())
	require.Empty(t, innerCall1.Input)
	require.Empty(t, innerCall1.Error)
	require.Empty(t, innerCall1.RevertReason)
	require.Contains(t, innerCall1.Gas.String(), "0x")
	require.Contains(t, innerCall1.GasUsed.String(), "0x")
}

func TestRPCBlockReceipt(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, creatorAddress := env.soloChain.NewEthereumAccountWithL2Funds()
	creator2, creatorAddress2 := env.soloChain.NewEthereumAccountWithL2Funds()
	contractABI, err := abi.JSON(strings.NewReader(evmtest.ISCTestContractABI))
	require.NoError(t, err)
	_, _, contractAddress := env.DeployEVMContract(creator, contractABI, evmtest.ISCTestContractBytecode)

	tx1 := types.MustSignNewTx(creator, types.NewEIP155Signer(big.NewInt(int64(env.ChainID))),
		&types.LegacyTx{
			Nonce:    env.NonceAt(creatorAddress),
			To:       &contractAddress,
			Value:    big.NewInt(123),
			Gas:      100000,
			GasPrice: big.NewInt(10000000000),
			Data:     lo.Must(contractABI.Pack("sendTo", common.Address{0x1}, big.NewInt(2))),
		})

	tx2 := types.MustSignNewTx(creator2, types.NewEIP155Signer(big.NewInt(int64(env.ChainID))),
		&types.LegacyTx{
			Nonce:    env.NonceAt(creatorAddress2),
			To:       &contractAddress,
			Value:    big.NewInt(321),
			Gas:      100000,
			GasPrice: big.NewInt(10000000000),
			Data:     lo.Must(contractABI.Pack("sendTo", common.Address{0x2}, big.NewInt(3))),
		})

	req1 := lo.Must(isc.NewEVMOffLedgerTxRequest(env.soloChain.ChainID, tx1))
	req2 := lo.Must(isc.NewEVMOffLedgerTxRequest(env.soloChain.ChainID, tx2))
	env.soloChain.RunRequestsSync([]isc.Request{req1, req2})

	bi := env.soloChain.GetLatestBlockInfo()
	require.EqualValues(t, 2, bi.NumSuccessfulRequests)

	receipts, err := env.Client.BlockReceipts(
		context.Background(),
		rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(env.BlockNumber())),
	)
	require.NoError(t, err)
	require.Len(t, receipts, 2)

	r1 := receipts[slices.IndexFunc(receipts, func(v *types.Receipt) bool {
		return v.TxHash == tx1.Hash()
	})]

	r2 := receipts[slices.IndexFunc(receipts, func(v *types.Receipt) bool {
		return v.TxHash == tx2.Hash()
	})]

	require.Equal(t, uint64(1), r1.Status)
	require.Equal(t, big.NewInt(4), r1.BlockNumber)
	require.Equal(t, uint64(1), r2.Status)
	require.Equal(t, big.NewInt(4), r2.BlockNumber)

	// Test the same block with its hash.
	block := env.BlockByNumber(new(big.Int).SetUint64(env.BlockNumber()))
	receipts, err = env.Client.BlockReceipts(
		context.Background(),
		rpc.BlockNumberOrHashWithHash(block.Hash(), false),
	)
	require.NoError(t, err)

	require.Len(t, receipts, 2)

	r1 = receipts[slices.IndexFunc(receipts, func(v *types.Receipt) bool {
		return v.TxHash == tx1.Hash()
	})]

	r2 = receipts[slices.IndexFunc(receipts, func(v *types.Receipt) bool {
		return v.TxHash == tx2.Hash()
	})]

	require.Equal(t, uint64(1), r1.Status)
	require.Equal(t, big.NewInt(4), r1.BlockNumber)
	require.Equal(t, uint64(1), r2.Status)
	require.Equal(t, big.NewInt(4), r2.BlockNumber)

	// Test "latest" block
	err = env.RawClient.CallContext(
		context.Background(),
		&receipts,
		"eth_getBlockReceipts",
		"latest")
	require.NoError(t, err)

	require.Len(t, receipts, 2)

	r1 = receipts[slices.IndexFunc(receipts, func(v *types.Receipt) bool {
		return v.TxHash == tx1.Hash()
	})]

	r2 = receipts[slices.IndexFunc(receipts, func(v *types.Receipt) bool {
		return v.TxHash == tx2.Hash()
	})]

	require.Equal(t, uint64(1), r1.Status)
	require.Equal(t, big.NewInt(4), r1.BlockNumber)
	require.Equal(t, uint64(1), r2.Status)
	require.Equal(t, big.NewInt(4), r2.BlockNumber)
}

func BenchmarkRPCEstimateGas(b *testing.B) {
	env := newSoloTestEnv(b)
	_, addr := env.soloChain.NewEthereumAccountWithL2Funds()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n, err := env.Client.EstimateGas(context.Background(), ethereum.CallMsg{
			From: addr,
		})
		require.NoError(b, err)
		require.NotZero(b, n)
	}
}

func TestSupportsInterfaceRPCEthCall(t *testing.T) {
	env := newSoloTestEnv(t)
	_, creatorAddress := env.soloChain.NewEthereumAccountWithL2Funds()

	blockNumber, err := env.Client.BlockNumber(context.Background())
	require.NoError(t, err)

	var contractAddr common.Address
	copy(contractAddr[:], iscmagic.AddressPrefix)

	cases := []struct {
		interfaceID string
		expected    bool
		description string
	}{
		{"01ffc9a7", true, "ERC165"},
		{"36372b07", false, "ERC20"},
		{"d9b67a26", false, "ERC721"},
	}

	for _, c := range cases {
		interfaceID := fmt.Sprintf("01ffc9a7%s00000000000000000000000000000000000000000000000000000000", c.interfaceID)
		t.Run(fmt.Sprintf("supportsInterface with parameter %s (%s) should return %t", c.interfaceID, c.description, c.expected), func(t *testing.T) {
			callMsg := ethereum.CallMsg{
				From: creatorAddress,
				To:   &contractAddr,
				Data: common.Hex2Bytes(interfaceID),
			}
			result, err := env.Client.CallContract(context.Background(), callMsg, big.NewInt(int64(blockNumber)))
			require.NoError(t, err)
			require.NotEmpty(t, result)
			decodedBool := new(big.Int).SetBytes(result).Uint64() == 1
			require.Equal(t, c.expected, decodedBool)
		})
	}
}
