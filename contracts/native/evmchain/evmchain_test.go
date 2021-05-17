package evmchain

import (
	"bytes"
	"crypto/ecdsa"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

var (
	faucetKey, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	faucetAddress = crypto.PubkeyToAddress(faucetKey.PublicKey)
	faucetSupply  = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))
)

func initEVMChain(t *testing.T) (*solo.Chain, *solo.Solo) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "evmchain", Interface.ProgramHash,
		FieldGenesisAlloc, EncodeGenesisAlloc(map[common.Address]core.GenesisAccount{
			faucetAddress: {Balance: faucetSupply},
		}),
	)
	require.NoError(t, err)
	return chain, env
}

func TestDeploy(t *testing.T) {
	initEVMChain(t)
}

func TestFaucetBalance(t *testing.T) {
	chain, _ := initEVMChain(t)

	ret, err := chain.CallView(Interface.Name, FuncGetBalance, FieldAddress, faucetAddress.Bytes())
	require.NoError(t, err)

	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(FieldBalance))
	require.Zero(t, faucetSupply.Cmp(bal))
}

func getNonceFor(t *testing.T, chain *solo.Chain, addr common.Address) uint64 {
	ret, err := chain.CallView(Interface.Name, FuncGetNonce, FieldAddress, addr.Bytes())
	require.NoError(t, err)

	nonce, ok, err := codec.DecodeUint64(ret.MustGet(FieldResult))
	require.NoError(t, err)
	require.True(t, ok)
	return nonce
}

type contractFnCallerGenerator func(sender *ecdsa.PrivateKey, name string, args ...interface{}) contractFnCaller
type contractFnCaller func(userWallet *ed25519.KeyPair, iotas uint64) (*types.Receipt, uint64, error)

func deployEVMContract(t *testing.T, chain *solo.Chain, env *solo.Solo, creator *ecdsa.PrivateKey, contractABI abi.ABI, contractBytecode []byte, args ...interface{}) (common.Address, contractFnCallerGenerator) {
	creatorAddress := crypto.PubkeyToAddress(creator.PublicKey)

	nonce := getNonceFor(t, chain, creatorAddress)

	// initialize number as 42
	constructorArguments, err := contractABI.Pack("", args...)
	require.NoError(t, err)

	data := append(contractBytecode, constructorArguments...)

	tx, err := types.SignTx(
		types.NewContractCreation(nonce, big.NewInt(0), evm.GasLimit, evm.GasPrice, data),
		evm.Signer(),
		faucetKey,
	)
	require.NoError(t, err)

	txdata, err := tx.MarshalBinary()
	require.NoError(t, err)

	txDataBlobHash, err := chain.UploadBlobOptimized(1024, nil, FieldTransactionData, txdata)
	require.NoError(t, err)

	_, err = chain.PostRequestSync(
		solo.NewCallParams(Interface.Name, FuncSendTransaction,
			FieldTransactionDataBlobHash, codec.EncodeHashValue(txDataBlobHash),
		).WithIotas(100000),
		nil,
	)
	require.NoError(t, err)

	contractAddress := crypto.CreateAddress(creatorAddress, nonce)

	callFn := func(sender *ecdsa.PrivateKey, name string, args ...interface{}) contractFnCaller {
		senderAddress := crypto.PubkeyToAddress(sender.PublicKey)

		nonce := getNonceFor(t, chain, senderAddress)

		callArguments, err := contractABI.Pack(name, args...)
		require.NoError(t, err)

		tx, err := types.SignTx(
			types.NewTransaction(nonce, contractAddress, big.NewInt(0), evm.GasLimit, evm.GasPrice, callArguments),
			evm.Signer(),
			faucetKey,
		)
		require.NoError(t, err)

		txdata, err := tx.MarshalBinary()
		require.NoError(t, err)

		return func(userWallet *ed25519.KeyPair, iotas uint64) (*types.Receipt, uint64, error) {
			if userWallet == nil {
				//create new user account
				userWallet, _ = env.NewKeyPairWithFunds()
			}
			result, err := chain.PostRequestSync(
				solo.NewCallParams(Interface.Name, FuncSendTransaction, FieldTransactionData, txdata).
					WithIotas(iotas),
				userWallet,
			)

			if err != nil {
				return nil, 0, err
			}

			var receipt *types.Receipt
			err = rlp.DecodeBytes(result.MustGet(FieldResult), &receipt)
			require.NoError(t, err)

			gasFee, _, err := codec.DecodeUint64(result.MustGet(FieldGasFee))
			require.NoError(t, err)

			return receipt, gasFee, nil
		}

	}

	return contractAddress, callFn
}

// helper to reuse code to call the `retrieve` view in the storage contract
func getCallRetrieveView(t *testing.T, chain *solo.Chain, contractAddress common.Address, contractABI abi.ABI) func() uint32 {
	return func() uint32 {
		callArguments, err := contractABI.Pack("retrieve")
		require.NoError(t, err)

		ret, err := chain.CallView(Interface.Name, FuncCallView,
			FieldAddress, contractAddress.Bytes(),
			FieldCallArguments, callArguments,
		)
		require.NoError(t, err)

		var v uint32
		err = contractABI.UnpackIntoInterface(&v, "retrieve", ret.MustGet(FieldResult))
		require.NoError(t, err)
		return v
	}
}

func TestStorageContract(t *testing.T) {
	chain, env := initEVMChain(t)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)

	// deploy solidity `storage` contract
	contractAddress, callFn := deployEVMContract(t, chain, env, faucetKey, contractABI, evmtest.StorageContractBytecode, uint32(42))

	retrieve := getCallRetrieveView(t, chain, contractAddress, contractABI)

	// call evmchain's FuncCallView to call EVM contract's `retrieve` view, get 42
	require.EqualValues(t, 42, retrieve())

	// call FuncSendTransaction with EVM tx that calls `store(43)`
	_, _, err = callFn(faucetKey, "store", uint32(43))(nil, 100000)
	require.NoError(t, err)

	// call `retrieve` view, get 43
	require.EqualValues(t, 43, retrieve())
}

func TestERC20Contract(t *testing.T) {
	chain, env := initEVMChain(t)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.ERC20ContractABI))
	require.NoError(t, err)

	// deploy solidity `erc20` contract
	contractAddress, callFn := deployEVMContract(t, chain, env, faucetKey, contractABI, evmtest.ERC20ContractBytecode, "TestCoin", "TEST")

	callIntViewFn := func(name string, args ...interface{}) *big.Int {
		callArguments, err := contractABI.Pack(name, args...)
		require.NoError(t, err)

		ret, err := chain.CallView(Interface.Name, FuncCallView,
			FieldAddress, contractAddress.Bytes(),
			FieldCallArguments, callArguments,
		)
		require.NoError(t, err)

		v := new(big.Int)
		err = contractABI.UnpackIntoInterface(&v, name, ret.MustGet(FieldResult))
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
	receipt, _, err := callFn(faucetKey, "transfer", recipientAddress, transferAmount)(nil, 100000)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	require.Equal(t, 1, len(receipt.Logs))

	// call `balanceOf` view => check balance of recipient = 1337 TestCoin
	require.Zero(t, callIntViewFn("balanceOf", recipientAddress).Cmp(transferAmount))
}

func TestGetCode(t *testing.T) {
	chain, env := initEVMChain(t)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.ERC20ContractABI))
	require.NoError(t, err)

	// deploy solidity `erc20` contract
	contractAddress, _ := deployEVMContract(t, chain, env, faucetKey, contractABI, evmtest.ERC20ContractBytecode, "TestCoin", "TEST")

	// get contract bytecode from EVM emulator
	ret, err := chain.CallView(Interface.Name, FuncGetCode, FieldAddress, contractAddress.Bytes())
	require.NoError(t, err)
	retrievedBytecode := ret.MustGet(FieldResult)

	//ensure returned bytecode matches the expected runtime bytecode
	require.True(t, bytes.Equal(retrievedBytecode, evmtest.ERC20ContractRuntimeBytecode), "bytecode retrieved from the chain must match the deployed bytecode")
}

func TestGasCharged(t *testing.T) {
	chain, env := initEVMChain(t)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)

	// deploy solidity `storage` contract
	contractAddress, callFn := deployEVMContract(t, chain, env, faucetKey, contractABI, evmtest.StorageContractBytecode, uint32(42))

	retrieve := getCallRetrieveView(t, chain, contractAddress, contractABI)

	// call evmchain's FuncCallView to call EVM contract's `retrieve` view, get 42
	require.EqualValues(t, 42, retrieve())

	userWallet, userAddress := env.NewKeyPairWithFunds()
	var initialBalance uint64 = env.GetAddressBalance(userAddress, ledgerstate.ColorIOTA)

	// call `store(999)` with enough gas
	receipt, gasFee, err := callFn(faucetKey, "store", uint32(999))(userWallet, initialBalance)
	require.NoError(t, err)
	require.Greater(t, gasFee, uint64(0))

	println(receipt) // TODO CHECK IF RECEIPT IS OKAY ????????!!!

	// call `retrieve` view, get 999
	require.EqualValues(t, 999, retrieve())

	// user on-chain account is credited with excess iotas (iotasSent - gasUsed)
	expectedUserBalance := initialBalance - gasFee

	var newBalance uint64 = env.GetAddressBalance(userAddress, ledgerstate.ColorIOTA)
	println(newBalance) // ????????????????????????????????????????????????????????????????????/

	env.AssertAddressIotas(userAddress, expectedUserBalance)

	// call `store(123)` without enough gas
	_, gasFee, err = callFn(faucetKey, "store", uint32(123))(userWallet, 1)
	require.Greater(t, gasFee, 0)
	require.Error(t, err)

	// call `retrieve` view, get 999 - which means store(123) failed and the previous state is kept
	require.EqualValues(t, 999, retrieve())

	// verify user on-chain account still has the same balance
	env.AssertAddressIotas(userAddress, expectedUserBalance)
}

// TODO check infinite loop gets stopped by gas used
