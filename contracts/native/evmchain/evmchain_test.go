package evmchain

import (
	"bytes"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

func TestDeploy(t *testing.T) {
	InitEVMChain(t)
}

func TestFaucetBalance(t *testing.T) {
	chain, _ := InitEVMChain(t)

	ret, err := chain.CallView(Interface.Name, FuncGetBalance, FieldAddress, FaucetAddress.Bytes())
	require.NoError(t, err)

	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(FieldBalance))
	require.Zero(t, FaucetSupply.Cmp(bal))
}

func TestStorageContract(t *testing.T) {
	chain, env := InitEVMChain(t)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)

	// deploy solidity `storage` contract
	contractAddress, callFn := DeployEVMContract(t, chain, env, FaucetKey, contractABI, evmtest.StorageContractBytecode, uint32(42))

	retrieve := GetCallRetrieveView(t, chain, contractAddress, contractABI)

	// call evmchain's FuncCallView to call EVM contract's `retrieve` view, get 42
	require.EqualValues(t, 42, retrieve())

	// call FuncSendTransaction with EVM tx that calls `store(43)`
	_, _, err = callFn(FaucetKey, "store", uint32(43))(nil, 100000)
	require.NoError(t, err)

	// call `retrieve` view, get 43
	require.EqualValues(t, 43, retrieve())
}

func TestERC20Contract(t *testing.T) {
	chain, env := InitEVMChain(t)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.ERC20ContractABI))
	require.NoError(t, err)

	// deploy solidity `erc20` contract
	contractAddress, callFn := DeployEVMContract(t, chain, env, FaucetKey, contractABI, evmtest.ERC20ContractBytecode, "TestCoin", "TEST")

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
	receipt, _, err := callFn(FaucetKey, "transfer", recipientAddress, transferAmount)(nil, 100000)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	require.Equal(t, 1, len(receipt.Logs))

	// call `balanceOf` view => check balance of recipient = 1337 TestCoin
	require.Zero(t, callIntViewFn("balanceOf", recipientAddress).Cmp(transferAmount))
}

func TestGetCode(t *testing.T) {
	chain, env := InitEVMChain(t)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.ERC20ContractABI))
	require.NoError(t, err)

	// deploy solidity `erc20` contract
	contractAddress, _ := DeployEVMContract(t, chain, env, FaucetKey, contractABI, evmtest.ERC20ContractBytecode, "TestCoin", "TEST")

	// get contract bytecode from EVM emulator
	ret, err := chain.CallView(Interface.Name, FuncGetCode, FieldAddress, contractAddress.Bytes())
	require.NoError(t, err)
	retrievedBytecode := ret.MustGet(FieldResult)

	//ensure returned bytecode matches the expected runtime bytecode
	require.True(t, bytes.Equal(retrievedBytecode, evmtest.ERC20ContractRuntimeBytecode), "bytecode retrieved from the chain must match the deployed bytecode")
}

func TestGasCharged(t *testing.T) {
	chain, env := InitEVMChain(t)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)

	// deploy solidity `storage` contract
	contractAddress, callFn := DeployEVMContract(t, chain, env, FaucetKey, contractABI, evmtest.StorageContractBytecode, uint32(42))

	retrieve := GetCallRetrieveView(t, chain, contractAddress, contractABI)

	// call evmchain's FuncCallView to call EVM contract's `retrieve` view, get 42
	require.EqualValues(t, 42, retrieve())

	userWallet, userAddress := env.NewKeyPairWithFunds()
	userAgentID := coretypes.NewAgentID(userAddress, 0)

	var initialBalance uint64 = env.GetAddressBalance(userAddress, ledgerstate.ColorIOTA)
	var iotasSent uint64 = initialBalance - 1

	// call `store(999)` with enough gas
	receipt, gasFee, err := callFn(FaucetKey, "store", uint32(42))(userWallet, iotasSent)
	require.NoError(t, err)
	require.Greater(t, gasFee, uint64(0))

	println(receipt) // TODO CHECK Why RECEIPT IS BROKEN

	// call `retrieve` view, get 42
	require.EqualValues(t, 42, retrieve())

	// user on-chain account is credited with excess iotas (iotasSent - gasUsed)
	expectedUserBalance := iotasSent - gasFee

	chain.AssertIotas(userAgentID, expectedUserBalance)

	// call `store(123)` without enough gas
	_, _, err = callFn(FaucetKey, "store", uint32(123))(userWallet, 1)
	require.Error(t, err)

	// call `retrieve` view, get 999 - which means store(123) failed and the previous state is kept
	require.EqualValues(t, 42, retrieve())

	// verify user on-chain account still has the same balance
	chain.AssertIotas(userAgentID, expectedUserBalance)
}

func getEvmOwner(t *testing.T, chain *solo.Chain) coretypes.AgentID {
	ret, err := chain.CallView(Interface.Name, FuncGetOwner)
	require.NoError(t, err)
	owner, _, err := codec.DecodeAgentID(ret.MustGet(FieldResult))
	require.NoError(t, err)
	return owner
}

func TestOwner(t *testing.T) {
	chain, env := InitEVMChain(t)

	// the default owner is correct
	owner := getEvmOwner(t, chain)
	require.True(t, owner.Equals(&chain.OriginatorAgentID))

	// only the owner can call the setOwner endpoint
	newUserWallet, newUserAddress := env.NewKeyPairWithFunds()
	newUserAgentId := coretypes.NewAgentID(newUserAddress, 0)
	_, err := chain.PostRequestSync(
		solo.NewCallParams(Interface.Name, FuncSetOwner, FieldEvmOwner, newUserAgentId).
			WithIotas(100000),
		newUserWallet,
	)
	require.Error(t, err)

	// ensure owner didn't change after a failed call
	owner = getEvmOwner(t, chain)
	require.True(t, owner.Equals(&chain.OriginatorAgentID))

	// current owner is able to set a new owner
	_, err = chain.PostRequestSync(
		solo.NewCallParams(Interface.Name, FuncSetOwner, FieldEvmOwner, newUserAgentId).
			WithIotas(100000),
		chain.OriginatorKeyPair,
	)
	require.NoError(t, err)

	owner = getEvmOwner(t, chain)
	require.True(t, owner.Equals(newUserAgentId))
}

func getGasPerIotas(t *testing.T, chain *solo.Chain) int64 {
	ret, err := chain.CallView(Interface.Name, FuncGetGasPerIota)
	require.NoError(t, err)
	gasPerIotas, _, err := codec.DecodeInt64(ret.MustGet(FieldResult))
	require.NoError(t, err)
	return gasPerIotas
}

func TestGasPerIotas(t *testing.T) {
	chain, env := InitEVMChain(t)

	//deploy storage contract to test the gas pricing
	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)
	_, callFn := DeployEVMContract(t, chain, env, FaucetKey, contractABI, evmtest.StorageContractBytecode, uint32(42))

	// the default value is correct
	gasPerIotas := getGasPerIotas(t, chain)
	require.Equal(t, gasPerIotas, DefaultGasPerIota)

	_, gasFee, err := callFn(FaucetKey, "store", uint32(43))(nil, 100000)
	require.NoError(t, err)
	initialGasFee := gasFee

	// only the owner can call the setGasPerIotas endpoint
	newGasPerIota := DefaultGasPerIota * 1000
	newUserWallet, _ := env.NewKeyPairWithFunds()
	_, err = chain.PostRequestSync(
		solo.NewCallParams(Interface.Name, FuncSetGasPerIota, FieldGasPerIota, newGasPerIota).
			WithIotas(100000),
		newUserWallet,
	)
	require.Error(t, err)

	// ensure gasPerIotas didn't change after a failed call
	gasPerIotas = getGasPerIotas(t, chain)
	require.Equal(t, gasPerIotas, DefaultGasPerIota)

	// current owner is able to set a new gasPerIotas
	_, err = chain.PostRequestSync(
		solo.NewCallParams(Interface.Name, FuncSetGasPerIota, FieldGasPerIota, newGasPerIota).
			WithIotas(100000),
		chain.OriginatorKeyPair,
	)
	require.NoError(t, err)

	gasPerIotas = getGasPerIotas(t, chain)
	require.Equal(t, gasPerIotas, newGasPerIota)

	// run an equivalent request and compare the gas fees
	_, gasFee, err = callFn(FaucetKey, "store", uint32(44))(nil, 100000)
	require.NoError(t, err)
	require.Less(t, gasFee, initialGasFee)
}

func postWithdrawalFeesReq(t *testing.T, chain *solo.Chain, wallet *ed25519.KeyPair) (dict.Dict, error) {
	return chain.PostRequestSync(
		solo.NewCallParams(Interface.Name, FuncWithdrawGasFees).
			WithIotas(1),
		wallet,
	)
}

func TestWithdrawalOwnerFees(t *testing.T) {
	chain, env := InitEVMChain(t)

	//deploy storage contract to test gas fees collected
	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)
	_, callFn := DeployEVMContract(t, chain, env, FaucetKey, contractABI, evmtest.StorageContractBytecode, uint32(42))

	// only the owner can call withdrawal
	user1Wallet, user1Address := env.NewKeyPairWithFunds()
	user1AgentID := coretypes.NewAgentID(user1Address, 0)

	_, err = postWithdrawalFeesReq(t, chain, user1Wallet)
	require.Error(t, err)

	// change owner to user1
	_, err = chain.PostRequestSync(
		solo.NewCallParams(Interface.Name, FuncSetOwner, FieldEvmOwner, user1AgentID).
			WithIotas(1),
		chain.OriginatorKeyPair,
	)
	require.NoError(t, err)

	// collect fees from contract deployment
	user1Balance0 := env.GetAddressBalance(user1Address, ledgerstate.ColorIOTA)

	_, err = postWithdrawalFeesReq(t, chain, user1Wallet)
	require.NoError(t, err)
	user1Balance1 := env.GetAddressBalance(user1Address, ledgerstate.ColorIOTA)

	require.Greater(t, user1Balance1, user1Balance0)

	// collect fees from a SC call, check that the collected fees matches the fees charged
	user1Balance2 := env.GetAddressBalance(user1Address, ledgerstate.ColorIOTA)
	_, chargedGasFee, err := callFn(FaucetKey, "store", uint32(43))(nil, 100000)
	require.NoError(t, err)

	_, err = postWithdrawalFeesReq(t, chain, user1Wallet)
	user1Balance3 := env.GetAddressBalance(user1Address, ledgerstate.ColorIOTA)

	require.NoError(t, err)
	require.Equal(t, user1Balance3, user1Balance2+chargedGasFee-1) // -1 because we sent 1 iota when requesting the funds withdrawal

	//try to withdrawal a second time, it should succeed, but owner balance shouldnt not change (there are no fees to withdraw)
	_, err = postWithdrawalFeesReq(t, chain, user1Wallet)
	require.NoError(t, err)
	user1Balance4 := env.GetAddressBalance(user1Address, ledgerstate.ColorIOTA)
	require.Equal(t, user1Balance3-1, user1Balance4)

	//try to withdrawal fees to another actor using using the FieldAgentId param
	_, chargedGasFee, err = callFn(FaucetKey, "store", uint32(44))(nil, 100000)
	require.NoError(t, err)

	_, user2Address := env.NewKeyPairWithFunds()
	user2AgentID := coretypes.NewAgentID(user2Address, 0)
	user2Balance0 := env.GetAddressBalance(user2Address, ledgerstate.ColorIOTA)
	_, err = chain.PostRequestSync(
		solo.NewCallParams(Interface.Name, FuncWithdrawGasFees, FieldAgentId, user2AgentID).WithIotas(1),
		user1Wallet,
	)
	require.NoError(t, err)
	user2Balance1 := env.GetAddressBalance(user2Address, ledgerstate.ColorIOTA)
	require.Equal(t, user2Balance1, user2Balance0+chargedGasFee)
}

func TestGasLimit(t *testing.T) {
	// TODO test that the gas used scales correctly with the amount of iotas sent

}
