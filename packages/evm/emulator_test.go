package evm

import (
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
		nonce, err := emu.PendingNonceAt(faucetAddress)
		require.NoError(t, err)

		tx, err := types.SignTx(
			types.NewTransaction(nonce, receiverAddress, transferAmount, 1e6, nil, nil),
			emu.Signer(),
			faucet,
		)
		require.NoError(t, err)

		err = emu.SendTransaction(tx)
		require.NoError(t, err)
		emu.Commit()

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

		{
			nonce, err := emu.PendingNonceAt(faucetAddress)
			require.NoError(t, err)

			tx, err := types.SignTx(
				types.NewTransaction(nonce, receiverAddress, transferAmount, 1e6, nil, nil),
				emu.Signer(),
				faucet,
			)
			require.NoError(t, err)

			err = emu.SendTransaction(tx)
			require.NoError(t, err)
			emu.Commit()
		}
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

func TestContract(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	defer db.Close()
	testContract(t, db)
}

func TestContractWithKVStoreBackend(t *testing.T) {
	db := rawdb.NewDatabase(NewKVAdapter(dict.New()))
	defer db.Close()
	testContract(t, db)
}

func testContract(t *testing.T, db ethdb.Database) {
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

	var contractAddress common.Address

	// contract creation tx
	{
		nonce, err := emu.PendingNonceAt(faucetAddress)
		require.NoError(t, err)

		txValue := big.NewInt(0)

		// initialize number as 42
		constructorArguments, err := contractABI.Pack("", uint32(42))
		require.NoError(t, err)
		require.NotEmpty(t, constructorArguments)

		data := append(evmtest.StorageContractBytecode, constructorArguments...)

		gasPrice, err := emu.SuggestGasPrice()
		require.NoError(t, err)

		gasLimit, err := emu.EstimateGas(ethereum.CallMsg{From: faucetAddress, To: nil, GasPrice: gasPrice, Value: txValue, Data: data})
		require.NoError(t, err)

		tx, err := types.SignTx(
			types.NewContractCreation(nonce, txValue, gasLimit, gasPrice, data),
			emu.Signer(),
			faucet,
		)
		require.NoError(t, err)

		err = emu.SendTransaction(tx)
		require.NoError(t, err)
		emu.Commit()

		// assertions
		{
			require.EqualValues(t, 1, emu.Blockchain().CurrentBlock().NumberU64())
			require.EqualValues(t, GasLimit, emu.Blockchain().CurrentBlock().Header().GasLimit)

			contractAddress = crypto.CreateAddress(faucetAddress, nonce)

			// verify contract address
			{
				receipt, err := emu.TransactionReceipt(tx.Hash())
				require.NoError(t, err)
				require.EqualValues(t, contractAddress, receipt.ContractAddress)
			}

			// verify contract code
			{
				code, err := emu.CodeAt(contractAddress, nil)
				require.NoError(t, err)
				require.NotEmpty(t, code)
			}
		}
	}

	// call `retrieve` view, get 42
	{
		callArguments, err := contractABI.Pack("retrieve")
		require.NoError(t, err)
		require.NotEmpty(t, callArguments)

		res, err := emu.CallContract(ethereum.CallMsg{
			To:   &contractAddress,
			Data: callArguments,
		}, nil)
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
		nonce, err := emu.PendingNonceAt(faucetAddress)
		require.NoError(t, err)

		callArguments, err := contractABI.Pack("store", uint32(43))
		require.NoError(t, err)

		txValue := big.NewInt(0)

		gasPrice, err := emu.SuggestGasPrice()
		require.NoError(t, err)

		gasLimit, err := emu.EstimateGas(ethereum.CallMsg{From: faucetAddress, To: &contractAddress, GasPrice: gasPrice, Value: txValue, Data: callArguments})
		require.NoError(t, err)

		tx, err := types.SignTx(
			types.NewTransaction(nonce, contractAddress, txValue, gasLimit, gasPrice, callArguments),
			emu.Signer(),
			faucet,
		)
		require.NoError(t, err)

		err = emu.SendTransaction(tx)
		require.NoError(t, err)
		emu.Commit()

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
