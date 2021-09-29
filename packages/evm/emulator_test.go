// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"crypto/ecdsa"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
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
		return GasLimitDefault - 1
	}
	return gas
}

func sendTransaction(t testing.TB, emu *EVMEmulator, sender *ecdsa.PrivateKey, receiverAddress common.Address, amount *big.Int, data []byte) *types.Receipt {
	senderAddress := crypto.PubkeyToAddress(sender.PublicKey)

	nonce, err := emu.PendingNonceAt(senderAddress)
	require.NoError(t, err)
	gas := estimateGas(t, emu, senderAddress, &receiverAddress, amount, data)

	tx, err := types.SignTx(
		types.NewTransaction(nonce, receiverAddress, amount, gas, GasPrice, data),
		emu.Signer(),
		sender,
	)
	require.NoError(t, err)

	_, err = emu.SendTransaction(tx)
	require.NoError(t, err)
	emu.Commit()

	receipt, err := emu.TransactionReceipt(tx.Hash())
	require.NoError(t, err)

	return receipt
}

func TestBlockchain(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	defer db.Close()
	testBlockchain(t, db)
}

func TestBlockchainWithKVStoreBackend(t *testing.T) {
	db := rawdb.NewDatabase(NewKVAdapter(dict.New()))
	defer db.Close()
	testBlockchain(t, db)
}

func testBlockchain(t testing.TB, db ethdb.Database) {
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

	InitGenesis(DefaultChainID, db, genesisAlloc, GasLimitDefault, 0)

	emu := NewEVMEmulator(db)
	defer emu.Close()

	genesis := emu.Blockchain().Genesis()

	// some assertions
	{
		require.NotNil(t, genesis)

		// the genesis block has block number 0
		require.EqualValues(t, 0, genesis.NumberU64())
		// the genesis block has 0 transactions
		require.EqualValues(t, 0, genesis.Transactions().Len())

		var genesisHash common.Hash
		{
			// assert that current block is genesis
			block := emu.Blockchain().CurrentBlock()
			require.NotNil(t, block)
			require.EqualValues(t, GasLimitDefault, block.Header().GasLimit)
			genesisHash = block.Hash()
		}

		{
			// same, getting the block by hash
			genesis2 := emu.Blockchain().GetBlockByHash(genesisHash)
			require.NotNil(t, genesis2)
			require.EqualValues(t, genesisHash, genesis2.Hash())
		}

		{
			state, err := emu.Blockchain().State()
			require.NoError(t, err)
			// check the balance of the faucet address
			require.EqualValues(t, faucetSupply, state.GetBalance(faucetAddress))
			require.EqualValues(t, big.NewInt(0), state.GetBalance(receiverAddress))
		}
	}

	transferAmount := big.NewInt(1000)

	// send a transaction transferring 1000 ETH to receiverAddress
	{
		sendTransaction(t, emu, faucet, receiverAddress, transferAmount, nil)

		require.EqualValues(t, 1, emu.Blockchain().CurrentBlock().NumberU64())
		require.EqualValues(t, GasLimitDefault, emu.Blockchain().CurrentBlock().Header().GasLimit)
	}

	{
		state, err := emu.Blockchain().State()
		require.NoError(t, err)
		// check the new balances
		require.EqualValues(t, (&big.Int{}).Sub(faucetSupply, transferAmount), state.GetBalance(faucetAddress))
		require.EqualValues(t, transferAmount, state.GetBalance(receiverAddress))
	}
}

func TestBlockchainPersistence(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	defer db.Close()
	testBlockchainPersistence(t, db)
}

func TestBlockchainPersistenceWithKVStoreBackend(t *testing.T) {
	db := rawdb.NewDatabase(NewKVAdapter(dict.New()))
	defer db.Close()
	testBlockchainPersistence(t, db)
}

func testBlockchainPersistence(t testing.TB, db ethdb.Database) {
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

	InitGenesis(DefaultChainID, db, genesisAlloc, GasLimitDefault, 0)

	// do a transfer using one instance of EVMEmulator
	func() {
		emu := NewEVMEmulator(db)
		defer emu.Close()

		sendTransaction(t, emu, faucet, receiverAddress, transferAmount, nil)
	}()

	// initialize a new EVMEmulator using the same DB and check the state
	{
		emu := NewEVMEmulator(db)
		defer emu.Close()

		state, err := emu.Blockchain().State()
		require.NoError(t, err)
		// check the new balances
		require.EqualValues(t, (&big.Int{}).Sub(faucetSupply, transferAmount), state.GetBalance(faucetAddress))
		require.EqualValues(t, transferAmount, state.GetBalance(receiverAddress))
	}
}

type contractFnCaller func(sender *ecdsa.PrivateKey, name string, args ...interface{}) *types.Receipt

func deployEVMContract(t testing.TB, emu *EVMEmulator, creator *ecdsa.PrivateKey, contractABI abi.ABI, contractBytecode []byte, args ...interface{}) (common.Address, contractFnCaller) {
	creatorAddress := crypto.PubkeyToAddress(creator.PublicKey)

	nonce, err := emu.PendingNonceAt(creatorAddress)
	require.NoError(t, err)

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
		types.NewContractCreation(nonce, txValue, gas, GasPrice, data),
		emu.Signer(),
		creator,
	)
	require.NoError(t, err)

	_, err = emu.SendTransaction(tx)
	require.NoError(t, err)
	emu.Commit()

	contractAddress := crypto.CreateAddress(creatorAddress, nonce)

	// assertions
	{
		require.EqualValues(t, 1, emu.Blockchain().CurrentBlock().NumberU64())
		require.EqualValues(t, GasLimitDefault, emu.Blockchain().CurrentBlock().Header().GasLimit)

		// verify contract address
		{
			receipt, err := emu.TransactionReceipt(tx.Hash())
			require.NoError(t, err)
			require.EqualValues(t, contractAddress, receipt.ContractAddress)
			require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
		}

		// verify contract code
		{
			code, err := emu.CodeAt(contractAddress, nil)
			require.NoError(t, err)
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
	db := rawdb.NewMemoryDatabase()
	defer db.Close()
	testStorageContract(t, db)
}

func TestStorageContractWithKVStoreBackend(t *testing.T) {
	db := rawdb.NewDatabase(NewKVAdapter(dict.New()))
	defer db.Close()
	testStorageContract(t, db)
}

func testStorageContract(t testing.TB, db ethdb.Database) {
	// faucet address with initial supply
	faucet, err := crypto.GenerateKey()
	require.NoError(t, err)
	faucetAddress := crypto.PubkeyToAddress(faucet.PublicKey)
	faucetSupply := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))

	genesisAlloc := map[common.Address]core.GenesisAccount{
		faucetAddress: {Balance: faucetSupply},
	}

	InitGenesis(DefaultChainID, db, genesisAlloc, GasLimitDefault, 0)

	emu := NewEVMEmulator(db)
	defer emu.Close()

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

		res, err := emu.CallContract(ethereum.CallMsg{To: &contractAddress, Data: callArguments}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, res)

		var v uint32
		err = contractABI.UnpackIntoInterface(&v, "retrieve", res)
		require.NoError(t, err)
		require.EqualValues(t, 42, v)

		// no state change
		require.EqualValues(t, 1, emu.Blockchain().CurrentBlock().NumberU64())
	}

	// send tx that calls `store(43)`
	{
		callFn(faucet, "store", uint32(43))
		require.EqualValues(t, 2, emu.Blockchain().CurrentBlock().NumberU64())
	}

	// call `retrieve` view again, get 43
	{
		callArguments, err := contractABI.Pack("retrieve")
		require.NoError(t, err)

		res, err := emu.CallContract(ethereum.CallMsg{
			To:   &contractAddress,
			Data: callArguments,
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, res)

		var v uint32
		err = contractABI.UnpackIntoInterface(&v, "retrieve", res)
		require.NoError(t, err)
		require.EqualValues(t, 43, v)

		// no state change
		require.EqualValues(t, 2, emu.Blockchain().CurrentBlock().NumberU64())
	}
}

func TestERC20Contract(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	defer db.Close()
	testERC20Contract(t, db)
}

func TestERC20ContractWithKVStoreBackend(t *testing.T) {
	db := rawdb.NewDatabase(NewKVAdapter(dict.New()))
	defer db.Close()
	testERC20Contract(t, db)
}

func testERC20Contract(t testing.TB, db ethdb.Database) {
	// faucet address with initial supply
	faucet, err := crypto.GenerateKey()
	require.NoError(t, err)
	faucetAddress := crypto.PubkeyToAddress(faucet.PublicKey)
	faucetSupply := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))

	genesisAlloc := map[common.Address]core.GenesisAccount{
		faucetAddress: {Balance: faucetSupply},
	}

	InitGenesis(DefaultChainID, db, genesisAlloc, GasLimitDefault, 0)

	emu := NewEVMEmulator(db)
	defer emu.Close()

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

		res, err := emu.CallContract(ethereum.CallMsg{To: &contractAddress, Data: callArguments}, nil)
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

func initBenchmark(b *testing.B) (*EVMEmulator, []*types.Transaction, dict.Dict) {
	d := dict.New()
	db := rawdb.NewDatabase(NewKVAdapter(d))

	// faucet address with initial supply
	faucet, err := crypto.GenerateKey()
	require.NoError(b, err)
	faucetAddress := crypto.PubkeyToAddress(faucet.PublicKey)
	faucetSupply := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))

	genesisAlloc := map[common.Address]core.GenesisAccount{
		faucetAddress: {Balance: faucetSupply},
	}

	InitGenesis(DefaultChainID, db, genesisAlloc, GasLimitDefault, 0)

	emu := NewEVMEmulator(db)
	defer emu.Close()

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
		senderAddress := crypto.PubkeyToAddress(sender.PublicKey)

		amount := big.NewInt(0)
		nonce := uint64(0)

		callArguments, err := contractABI.Pack("store", uint32(i))
		require.NoError(b, err)

		gas := estimateGas(b, emu, senderAddress, &contractAddress, amount, callArguments)

		txs[i], err = types.SignTx(
			types.NewTransaction(nonce, contractAddress, amount, gas, GasPrice, callArguments),
			emu.Signer(),
			sender,
		)
		require.NoError(b, err)
	}

	return emu, txs, d
}

// benchmarkEVMEmulator is a benchmark for the EVMEmulator that sends an EVM transaction
// that calls `storage.store()`, committing a block every k transactions.
//
// run with: go test -benchmem -cpu=1 -run=' ' -bench='Bench.*'
//
// To generate mem and cpu profiles, add -cpuprofile=cpu.out -memprofile=mem.out
// Then: go tool pprof -http :8080 {cpu,mem}.out
func benchmarkEVMEmulator(b *testing.B, k int) {
	// setup: deploy the storage contract and prepare N transactions to send
	emu, txs, db := initBenchmark(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := emu.SendTransaction(txs[i])
		require.NoError(b, err)

		// commit a block every n txs
		if i%k == 0 {
			emu.Commit()
		}
	}
	emu.Commit()

	b.ReportMetric(dbSize(db)/float64(b.N), "db:bytes/op")
}

func dbSize(db kv.KVStore) float64 {
	r := float64(0)
	db.MustIterate("", func(key kv.Key, value []byte) bool {
		r += float64(len(key) + len(value))
		return true
	})
	return r
}

func BenchmarkEVMEmulator1(b *testing.B)   { benchmarkEVMEmulator(b, 1) }
func BenchmarkEVMEmulator10(b *testing.B)  { benchmarkEVMEmulator(b, 10) }
func BenchmarkEVMEmulator50(b *testing.B)  { benchmarkEVMEmulator(b, 50) }
func BenchmarkEVMEmulator100(b *testing.B) { benchmarkEVMEmulator(b, 100) }
