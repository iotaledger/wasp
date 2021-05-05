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
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/stretchr/testify/require"
)

func sendTransaction(t *testing.T, emu *EVMEmulator, sender *ecdsa.PrivateKey, receiverAddress common.Address, amount *big.Int, data []byte) {
	senderAddress := crypto.PubkeyToAddress(sender.PublicKey)

	nonce, err := emu.PendingNonceAt(senderAddress)
	require.NoError(t, err)

	tx, err := types.SignTx(
		types.NewTransaction(nonce, receiverAddress, amount, GasLimit, GasPrice, data),
		emu.Signer(),
		sender,
	)
	require.NoError(t, err)

	err = emu.SendTransaction(tx)
	require.NoError(t, err)
	emu.Commit()
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

func testBlockchain(t *testing.T, db ethdb.Database) {
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

	InitGenesis(db, genesisAlloc)

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
			require.EqualValues(t, GasLimit, block.Header().GasLimit)
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
		require.EqualValues(t, GasLimit, emu.Blockchain().CurrentBlock().Header().GasLimit)
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

func testBlockchainPersistence(t *testing.T, db ethdb.Database) {
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

	InitGenesis(db, genesisAlloc)

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

func deployEVMContract(t *testing.T, emu *EVMEmulator, creator *ecdsa.PrivateKey, contractABI abi.ABI, contractBytecode []byte, args ...interface{}) (common.Address, func(sender *ecdsa.PrivateKey, name string, args ...interface{})) {
	creatorAddress := crypto.PubkeyToAddress(creator.PublicKey)

	nonce, err := emu.PendingNonceAt(creatorAddress)
	require.NoError(t, err)

	txValue := big.NewInt(0)

	// initialize number as 42
	constructorArguments, err := contractABI.Pack("", args...)
	require.NoError(t, err)
	require.NotEmpty(t, constructorArguments)

	data := append(contractBytecode, constructorArguments...)

	tx, err := types.SignTx(
		types.NewContractCreation(nonce, txValue, GasLimit, GasPrice, data),
		emu.Signer(),
		creator,
	)
	require.NoError(t, err)

	err = emu.SendTransaction(tx)
	require.NoError(t, err)
	emu.Commit()

	contractAddress := crypto.CreateAddress(creatorAddress, nonce)

	// assertions
	{
		require.EqualValues(t, 1, emu.Blockchain().CurrentBlock().NumberU64())
		require.EqualValues(t, GasLimit, emu.Blockchain().CurrentBlock().Header().GasLimit)

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

	callFn := func(sender *ecdsa.PrivateKey, name string, args ...interface{}) {
		callArguments, err := contractABI.Pack(name, args...)
		require.NoError(t, err)
		sendTransaction(t, emu, sender, contractAddress, big.NewInt(0), callArguments)
	}

	return contractAddress, callFn
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

func testStorageContract(t *testing.T, db ethdb.Database) {
	// faucet address with initial supply
	faucet, err := crypto.GenerateKey()
	require.NoError(t, err)
	faucetAddress := crypto.PubkeyToAddress(faucet.PublicKey)
	faucetSupply := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))

	genesisAlloc := map[common.Address]core.GenesisAccount{
		faucetAddress: {Balance: faucetSupply},
	}

	InitGenesis(db, genesisAlloc)

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

func testERC20Contract(t *testing.T, db ethdb.Database) {
	// faucet address with initial supply
	faucet, err := crypto.GenerateKey()
	require.NoError(t, err)
	faucetAddress := crypto.PubkeyToAddress(faucet.PublicKey)
	faucetSupply := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))

	genesisAlloc := map[common.Address]core.GenesisAccount{
		faucetAddress: {Balance: faucetSupply},
	}

	InitGenesis(db, genesisAlloc)

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
	callFn(erc20Owner, "transfer", recipientAddress, transferAmount)

	// call `balanceOf` view => check balance of recipient = 1337 TestCoin
	require.Zero(t, callIntViewFn("balanceOf", recipientAddress).Cmp(transferAmount))

	// call `transferFrom` as recipient without allowance => get error
	callFn(recipient, "transferFrom", erc20OwnerAddress, recipientAddress, transferAmount)

	// call `balanceOf` view => check balance of recipient = 1337 TestCoin
	require.Zero(t, callIntViewFn("balanceOf", recipientAddress).Cmp(transferAmount))

	// call `approve` as erc20Owner
	callFn(erc20Owner, "approve", recipientAddress, transferAmount)

	// call `transferFrom` as recipient with allowance => ok
	callFn(recipient, "transferFrom", erc20OwnerAddress, recipientAddress, transferAmount)

	// call `balanceOf` view => check balance of recipient = 2 * 1337 TestCoin
	require.Zero(t, callIntViewFn("balanceOf", recipientAddress).Cmp(new(big.Int).Mul(transferAmount, big.NewInt(2))))
}
