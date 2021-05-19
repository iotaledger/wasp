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

func initEVMChain(t *testing.T) *solo.Chain {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "evmchain", Interface.ProgramHash,
		FieldGenesisAlloc, EncodeGenesisAlloc(map[common.Address]core.GenesisAccount{
			faucetAddress: {Balance: faucetSupply},
		}),
	)
	require.NoError(t, err)
	return chain
}

func TestDeploy(t *testing.T) {
	initEVMChain(t)
}

func TestFaucetBalance(t *testing.T) {
	chain := initEVMChain(t)

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

type contractFnCaller func(sender *ecdsa.PrivateKey, name string, args ...interface{}) *types.Receipt

func deployEVMContract(t *testing.T, chain *solo.Chain, creator *ecdsa.PrivateKey, contractABI abi.ABI, contractBytecode []byte, args ...interface{}) (common.Address, contractFnCaller) {
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

	req, toUpload := solo.NewCallParamsOptimized(Interface.Name, FuncSendTransaction, 1024,
		FieldTransactionData, txdata,
	)
	req.WithIotas(1)
	for _, v := range toUpload {
		chain.Env.PutBlobDataIntoRegistry(v)
	}

	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	contractAddress := crypto.CreateAddress(creatorAddress, nonce)

	callFn := func(sender *ecdsa.PrivateKey, name string, args ...interface{}) *types.Receipt {
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

		_, err = chain.PostRequestSync(
			solo.NewCallParams(Interface.Name, FuncSendTransaction, FieldTransactionData, txdata).
				WithIotas(1),
			nil,
		)
		require.NoError(t, err)

		var receipt *evmchain.Receipt
		{
			ret, err := chain.CallView(Interface.Name, FuncGetReceipt,
				FieldTransactionHash, tx.Hash().Bytes(),
			)
			require.NoError(t, err)

			err = evmchain.DecodeReceipt(ret.MustGet(FieldResult), &receipt)
			require.NoError(t, err)
		}
		return receipt
	}

	return contractAddress, callFn
}

func TestStorageContract(t *testing.T) {
	chain := initEVMChain(t)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)

	// deploy solidity `storage` contract
	contractAddress, callFn := deployEVMContract(t, chain, faucetKey, contractABI, evmtest.StorageContractBytecode, uint32(42))

	retrieve := func() uint32 {
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

	// call evmchain's FuncCallView to call EVM contract's `retrieve` view, get 42
	require.EqualValues(t, 42, retrieve())

	// call FuncSendTransaction with EVM tx that calls `store(43)`
	callFn(faucetKey, "store", uint32(43))

	// call `retrieve` view, get 43
	require.EqualValues(t, 43, retrieve())
}

func TestERC20Contract(t *testing.T) {
	chain := initEVMChain(t)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.ERC20ContractABI))
	require.NoError(t, err)

	// deploy solidity `erc20` contract
	contractAddress, callFn := deployEVMContract(t, chain, faucetKey, contractABI, evmtest.ERC20ContractBytecode, "TestCoin", "TEST")

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
	receipt := callFn(faucetKey, "transfer", recipientAddress, transferAmount)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	require.Equal(t, 1, len(receipt.Logs))

	// call `balanceOf` view => check balance of recipient = 1337 TestCoin
	require.Zero(t, callIntViewFn("balanceOf", recipientAddress).Cmp(transferAmount))
}

func TestGetCode(t *testing.T) {
	chain := initEVMChain(t)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.ERC20ContractABI))
	require.NoError(t, err)

	// deploy solidity `erc20` contract
	contractAddress, _ := deployEVMContract(t, chain, faucetKey, contractABI, evmtest.ERC20ContractBytecode, "TestCoin", "TEST")

	// get contract bytecode from EVM emulator
	ret, err := chain.CallView(Interface.Name, FuncGetCode, FieldAddress, contractAddress.Bytes())
	require.NoError(t, err)
	retrievedBytecode := ret.MustGet(FieldResult)

	// ensure returned bytecode matches the expected runtime bytecode
	require.True(t, bytes.Equal(retrievedBytecode, evmtest.ERC20ContractRuntimeBytecode), "bytecode retrieved from the chain must match the deployed bytecode")
}
