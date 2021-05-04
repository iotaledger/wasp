package evmchain

import (
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

func TestContract(t *testing.T) {
	chain := initEVMChain(t)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)

	var contractAddress common.Address

	// deploy solidity contract
	{
		nonce := uint64(0) // TODO: add getNonce endpoint?

		txValue := big.NewInt(0)

		// initialize number as 42
		constructorArguments, err := contractABI.Pack("", uint32(42))
		require.NoError(t, err)

		data := append(evmtest.StorageContractBytecode, constructorArguments...)

		tx, err := types.SignTx(
			types.NewContractCreation(nonce, txValue, evm.GasLimit, evm.GasPrice, data),
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

		contractAddress = crypto.CreateAddress(faucetAddress, nonce)
	}

	// call evmchain's FuncCallView to call EVM contract's `retrieve` view, get 42
	{
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
		require.EqualValues(t, 42, v)
	}

	// call FuncSendTransaction with EVM tx that calls `store(43)`
	{
		nonce := uint64(1) // TODO: add getNonce endpoint?

		callArguments, err := contractABI.Pack("store", uint32(43))
		require.NoError(t, err)

		txValue := big.NewInt(0)

		tx, err := types.SignTx(
			types.NewTransaction(nonce, contractAddress, txValue, evm.GasLimit, evm.GasPrice, callArguments),
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
	}

	// call `retrieve` view, get 43
	{
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
		require.EqualValues(t, 43, v)
	}
}
