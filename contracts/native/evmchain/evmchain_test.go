package evmchain

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

func TestDeploy(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "evmchain", Interface.ProgramHash)
	require.NoError(t, err)
}

func TestFaucetBalance(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "evmchain", Interface.ProgramHash)
	require.NoError(t, err)

	ret, err := chain.CallView(Interface.Name, FuncGetBalance, FieldAddress, FaucetAddress.Bytes())
	require.NoError(t, err)

	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(FieldBalance))
	require.Zero(t, FaucetSupply.Cmp(bal))
}

func TestContract(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "evmchain", Interface.ProgramHash)
	require.NoError(t, err)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)

	var contractAddress common.Address

	gasPrice := big.NewInt(1)
	gasLimit := evm.GasLimit

	// deploy solidity contract
	{
		nonce := uint64(0) // TODO: add getNonce endpoint?

		txValue := big.NewInt(0)

		// initialize number as 42
		constructorArguments, err := contractABI.Pack("", uint32(42))
		require.NoError(t, err)

		data := append(evmtest.StorageContractBytecode, constructorArguments...)

		tx, err := types.SignTx(
			types.NewContractCreation(nonce, txValue, gasLimit, gasPrice, data),
			evm.Signer(),
			FaucetKey,
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

		contractAddress = crypto.CreateAddress(FaucetAddress, nonce)
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
			types.NewTransaction(nonce, contractAddress, txValue, gasLimit, gasPrice, callArguments),
			evm.Signer(),
			FaucetKey,
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
