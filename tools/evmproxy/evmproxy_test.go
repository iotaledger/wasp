package main

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/tools/evmproxy/service"
	"github.com/stretchr/testify/require"
)

type env struct {
	t      *testing.T
	server *rpc.Server
	client *ethclient.Client
}

func newEnv(t *testing.T) *env {
	soloEVMChain := service.NewEVMChain(service.NewSoloBackend(core.GenesisAlloc{
		faucetAddress: {Balance: faucetSupply},
	}))

	rpcsrv := NewRPCServer(soloEVMChain)
	t.Cleanup(rpcsrv.Stop)

	client := ethclient.NewClient(rpc.DialInProc(rpcsrv))
	t.Cleanup(client.Close)

	return &env{t, rpcsrv, client}
}

func generateKey(t *testing.T) (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return key, addr
}

var requestFundsAmount = big.NewInt(1e18) // 1 ETH

func (e *env) requestFunds(target common.Address) *types.Transaction {
	nonce, err := e.client.NonceAt(context.Background(), faucetAddress, nil)
	require.NoError(e.t, err)
	tx, err := types.SignTx(
		types.NewTransaction(nonce, target, requestFundsAmount, evm.GasLimit, evm.GasPrice, nil),
		evm.Signer(),
		faucetKey,
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

	tx, err := types.SignTx(
		types.NewContractCreation(nonce, big.NewInt(0), evm.GasLimit, evm.GasPrice, data),
		evm.Signer(),
		creator,
	)
	require.NoError(e.t, err)

	err = e.client.SendTransaction(context.Background(), tx)
	require.NoError(e.t, err)

	return tx, crypto.CreateAddress(creatorAddress, nonce)
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
	require.NoError(e.t, err)
	return block
}

func (e *env) balance(address common.Address) *big.Int {
	bal, err := e.client.BalanceAt(context.Background(), address, nil)
	require.NoError(e.t, err)
	return bal
}

func (e *env) code(address common.Address) []byte {
	bal, err := e.client.CodeAt(context.Background(), address, nil)
	require.NoError(e.t, err)
	return bal
}

func (e *env) txReceipt(hash common.Hash) *types.Receipt {
	r, err := e.client.TransactionReceipt(context.Background(), hash)
	require.NoError(e.t, err)
	return r
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
	require.EqualValues(t, 0, env.nonceAt(faucetAddress))
	env.requestFunds(receiverAddress)
	require.EqualValues(t, 1, env.nonceAt(faucetAddress))
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
	require.EqualValues(t, 0, env.blockByHash(env.blockByNumber(big.NewInt(0)).Hash()).Number().Uint64())
	env.requestFunds(receiverAddress)
	require.EqualValues(t, 1, env.blockByHash(env.blockByNumber(big.NewInt(1)).Hash()).Number().Uint64())
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
