package jsonrpc

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	ethlog "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/stretchr/testify/require"
)

type env struct {
	t         *testing.T
	server    *rpc.Server
	rawClient *rpc.Client
	client    *ethclient.Client
}

func newEnv(t *testing.T) *env {
	ethlog.Root().SetHandler(ethlog.FuncHandler(func(r *ethlog.Record) error {
		if r.Lvl <= ethlog.LvlWarn {
			t.Logf("[%s] %s", r.Lvl.AlignedString(), r.Msg)
		}
		return nil
	}))

	solo := NewSoloBackend(core.GenesisAlloc{
		evmtest.FaucetAddress: {Balance: evmtest.FaucetSupply},
	})
	soloEVMChain := NewEVMChain(solo)

	signer, _ := solo.Env.NewKeyPairWithFunds()

	accountManager := NewAccountManager(evmtest.Accounts)

	rpcsrv := NewServer(soloEVMChain, signer, accountManager)
	t.Cleanup(rpcsrv.Stop)

	rawClient := rpc.DialInProc(rpcsrv)
	client := ethclient.NewClient(rawClient)
	t.Cleanup(client.Close)

	return &env{t, rpcsrv, rawClient, client}
}

func generateKey(t *testing.T) (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return key, addr
}

var requestFundsAmount = big.NewInt(1e18) // 1 ETH

func (e *env) requestFunds(target common.Address) *types.Transaction {
	nonce, err := e.client.NonceAt(context.Background(), evmtest.FaucetAddress, nil)
	require.NoError(e.t, err)
	tx, err := types.SignTx(
		types.NewTransaction(nonce, target, requestFundsAmount, evm.TxGas, evm.GasPrice, nil),
		evm.Signer(),
		evmtest.FaucetKey,
	)
	require.NoError(e.t, err)
	err = e.client.SendTransaction(context.Background(), tx)
	require.NoError(e.t, err)
	return tx
}

func (e *env) deployEVMContract(creator *ecdsa.PrivateKey, contractABI abi.ABI, contractBytecode []byte, args ...interface{}) (*types.Transaction, common.Address) {
	creatorAddress := crypto.PubkeyToAddress(creator.PublicKey)

	nonce := e.nonceAt(creatorAddress)

	constructorArguments, err := contractABI.Pack("", args...)
	require.NoError(e.t, err)

	data := append(contractBytecode, constructorArguments...)

	value := big.NewInt(0)

	gasLimit := e.estimateGas(ethereum.CallMsg{
		From:     creatorAddress,
		To:       nil, // contract creation
		GasPrice: evm.GasPrice,
		Value:    value,
		Data:     data,
	})

	tx, err := types.SignTx(
		types.NewContractCreation(nonce, value, gasLimit, evm.GasPrice, data),
		evm.Signer(),
		creator,
	)
	require.NoError(e.t, err)

	err = e.client.SendTransaction(context.Background(), tx)
	require.NoError(e.t, err)

	return tx, crypto.CreateAddress(creatorAddress, nonce)
}

func (e *env) estimateGas(msg ethereum.CallMsg) uint64 {
	gas, err := e.client.EstimateGas(context.Background(), msg)
	require.NoError(e.t, err)
	return gas
}

func (e *env) nonceAt(address common.Address) uint64 {
	nonce, err := e.client.NonceAt(context.Background(), address, nil)
	require.NoError(e.t, err)
	return nonce
}

func (e *env) blockNumber() uint64 {
	blockNumber, err := e.client.BlockNumber(context.Background())
	require.NoError(e.t, err)
	return blockNumber
}

func (e *env) blockByNumber(number *big.Int) *types.Block {
	block, err := e.client.BlockByNumber(context.Background(), number)
	require.NoError(e.t, err)
	return block
}

func (e *env) blockByHash(hash common.Hash) *types.Block {
	block, err := e.client.BlockByHash(context.Background(), hash)
	if err == ethereum.NotFound {
		return nil
	}
	require.NoError(e.t, err)
	return block
}

func (e *env) transactionByHash(hash common.Hash) *types.Transaction {
	tx, isPending, err := e.client.TransactionByHash(context.Background(), hash)
	if err == ethereum.NotFound {
		return nil
	}
	require.NoError(e.t, err)
	require.False(e.t, isPending)
	return tx
}

func (e *env) transactionByBlockHashAndIndex(blockHash common.Hash, index uint) *types.Transaction {
	tx, err := e.client.TransactionInBlock(context.Background(), blockHash, uint(index))
	if err == ethereum.NotFound {
		return nil
	}
	require.NoError(e.t, err)
	return tx
}

func (e *env) uncleByBlockHashAndIndex(blockHash common.Hash, index uint) map[string]interface{} {
	var uncle map[string]interface{}
	err := e.rawClient.Call(&uncle, "eth_getUncleByBlockHashAndIndex", blockHash, hexutil.Uint(index))
	require.NoError(e.t, err)
	return uncle
}

func (e *env) transactionByBlockNumberAndIndex(blockNumber *big.Int, index uint) *RPCTransaction {
	var tx *RPCTransaction
	err := e.rawClient.Call(&tx, "eth_getTransactionByBlockNumberAndIndex", (*hexutil.Big)(blockNumber), hexutil.Uint(index))
	require.NoError(e.t, err)
	return tx
}

func (e *env) uncleByBlockNumberAndIndex(blockNumber *big.Int, index uint) map[string]interface{} {
	var uncle map[string]interface{}
	err := e.rawClient.Call(&uncle, "eth_getUncleByBlockNumberAndIndex", (*hexutil.Big)(blockNumber), hexutil.Uint(index))
	require.NoError(e.t, err)
	return uncle
}

func (e *env) blockTransactionCountByHash(hash common.Hash) uint {
	n, err := e.client.TransactionCount(context.Background(), hash)
	require.NoError(e.t, err)
	return n
}

func (e *env) uncleCountByBlockHash(hash common.Hash) uint {
	var res hexutil.Uint
	err := e.rawClient.Call(&res, "eth_getUncleCountByBlockHash", hash)
	require.NoError(e.t, err)
	return uint(res)
}

func (e *env) blockTransactionCountByNumber() uint {
	// the client only supports calling this method with "pending"
	n, err := e.client.PendingTransactionCount(context.Background())
	require.NoError(e.t, err)
	return n
}

func (e *env) uncleCountByBlockNumber(blockNumber *big.Int) uint {
	var res hexutil.Uint
	err := e.rawClient.Call(&res, "eth_getUncleCountByBlockNumber", (*hexutil.Big)(blockNumber))
	require.NoError(e.t, err)
	return uint(res)
}

func (e *env) balance(address common.Address) *big.Int {
	bal, err := e.client.BalanceAt(context.Background(), address, nil)
	require.NoError(e.t, err)
	return bal
}

func (e *env) code(address common.Address) []byte {
	code, err := e.client.CodeAt(context.Background(), address, nil)
	require.NoError(e.t, err)
	return code
}

func (e *env) storage(address common.Address, key common.Hash) []byte {
	data, err := e.client.StorageAt(context.Background(), address, key, nil)
	require.NoError(e.t, err)
	return data
}

func (e *env) txReceipt(hash common.Hash) *types.Receipt {
	r, err := e.client.TransactionReceipt(context.Background(), hash)
	require.NoError(e.t, err)
	return r
}

func (e *env) accounts() []common.Address {
	var res []common.Address
	err := e.rawClient.Call(&res, "eth_accounts")
	require.NoError(e.t, err)
	return res
}

func (e *env) sign(address common.Address, data []byte) []byte {
	var res hexutil.Bytes
	err := e.rawClient.Call(&res, "eth_sign", address, hexutil.Bytes(data))
	require.NoError(e.t, err)
	return res
}

func (e *env) signTransaction(args *SendTxArgs) []byte {
	var res hexutil.Bytes
	err := e.rawClient.Call(&res, "eth_signTransaction", args)
	require.NoError(e.t, err)
	return res
}

func (e *env) sendTransaction(args *SendTxArgs) common.Hash {
	var res common.Hash
	err := e.rawClient.Call(&res, "eth_sendTransaction", args)
	require.NoError(e.t, err)
	return res
}

func TestRPCGetBalance(t *testing.T) {
	env := newEnv(t)
	_, receiverAddress := generateKey(t)
	require.Zero(t, big.NewInt(0).Cmp(env.balance(receiverAddress)))
	env.requestFunds(receiverAddress)
	require.Zero(t, big.NewInt(1e18).Cmp(env.balance(receiverAddress)))
}

func TestRPCGetCode(t *testing.T) {
	env := newEnv(t)
	creator, creatorAddress := generateKey(t)

	// account address
	{
		env.requestFunds(creatorAddress)
		require.Empty(t, env.code(creatorAddress))
	}
	// contract address
	{
		contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
		require.NoError(t, err)
		_, contractAddress := env.deployEVMContract(creator, contractABI, evmtest.StorageContractBytecode, uint32(42))
		require.NotEmpty(t, env.code(contractAddress))
	}
}

func TestRPCGetStorage(t *testing.T) {
	env := newEnv(t)
	creator, creatorAddress := generateKey(t)

	env.requestFunds(creatorAddress)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)
	_, contractAddress := env.deployEVMContract(creator, contractABI, evmtest.StorageContractBytecode, uint32(42))

	// first static variable in contract (uint32 n) has slot 0. See:
	// https://docs.soliditylang.org/en/v0.6.6/miscellaneous.html#layout-of-state-variables-in-storage
	slot := common.Hash{}
	ret := env.storage(contractAddress, slot)

	var v uint32
	err = contractABI.UnpackIntoInterface(&v, "retrieve", ret)
	require.NoError(t, err)
	require.Equal(t, uint32(42), v)
}

func TestRPCBlockNumber(t *testing.T) {
	env := newEnv(t)
	_, receiverAddress := generateKey(t)
	require.EqualValues(t, 0, env.blockNumber())
	env.requestFunds(receiverAddress)
	require.EqualValues(t, 1, env.blockNumber())
}

func TestRPCGetTransactionCount(t *testing.T) {
	env := newEnv(t)
	_, receiverAddress := generateKey(t)
	require.EqualValues(t, 0, env.nonceAt(evmtest.FaucetAddress))
	env.requestFunds(receiverAddress)
	require.EqualValues(t, 1, env.nonceAt(evmtest.FaucetAddress))
}

func TestRPCGetBlockByNumber(t *testing.T) {
	env := newEnv(t)
	_, receiverAddress := generateKey(t)
	require.EqualValues(t, 0, env.blockByNumber(big.NewInt(0)).Number().Uint64())
	env.requestFunds(receiverAddress)
	require.EqualValues(t, 1, env.blockByNumber(big.NewInt(1)).Number().Uint64())
}

func TestRPCGetBlockByHash(t *testing.T) {
	env := newEnv(t)
	_, receiverAddress := generateKey(t)
	require.Nil(t, env.blockByHash(common.Hash{}))
	require.EqualValues(t, 0, env.blockByHash(env.blockByNumber(big.NewInt(0)).Hash()).Number().Uint64())
	env.requestFunds(receiverAddress)
	require.EqualValues(t, 1, env.blockByHash(env.blockByNumber(big.NewInt(1)).Hash()).Number().Uint64())
}

func TestRPCGetTransactionByHash(t *testing.T) {
	env := newEnv(t)
	_, receiverAddress := generateKey(t)
	require.Nil(t, env.transactionByHash(common.Hash{}))
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(big.NewInt(1))
	tx := env.transactionByHash(block1.Transactions()[0].Hash())
	require.Equal(t, block1.Transactions()[0].Hash(), tx.Hash())
}

func TestRPCGetTransactionByBlockHashAndIndex(t *testing.T) {
	env := newEnv(t)
	_, receiverAddress := generateKey(t)
	require.Nil(t, env.transactionByBlockHashAndIndex(common.Hash{}, 0))
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(big.NewInt(1))
	tx := env.transactionByBlockHashAndIndex(block1.Hash(), 0)
	require.Equal(t, block1.Transactions()[0].Hash(), tx.Hash())
}

func TestRPCGetUncleByBlockHashAndIndex(t *testing.T) {
	env := newEnv(t)
	_, receiverAddress := generateKey(t)
	require.Nil(t, env.uncleByBlockHashAndIndex(common.Hash{}, 0))
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(big.NewInt(1))
	require.Nil(t, env.uncleByBlockHashAndIndex(block1.Hash(), 0))
}

func TestRPCGetTransactionByBlockNumberAndIndex(t *testing.T) {
	env := newEnv(t)
	_, receiverAddress := generateKey(t)
	require.Nil(t, env.transactionByBlockNumberAndIndex(big.NewInt(3), 0))
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(big.NewInt(1))
	tx := env.transactionByBlockNumberAndIndex(block1.Number(), 0)
	require.EqualValues(t, block1.Hash(), *tx.BlockHash)
	require.EqualValues(t, 0, *tx.TransactionIndex)
}

func TestRPCGetUncleByBlockNumberAndIndex(t *testing.T) {
	env := newEnv(t)
	_, receiverAddress := generateKey(t)
	require.Nil(t, env.uncleByBlockNumberAndIndex(big.NewInt(3), 0))
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(big.NewInt(1))
	require.Nil(t, env.uncleByBlockNumberAndIndex(block1.Number(), 0))
}

func TestRPCGetTransactionCountByHash(t *testing.T) {
	env := newEnv(t)
	_, receiverAddress := generateKey(t)
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(big.NewInt(1))
	require.Positive(t, len(block1.Transactions()))
	require.EqualValues(t, len(block1.Transactions()), env.blockTransactionCountByHash(block1.Hash()))
	require.EqualValues(t, 0, env.blockTransactionCountByHash(common.Hash{}))
}

func TestRPCGetUncleCountByBlockHash(t *testing.T) {
	env := newEnv(t)
	_, receiverAddress := generateKey(t)
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(big.NewInt(1))
	require.Zero(t, len(block1.Uncles()))
	require.EqualValues(t, len(block1.Uncles()), env.uncleCountByBlockHash(block1.Hash()))
	require.EqualValues(t, 0, env.uncleCountByBlockHash(common.Hash{}))
}

func TestRPCGetTransactionCountByNumber(t *testing.T) {
	env := newEnv(t)
	_, receiverAddress := generateKey(t)
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(nil)
	require.Positive(t, len(block1.Transactions()))
	require.EqualValues(t, len(block1.Transactions()), env.blockTransactionCountByNumber())
}

func TestRPCGetUncleCountByBlockNumber(t *testing.T) {
	env := newEnv(t)
	_, receiverAddress := generateKey(t)
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(big.NewInt(1))
	require.Zero(t, len(block1.Uncles()))
	require.EqualValues(t, len(block1.Uncles()), env.uncleCountByBlockNumber(big.NewInt(1)))
}

func TestRPCAccounts(t *testing.T) {
	env := newEnv(t)
	accounts := env.accounts()
	require.Equal(t, len(evmtest.Accounts), len(accounts))
}

func TestRPCSign(t *testing.T) {
	env := newEnv(t)
	signed := env.sign(evmtest.AccountAddress(0), []byte("hello"))
	require.NotEmpty(t, signed)
}

func TestRPCSignTransaction(t *testing.T) {
	env := newEnv(t)

	from := evmtest.AccountAddress(0)
	to := evmtest.AccountAddress(1)
	gas := hexutil.Uint64(evm.TxGas)
	nonce := hexutil.Uint64(env.nonceAt(from))
	signed := env.signTransaction(&SendTxArgs{
		From:     from,
		To:       &to,
		Gas:      &gas,
		GasPrice: (*hexutil.Big)(evm.GasPrice),
		Value:    (*hexutil.Big)(requestFundsAmount),
		Nonce:    &nonce,
	})
	require.NotEmpty(t, signed)

	// test that the signed tx can be sent
	env.requestFunds(from)
	err := env.rawClient.Call(nil, "eth_sendRawTransaction", hexutil.Encode(signed))
	require.NoError(t, err)
}

func TestRPCSendTransaction(t *testing.T) {
	env := newEnv(t)

	from := evmtest.AccountAddress(0)
	env.requestFunds(from)

	to := evmtest.AccountAddress(1)
	gas := hexutil.Uint64(evm.TxGas)
	nonce := hexutil.Uint64(env.nonceAt(from))
	txHash := env.sendTransaction(&SendTxArgs{
		From:     from,
		To:       &to,
		Gas:      &gas,
		GasPrice: (*hexutil.Big)(evm.GasPrice),
		Value:    (*hexutil.Big)(requestFundsAmount),
		Nonce:    &nonce,
	})
	require.NotEqualValues(t, common.Hash{}, txHash)
}

func TestRPCGetTxReceipt(t *testing.T) {
	env := newEnv(t)
	creator, creatorAddr := generateKey(t)

	// regular transaction
	{
		tx := env.requestFunds(creatorAddr)
		receipt := env.txReceipt(tx.Hash())

		require.EqualValues(t, types.LegacyTxType, receipt.Type)
		require.EqualValues(t, types.ReceiptStatusSuccessful, receipt.Status)
		require.NotZero(t, receipt.CumulativeGasUsed)
		require.EqualValues(t, types.Bloom{}, receipt.Bloom)
		require.EqualValues(t, 0, len(receipt.Logs))

		require.EqualValues(t, tx.Hash(), receipt.TxHash)
		require.EqualValues(t, common.Address{}, receipt.ContractAddress)
		require.NotZero(t, receipt.GasUsed)

		require.EqualValues(t, big.NewInt(1), receipt.BlockNumber)
		require.EqualValues(t, env.blockByNumber(big.NewInt(1)).Hash(), receipt.BlockHash)
		require.EqualValues(t, 0, receipt.TransactionIndex)
	}

	// contract creation
	{
		contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
		require.NoError(t, err)
		tx, contractAddress := env.deployEVMContract(creator, contractABI, evmtest.StorageContractBytecode, uint32(42))
		receipt := env.txReceipt(tx.Hash())

		require.EqualValues(t, types.LegacyTxType, receipt.Type)
		require.EqualValues(t, types.ReceiptStatusSuccessful, receipt.Status)
		require.NotZero(t, receipt.CumulativeGasUsed)
		require.EqualValues(t, types.Bloom{}, receipt.Bloom)
		require.EqualValues(t, 0, len(receipt.Logs))

		require.EqualValues(t, tx.Hash(), receipt.TxHash)
		require.EqualValues(t, contractAddress, receipt.ContractAddress)
		require.NotZero(t, receipt.GasUsed)

		require.EqualValues(t, big.NewInt(2), receipt.BlockNumber)
		require.EqualValues(t, env.blockByNumber(big.NewInt(2)).Hash(), receipt.BlockHash)
		require.EqualValues(t, 0, receipt.TransactionIndex)
	}
}

func TestRPCCall(t *testing.T) {
	env := newEnv(t)
	creator, creatorAddress := generateKey(t)
	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)
	_, contractAddress := env.deployEVMContract(creator, contractABI, evmtest.StorageContractBytecode, uint32(42))

	callArguments, err := contractABI.Pack("retrieve")
	require.NoError(t, err)

	ret, err := env.client.CallContract(context.Background(), ethereum.CallMsg{
		From: creatorAddress,
		To:   &contractAddress,
		Data: callArguments,
	}, nil)
	require.NoError(t, err)

	var v uint32
	err = contractABI.UnpackIntoInterface(&v, "retrieve", ret)
	require.NoError(t, err)
	require.Equal(t, uint32(42), v)
}
