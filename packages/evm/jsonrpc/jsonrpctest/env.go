// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package jsonrpctest provides testing utilities for the jsonrpc package
package jsonrpctest

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/isc"
)

// Env is a testing environment for the EVM JSON-RPC support, allowing to run the same tests
// with both Solo and cluster backends.
type Env struct {
	T                     testing.TB
	Client                *ethclient.Client
	RawClient             *rpc.Client
	ChainID               uint16
	accountManager        *jsonrpc.AccountManager
	WaitTxConfirmed       func(common.Hash) error
	NewAccountWithL2Funds func(baseTokens ...coin.Value) (*ecdsa.PrivateKey, common.Address)
}

func (e *Env) Signer() types.Signer {
	return evmutil.Signer(big.NewInt(int64(e.ChainID)))
}

func (e *Env) DeployEVMContract(creator *ecdsa.PrivateKey, contractABI abi.ABI, contractBytecode []byte, args ...interface{}) (*types.Transaction, *types.Receipt, common.Address) {
	creatorAddress := crypto.PubkeyToAddress(creator.PublicKey)

	nonce := e.NonceAt(creatorAddress)

	constructorArguments, err := contractABI.Pack("", args...)
	require.NoError(e.T, err)

	data := concatenate(contractBytecode, constructorArguments)

	value := big.NewInt(0)

	gasLimit := e.estimateGas(ethereum.CallMsg{
		From:  creatorAddress,
		To:    nil, // contract creation
		Value: value,
		Data:  data,
	})

	tx, err := types.SignTx(
		types.NewContractCreation(nonce, value, gasLimit, e.MustGetGasPrice(), data),
		e.Signer(),
		creator,
	)
	require.NoError(e.T, err)

	receipt := e.mustSendTransactionAndWait(tx)

	addr := crypto.CreateAddress(creatorAddress, nonce)

	e.T.Logf("deployed EVM contract %s", addr)
	return tx, receipt, addr
}

func (e *Env) mustSendTransactionAndWait(tx *types.Transaction) *types.Receipt {
	r, err := e.SendTransactionAndWait(tx)
	require.NoError(e.T, err)
	return r
}

func (e *Env) SendTransactionAndWait(tx *types.Transaction) (*types.Receipt, error) {
	if err := e.Client.SendTransaction(context.Background(), tx); err != nil {
		return nil, err
	}

	if e.WaitTxConfirmed != nil {
		if err := e.WaitTxConfirmed(tx.Hash()); err != nil {
			return nil, err
		}
	}

	return e.TxReceipt(tx.Hash())
}

func (e *Env) deployStorageContract(creator *ecdsa.PrivateKey) (*types.Transaction, common.Address, abi.ABI) {
	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(e.T, err)
	tx, _, addr := e.DeployEVMContract(creator, contractABI, evmtest.StorageContractBytecode, uint32(42))
	return tx, addr, contractABI
}

func concatenate(a, b []byte) []byte {
	r := make([]byte, 0, len(a)+len(b))
	r = append(r, a...)
	r = append(r, b...)
	return r
}

func (e *Env) estimateGas(msg ethereum.CallMsg) uint64 {
	gas, err := e.Client.EstimateGas(context.Background(), msg)
	require.NoError(e.T, err)
	return gas
}

func (e *Env) NonceAt(address common.Address) uint64 {
	nonce, err := e.Client.NonceAt(context.Background(), address, nil)
	require.NoError(e.T, err)
	return nonce
}

func (e *Env) BlockNumber() uint64 {
	blockNumber, err := e.Client.BlockNumber(context.Background())
	require.NoError(e.T, err)
	return blockNumber
}

func (e *Env) BlockByNumber(number *big.Int) *types.Block {
	block, err := e.Client.BlockByNumber(context.Background(), number)
	require.NoError(e.T, err)
	return block
}

func (e *Env) BlockByHash(hash common.Hash) *types.Block {
	block, err := e.Client.BlockByHash(context.Background(), hash)
	if errors.Is(err, ethereum.NotFound) {
		return nil
	}
	require.NoError(e.T, err)
	return block
}

func (e *Env) TransactionByHash(hash common.Hash) *types.Transaction {
	tx, isPending, err := e.Client.TransactionByHash(context.Background(), hash)
	if errors.Is(err, ethereum.NotFound) {
		return nil
	}
	require.NoError(e.T, err)
	require.False(e.T, isPending)
	return tx
}

func (e *Env) TransactionByBlockHashAndIndex(blockHash common.Hash, index uint) *types.Transaction {
	tx, err := e.Client.TransactionInBlock(context.Background(), blockHash, index)
	if errors.Is(err, ethereum.NotFound) {
		return nil
	}
	require.NoError(e.T, err)
	return tx
}

func (e *Env) UncleByBlockHashAndIndex(blockHash common.Hash, index uint) map[string]interface{} {
	var uncle map[string]interface{}
	err := e.RawClient.Call(&uncle, "eth_getUncleByBlockHashAndIndex", blockHash, hexutil.Uint(index))
	require.NoError(e.T, err)
	return uncle
}

func (e *Env) TransactionByBlockNumberAndIndex(blockNumber *big.Int, index uint) (*jsonrpc.RPCTransaction, error) {
	var tx *jsonrpc.RPCTransaction
	err := e.RawClient.Call(&tx, "eth_getTransactionByBlockNumberAndIndex", (*hexutil.Big)(blockNumber), hexutil.Uint(index))
	return tx, err
}

func (e *Env) UncleByBlockNumberAndIndex(blockNumber *big.Int, index uint) map[string]interface{} {
	var uncle map[string]interface{}
	err := e.RawClient.Call(&uncle, "eth_getUncleByBlockNumberAndIndex", (*hexutil.Big)(blockNumber), hexutil.Uint(index))
	require.NoError(e.T, err)
	return uncle
}

func (e *Env) BlockTransactionCountByHash(hash common.Hash) uint {
	n, err := e.Client.TransactionCount(context.Background(), hash)
	require.NoError(e.T, err)
	return n
}

func (e *Env) UncleCountByBlockHash(hash common.Hash) uint {
	var res hexutil.Uint
	err := e.RawClient.Call(&res, "eth_getUncleCountByBlockHash", hash)
	require.NoError(e.T, err)
	return uint(res)
}

func (e *Env) BlockTransactionCountByNumber() uint {
	// the client only supports calling this method with "pending"
	n, err := e.Client.PendingTransactionCount(context.Background())
	require.NoError(e.T, err)
	return n
}

func (e *Env) UncleCountByBlockNumber(blockNumber *big.Int) uint {
	var res hexutil.Uint
	err := e.RawClient.Call(&res, "eth_getUncleCountByBlockNumber", (*hexutil.Big)(blockNumber))
	require.NoError(e.T, err)
	return uint(res)
}

func (e *Env) BalanceAt(address common.Address, blockNumber *big.Int) *big.Int {
	bal, err := e.Client.BalanceAt(context.Background(), address, blockNumber)
	require.NoError(e.T, err)
	return bal
}

func (e *Env) Balance(address common.Address) *big.Int {
	return e.BalanceAt(address, nil)
}

func (e *Env) Code(address common.Address) []byte {
	code, err := e.Client.CodeAt(context.Background(), address, nil)
	require.NoError(e.T, err)
	return code
}

func (e *Env) Storage(address common.Address, key common.Hash) []byte {
	data, err := e.Client.StorageAt(context.Background(), address, key, nil)
	require.NoError(e.T, err)
	return data
}

func (e *Env) TxReceipt(hash common.Hash) (*types.Receipt, error) {
	return e.Client.TransactionReceipt(context.Background(), hash)
}

func (e *Env) MustTxReceipt(hash common.Hash) *types.Receipt {
	r, err := e.TxReceipt(hash)
	require.NoError(e.T, err)
	return r
}

func (e *Env) Accounts() []common.Address {
	var res []common.Address
	err := e.RawClient.Call(&res, "eth_accounts")
	require.NoError(e.T, err)
	return res
}

func (e *Env) Sign(address common.Address, data []byte) []byte {
	var res hexutil.Bytes
	err := e.RawClient.Call(&res, "eth_sign", address, hexutil.Bytes(data))
	require.NoError(e.T, err)
	return res
}

func (e *Env) SignTransaction(args *jsonrpc.SendTxArgs) []byte {
	var res hexutil.Bytes
	err := e.RawClient.Call(&res, "eth_signTransaction", args)
	require.NoError(e.T, err)
	return res
}

func (e *Env) SendTransaction(args *jsonrpc.SendTxArgs) (common.Hash, error) {
	var res common.Hash
	err := e.RawClient.Call(&res, "eth_sendTransaction", args)
	return res, err
}

func (e *Env) MustSendTransaction(args *jsonrpc.SendTxArgs) common.Hash {
	res, err := e.SendTransaction(args)
	require.NoError(e.T, err)
	return res
}

func (e *Env) MustGetGasPrice() *big.Int {
	res, err := e.Client.SuggestGasPrice(context.Background())
	require.NoError(e.T, err)
	return res
}

func (e *Env) getLogs(q ethereum.FilterQuery) []types.Log {
	logs, err := e.Client.FilterLogs(context.Background(), q)
	require.NoError(e.T, err)
	return logs
}

func (e *Env) traceTransactionWithCallTracer(txHash common.Hash) (jsonrpc.CallFrame, error) {
	var res json.RawMessage
	// we have to use the raw client, because the normal client does not support debug methods
	tracer := "callTracer"
	err := e.RawClient.CallContext(
		context.Background(),
		&res,
		"debug_traceTransaction",
		txHash,
		tracers.TraceConfig{Tracer: &tracer},
	)
	if err != nil {
		return jsonrpc.CallFrame{}, err
	}
	trace := jsonrpc.CallFrame{}
	err = json.Unmarshal(res, &trace)
	require.NoError(e.T, err)
	return trace, nil
}

func (e *Env) traceTransactionWithPrestate(txHash common.Hash) (jsonrpc.PrestateAccountMap, error) {
	var res json.RawMessage
	// we have to use the raw client, because the normal client does not support debug methods
	tracer := "prestateTracer"
	err := e.RawClient.CallContext(
		context.Background(),
		&res,
		"debug_traceTransaction",
		txHash,
		tracers.TraceConfig{
			Tracer:       &tracer,
			TracerConfig: []byte(`{"diffMode": false}`),
		},
	)
	if err != nil {
		return nil, err
	}
	var ret jsonrpc.PrestateAccountMap
	err = json.Unmarshal(res, &ret)
	require.NoError(e.T, err)
	return ret, nil
}

func (e *Env) traceTransactionWithPrestateDiff(txHash common.Hash) (jsonrpc.PrestateDiffResult, error) {
	var res json.RawMessage
	// we have to use the raw client, because the normal client does not support debug methods
	tracer := "prestateTracer"
	err := e.RawClient.CallContext(
		context.Background(),
		&res,
		"debug_traceTransaction",
		txHash,
		tracers.TraceConfig{
			Tracer:       &tracer,
			TracerConfig: []byte(`{"diffMode": true}`),
		},
	)
	if err != nil {
		return jsonrpc.PrestateDiffResult{}, err
	}
	var ret jsonrpc.PrestateDiffResult
	err = json.Unmarshal(res, &ret)
	require.NoError(e.T, err)
	return ret, nil
}

func (e *Env) TestRPCGetLogs() {
	creator, creatorAddress := e.NewAccountWithL2Funds()
	contractABI, err := abi.JSON(strings.NewReader(evmtest.ERC20ContractABI))
	require.NoError(e.T, err)
	contractAddress := crypto.CreateAddress(creatorAddress, e.NonceAt(creatorAddress))

	filterQuery := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	require.Empty(e.T, e.getLogs(filterQuery))

	_, receipt, _ := e.DeployEVMContract(creator, contractABI, evmtest.ERC20ContractBytecode, "TestCoin", "TEST")
	require.Equal(e.T, 1, len(receipt.Logs))

	require.Equal(e.T, 1, len(e.getLogs(filterQuery)))

	_, recipientAddress := e.NewAccountWithL2Funds()
	callArguments, err := contractABI.Pack("transfer", recipientAddress, big.NewInt(1337))
	require.NoError(e.T, err)
	value := big.NewInt(0)
	gas := e.estimateGas(ethereum.CallMsg{
		From:  creatorAddress,
		To:    &contractAddress,
		Value: value,
		Data:  callArguments,
	})
	transferTx, err := types.SignTx(
		types.NewTransaction(e.NonceAt(creatorAddress), contractAddress, value, gas, e.MustGetGasPrice(), callArguments),
		e.Signer(),
		creator,
	)
	require.NoError(e.T, err)
	transferReceipt := e.mustSendTransactionAndWait(transferTx)
	require.Equal(e.T, 1, len(transferReceipt.Logs))
	require.Equal(e.T, 2, len(e.getLogs(filterQuery)))
}

func (e *Env) TestRPCInvalidNonce() {
	from, _ := e.NewAccountWithL2Funds()
	_, toAddress := e.NewAccountWithL2Funds()
	// try sending correct nonces in invalid order 1,2, then 0 - this should succeed
	createTx := func(nonce uint64) *types.Transaction {
		tx, err := types.SignTx(
			types.NewTransaction(nonce, toAddress, big.NewInt(0), math.MaxUint64, e.MustGetGasPrice(), nil),
			e.Signer(),
			from,
		)
		require.NoError(e.T, err)
		return tx
	}

	err := e.Client.SendTransaction(context.Background(), createTx(1))
	require.NoError(e.T, err)
	err = e.Client.SendTransaction(context.Background(), createTx(2))
	require.NoError(e.T, err)
	_, err = e.SendTransactionAndWait(createTx(0))
	require.NoError(e.T, err)

	// try sending nonce 0 again
	_, err = e.SendTransactionAndWait(createTx(0))
	require.Error(e.T, err)
	require.Regexp(e.T, `invalid transaction nonce: got 0, want 3`, err.Error())
	_, ok := err.(*isc.VMError)
	require.False(e.T, ok)
}

func (e *Env) TestRPCGasLimitTooLow() {
	from, fromAddress := e.NewAccountWithL2Funds()
	_, toAddress := e.NewAccountWithL2Funds()
	value := big.NewInt(0)
	nonce := e.NonceAt(fromAddress)
	gasLimit := uint64(1) // lower than intrinsic gas
	tx, err := types.SignTx(
		types.NewTransaction(nonce, toAddress, value, gasLimit, e.MustGetGasPrice(), nil),
		e.Signer(),
		from,
	)
	require.NoError(e.T, err)

	_, err = e.SendTransactionAndWait(tx)
	require.Error(e.T, err)
	require.Regexp(e.T, "intrinsic gas too low", err.Error())
	_, ok := err.(*isc.VMError)
	require.False(e.T, ok)
}

func (e *Env) TestGasPrice() {
	gasPrice := e.MustGetGasPrice()
	require.NotZero(e.T, gasPrice.Uint64())

	// assert sending txs with lower than set gas is not allowed
	from, _ := e.NewAccountWithL2Funds()
	tx, err := types.SignTx(
		types.NewTransaction(0, common.Address{}, big.NewInt(123), math.MaxUint64, new(big.Int).Sub(gasPrice, big.NewInt(1)), nil),
		e.Signer(),
		from,
	)
	require.NoError(e.T, err)

	_, err = e.SendTransactionAndWait(tx)
	require.Error(e.T, err)
	require.Regexp(e.T, `insufficient gas price: got \d+, minimum is \d+`, err.Error())
}

func (e *Env) TestRPCAccessHistoricalState() {
	firstBlockNumber := e.BlockNumber()

	var addrs []common.Address

	for i := 0; i < 10; i++ {
		_, addr := e.NewAccountWithL2Funds()
		addrs = append(addrs, addr)
	}

	require.Equal(e.T, firstBlockNumber+uint64(len(addrs)), e.BlockNumber())

	for i := 0; i < len(addrs); i++ {
		addr := addrs[i]
		n := firstBlockNumber + uint64(i)
		require.Zero(e.T, e.BalanceAt(addr, new(big.Int).SetUint64(n)).Uint64())
		require.NotZero(e.T, e.BalanceAt(addr, new(big.Int).SetUint64(n+1)).Uint64())
	}
}
