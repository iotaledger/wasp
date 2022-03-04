// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package emulator

import (
	"crypto/ecdsa"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/stretchr/testify/require"
)

func estimateGas(t testing.TB, emu *EVMEmulator, from common.Address, to *common.Address, value *big.Int, data []byte) uint64 {
	gas, err := emu.EstimateGas(ethereum.CallMsg{
		From:  from,
		To:    to,
		Value: value,
		Data:  data,
	})
	if err != nil {
		t.Logf("%v", err)
		return evm.BlockGasLimitDefault - 1
	}
	return gas
}

func sendTransaction(t testing.TB, emu *EVMEmulator, sender *ecdsa.PrivateKey, receiverAddress common.Address, amount *big.Int, data []byte) *types.Receipt {
	senderAddress := crypto.PubkeyToAddress(sender.PublicKey)

	nonce := emu.StateDB().GetNonce(senderAddress)
	gas := estimateGas(t, emu, senderAddress, &receiverAddress, amount, data)

	tx, err := types.SignTx(
		types.NewTransaction(nonce, receiverAddress, amount, gas, evm.GasPrice, data),
		emu.Signer(),
		sender,
	)
	require.NoError(t, err)

	receipt, _, err := emu.SendTransaction(tx, tx.Gas())
	require.NoError(t, err)
	emu.MintBlock()

	return receipt
}

func TestBlockchain(t *testing.T) {
	// faucet address with initial supply
	faucet, err := crypto.GenerateKey()
	require.NoError(t, err)
	faucetAddress := crypto.PubkeyToAddress(faucet.PublicKey)
	faucetSupply := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))

	// another account
	receiver, err := crypto.GenerateKey()
	require.NoError(t, err)
	receiverAddress := crypto.PubkeyToAddress(receiver.PublicKey)

	genesisAlloc := map[common.Address]core.GenesisAccount{
		faucetAddress: {Balance: faucetSupply},
	}

	db := dict.Dict{}
	Init(db, evm.DefaultChainID, evm.BlockKeepAll, evm.BlockGasLimitDefault, 0, genesisAlloc)
	emu := NewEVMEmulator(db, 1, nil)

	// some assertions
	{
		require.EqualValues(t, evm.BlockGasLimitDefault, emu.BlockchainDB().GetGasLimit())
		require.EqualValues(t, evm.DefaultChainID, emu.BlockchainDB().GetChainID())

		genesis := emu.BlockchainDB().GetBlockByNumber(0)
		require.NotNil(t, genesis)

		// the genesis block has block number 0
		require.EqualValues(t, 0, genesis.NumberU64())
		// the genesis block has 0 transactions
		require.EqualValues(t, 0, genesis.Transactions().Len())

		{
			// assert that current block is genesis
			block := emu.BlockchainDB().GetCurrentBlock()
			require.NotNil(t, block)
			require.EqualValues(t, evm.BlockGasLimitDefault, block.Header().GasLimit)
		}

		{
			// same, getting the block by hash
			genesis2 := emu.BlockchainDB().GetBlockByHash(genesis.Hash())
			require.NotNil(t, genesis2)
			require.EqualValues(t, genesis.Hash(), genesis2.Hash())
		}

		{
			state := emu.StateDB()
			// check the balance of the faucet address
			require.EqualValues(t, faucetSupply, state.GetBalance(faucetAddress))
			require.EqualValues(t, big.NewInt(0), state.GetBalance(receiverAddress))
		}
	}

	transferAmount := big.NewInt(1000)

	// send a transaction transferring 1000 ETH to receiverAddress
	{
		receipt := sendTransaction(t, emu, faucet, receiverAddress, transferAmount, nil)

		require.EqualValues(t, 1, emu.BlockchainDB().GetNumber())
		block := emu.BlockchainDB().GetCurrentBlock()
		require.EqualValues(t, 1, block.Header().Number.Uint64())
		require.EqualValues(t, evm.BlockGasLimitDefault, block.Header().GasLimit)
		require.EqualValues(t, receipt.Bloom, block.Bloom())
		require.EqualValues(t, receipt.GasUsed, block.GasUsed())
		require.EqualValues(t, emu.BlockchainDB().GetBlockByNumber(0).Hash(), block.ParentHash())
	}

	{
		state := emu.StateDB()
		// check the new balances
		require.EqualValues(t, (&big.Int{}).Sub(faucetSupply, transferAmount), state.GetBalance(faucetAddress))
		require.EqualValues(t, transferAmount, state.GetBalance(receiverAddress))
	}
}

func TestBlockchainPersistence(t *testing.T) {
	// faucet address with initial supply
	faucet, err := crypto.GenerateKey()
	require.NoError(t, err)
	faucetAddress := crypto.PubkeyToAddress(faucet.PublicKey)
	faucetSupply := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))

	genesisAlloc := map[common.Address]core.GenesisAccount{
		faucetAddress: {Balance: faucetSupply},
	}

	// another account
	receiver, err := crypto.GenerateKey()
	require.NoError(t, err)
	receiverAddress := crypto.PubkeyToAddress(receiver.PublicKey)
	transferAmount := big.NewInt(1000)

	db := dict.Dict{}
	Init(db, evm.DefaultChainID, evm.BlockKeepAll, evm.BlockGasLimitDefault, 0, genesisAlloc)

	// do a transfer using one instance of EVMEmulator
	func() {
		emu := NewEVMEmulator(db, 1, nil)
		sendTransaction(t, emu, faucet, receiverAddress, transferAmount, nil)
	}()

	// initialize a new EVMEmulator using the same DB and check the state
	{
		emu := NewEVMEmulator(db, 2, nil)
		state := emu.StateDB()
		// check the new balances
		require.EqualValues(t, (&big.Int{}).Sub(faucetSupply, transferAmount), state.GetBalance(faucetAddress))
		require.EqualValues(t, transferAmount, state.GetBalance(receiverAddress))
	}
}

type contractFnCaller func(sender *ecdsa.PrivateKey, name string, args ...interface{}) *types.Receipt

func deployEVMContract(t testing.TB, emu *EVMEmulator, creator *ecdsa.PrivateKey, contractABI abi.ABI, contractBytecode []byte, args ...interface{}) (common.Address, contractFnCaller) {
	creatorAddress := crypto.PubkeyToAddress(creator.PublicKey)

	nonce := emu.StateDB().GetNonce(creatorAddress)

	txValue := big.NewInt(0)

	// initialize number as 42
	constructorArguments, err := contractABI.Pack("", args...)
	require.NoError(t, err)
	require.NotEmpty(t, constructorArguments)

	data := []byte{}
	data = append(data, contractBytecode...)
	data = append(data, constructorArguments...)

	gas := estimateGas(t, emu, creatorAddress, nil, txValue, data)
	require.NoError(t, err)

	tx, err := types.SignTx(
		types.NewContractCreation(nonce, txValue, gas, evm.GasPrice, data),
		emu.Signer(),
		creator,
	)
	require.NoError(t, err)

	receipt, _, err := emu.SendTransaction(tx, tx.Gas())
	require.NoError(t, err)
	emu.MintBlock()

	contractAddress := crypto.CreateAddress(creatorAddress, nonce)

	// assertions
	{
		require.EqualValues(t, 1, emu.BlockchainDB().GetNumber())

		// verify contract address
		{
			require.EqualValues(t, contractAddress, receipt.ContractAddress)
			require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
		}

		// verify contract code
		{
			code := emu.StateDB().GetCode(contractAddress)
			require.NotEmpty(t, code)
		}
	}

	callFn := func(sender *ecdsa.PrivateKey, name string, args ...interface{}) *types.Receipt {
		callArguments, err := contractABI.Pack(name, args...)
		require.NoError(t, err)
		receipt := sendTransaction(t, emu, sender, contractAddress, big.NewInt(0), callArguments)
		t.Logf("callFn %s Status: %d", name, receipt.Status)
		t.Logf("Logs:")
		for _, log := range receipt.Logs {
			t.Logf("  - %s %x", log.Address, log.Data)
			ev, err := contractABI.EventByID(log.Topics[0])
			if err != nil {
				t.Logf("    - log is not an event: %+v", log.Topics)
			} else {
				t.Logf("    - %s %+v", ev, log.Topics[1:])
			}
		}
		return receipt
	}

	return contractAddress, contractFnCaller(callFn)
}

func TestStorageContract(t *testing.T) {
	// faucet address with initial supply
	faucet, err := crypto.GenerateKey()
	require.NoError(t, err)
	faucetAddress := crypto.PubkeyToAddress(faucet.PublicKey)
	faucetSupply := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))

	genesisAlloc := map[common.Address]core.GenesisAccount{
		faucetAddress: {Balance: faucetSupply},
	}

	db := dict.Dict{}
	Init(db, evm.DefaultChainID, evm.BlockKeepAll, evm.BlockGasLimitDefault, 0, genesisAlloc)
	emu := NewEVMEmulator(db, 1, nil)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)

	contractAddress, callFn := deployEVMContract(
		t,
		emu,
		faucet,
		contractABI,
		evmtest.StorageContractBytecode,
		uint32(42),
	)

	// call `retrieve` view, get 42
	{
		callArguments, err := contractABI.Pack("retrieve")
		require.NoError(t, err)
		require.NotEmpty(t, callArguments)

		res, err := emu.CallContract(ethereum.CallMsg{To: &contractAddress, Data: callArguments})
		require.NoError(t, err)
		require.NotEmpty(t, res)

		var v uint32
		err = contractABI.UnpackIntoInterface(&v, "retrieve", res)
		require.NoError(t, err)
		require.EqualValues(t, 42, v)

		// no state change
		require.EqualValues(t, 1, emu.BlockchainDB().GetNumber())
	}

	// send tx that calls `store(43)`
	{
		callFn(faucet, "store", uint32(43))
		require.EqualValues(t, 2, emu.BlockchainDB().GetNumber())
	}

	// call `retrieve` view again, get 43
	{
		callArguments, err := contractABI.Pack("retrieve")
		require.NoError(t, err)

		res, err := emu.CallContract(ethereum.CallMsg{
			To:   &contractAddress,
			Data: callArguments,
		})
		require.NoError(t, err)
		require.NotEmpty(t, res)

		var v uint32
		err = contractABI.UnpackIntoInterface(&v, "retrieve", res)
		require.NoError(t, err)
		require.EqualValues(t, 43, v)

		// no state change
		require.EqualValues(t, 2, emu.BlockchainDB().GetNumber())
	}
}

func TestERC20Contract(t *testing.T) {
	// faucet address with initial supply
	faucet, err := crypto.GenerateKey()
	require.NoError(t, err)
	faucetAddress := crypto.PubkeyToAddress(faucet.PublicKey)
	faucetSupply := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))

	genesisAlloc := map[common.Address]core.GenesisAccount{
		faucetAddress: {Balance: faucetSupply},
	}

	db := dict.Dict{}
	Init(db, evm.DefaultChainID, evm.BlockKeepAll, evm.BlockGasLimitDefault, 0, genesisAlloc)
	emu := NewEVMEmulator(db, 1, nil)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.ERC20ContractABI))
	require.NoError(t, err)

	erc20Owner, err := crypto.GenerateKey()
	require.NoError(t, err)
	erc20OwnerAddress := crypto.PubkeyToAddress(erc20Owner.PublicKey)

	contractAddress, callFn := deployEVMContract(
		t,
		emu,
		erc20Owner,
		contractABI,
		evmtest.ERC20ContractBytecode,
		"TestCoin",
		"TEST",
	)

	callIntViewFn := func(name string, args ...interface{}) *big.Int {
		callArguments, err := contractABI.Pack(name, args...)
		require.NoError(t, err)

		res, err := emu.CallContract(ethereum.CallMsg{To: &contractAddress, Data: callArguments})
		require.NoError(t, err)

		v := new(big.Int)
		err = contractABI.UnpackIntoInterface(&v, name, res)
		require.NoError(t, err)
		return v
	}

	// call `totalSupply` view
	{
		v := callIntViewFn("totalSupply")
		// 100 * 10^18
		expected := new(big.Int).Mul(big.NewInt(100), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
		require.Zero(t, v.Cmp(expected))
	}

	recipient, err := crypto.GenerateKey()
	require.NoError(t, err)
	recipientAddress := crypto.PubkeyToAddress(recipient.PublicKey)

	transferAmount := big.NewInt(1337)

	// call `transfer` => send 1337 TestCoin to recipientAddress
	receipt := callFn(erc20Owner, "transfer", recipientAddress, transferAmount)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	require.Equal(t, 1, len(receipt.Logs))

	// call `balanceOf` view => check balance of recipient = 1337 TestCoin
	require.Zero(t, callIntViewFn("balanceOf", recipientAddress).Cmp(transferAmount))

	// call `transferFrom` as recipient without allowance => get error
	receipt = callFn(recipient, "transferFrom", erc20OwnerAddress, recipientAddress, transferAmount)
	require.Equal(t, types.ReceiptStatusFailed, receipt.Status)
	require.Equal(t, 0, len(receipt.Logs))

	// call `balanceOf` view => check balance of recipient = 1337 TestCoin
	require.Zero(t, callIntViewFn("balanceOf", recipientAddress).Cmp(transferAmount))

	// call `approve` as erc20Owner
	receipt = callFn(erc20Owner, "approve", recipientAddress, transferAmount)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	require.Equal(t, 1, len(receipt.Logs))

	// call `transferFrom` as recipient with allowance => ok
	receipt = callFn(recipient, "transferFrom", erc20OwnerAddress, recipientAddress, transferAmount)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	require.Equal(t, 1, len(receipt.Logs))

	// call `balanceOf` view => check balance of recipient = 2 * 1337 TestCoin
	require.Zero(t, callIntViewFn("balanceOf", recipientAddress).Cmp(new(big.Int).Mul(transferAmount, big.NewInt(2))))
}

// TODO: test a contract calling selfdestruct

func initBenchmark(b *testing.B) (*EVMEmulator, []*types.Transaction, dict.Dict) {
	// faucet address with initial supply
	faucet, err := crypto.GenerateKey()
	require.NoError(b, err)
	faucetAddress := crypto.PubkeyToAddress(faucet.PublicKey)
	faucetSupply := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))

	genesisAlloc := map[common.Address]core.GenesisAccount{
		faucetAddress: {Balance: faucetSupply},
	}

	db := dict.Dict{}
	Init(db, evm.DefaultChainID, evm.BlockKeepAll, evm.BlockGasLimitDefault, 0, genesisAlloc)
	emu := NewEVMEmulator(db, 1, nil)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(b, err)

	contractAddress, _ := deployEVMContract(
		b,
		emu,
		faucet,
		contractABI,
		evmtest.StorageContractBytecode,
		uint32(42),
	)

	txs := make([]*types.Transaction, b.N)
	for i := 0; i < b.N; i++ {
		sender, err := crypto.GenerateKey() // send from a new address so that nonce is always 0
		require.NoError(b, err)

		amount := big.NewInt(0)
		nonce := uint64(0)

		callArguments, err := contractABI.Pack("store", uint32(i))
		require.NoError(b, err)

		gas := evm.BlockGasLimitDefault

		txs[i], err = types.SignTx(
			types.NewTransaction(nonce, contractAddress, amount, gas, evm.GasPrice, callArguments),
			emu.Signer(),
			sender,
		)
		require.NoError(b, err)
	}

	return emu, txs, db
}

// benchmarkEVMEmulator is a benchmark for the EVMEmulator that sends N EVM transactions
// calling `storage.store()`, committing a block every k transactions.
//
// run with: go test -benchmem -cpu=1 -run=' ' -bench='Bench.*'
//
// To generate mem and cpu profiles, add -cpuprofile=cpu.out -memprofile=mem.out
// Then: go tool pprof -http :8080 {cpu,mem}.out
func benchmarkEVMEmulator(b *testing.B, k int) {
	// setup: deploy the storage contract and prepare N transactions to send
	emu, txs, db := initBenchmark(b)

	var chunks [][]*types.Transaction
	var chunk []*types.Transaction
	for _, tx := range txs {
		chunk = append(chunk, tx)
		if len(chunk) == k {
			chunks = append(chunks, chunk)
			chunk = nil
		}
	}
	if len(chunk) > 0 {
		chunks = append(chunks, chunk)
	}

	b.ResetTimer()
	for _, chunk := range chunks {
		for _, tx := range chunk {
			receipt, _, err := emu.SendTransaction(tx, tx.Gas())
			require.NoError(b, err)
			require.Equal(b, types.ReceiptStatusSuccessful, receipt.Status)
		}
		emu.MintBlock()
	}

	b.ReportMetric(dbSize(db)/float64(b.N), "db:bytes/op")
}

func BenchmarkEVMEmulator1(b *testing.B)   { benchmarkEVMEmulator(b, 1) }
func BenchmarkEVMEmulator10(b *testing.B)  { benchmarkEVMEmulator(b, 10) }
func BenchmarkEVMEmulator50(b *testing.B)  { benchmarkEVMEmulator(b, 50) }
func BenchmarkEVMEmulator100(b *testing.B) { benchmarkEVMEmulator(b, 100) }

func dbSize(db kv.KVStore) float64 {
	r := float64(0)
	db.MustIterate("", func(key kv.Key, value []byte) bool {
		r += float64(len(key) + len(value))
		return true
	})
	return r
}
