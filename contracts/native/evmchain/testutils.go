package evmchain

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
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

var (
	TestFaucetKey, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	TestFaucetAddress = crypto.PubkeyToAddress(TestFaucetKey.PublicKey)
	TestFaucetSupply  = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))
)

func InitEVMChain(t *testing.T) (*solo.Chain, *solo.Solo) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "evmchain", Interface.ProgramHash,
		FieldGenesisAlloc, EncodeGenesisAlloc(map[common.Address]core.GenesisAccount{
			TestFaucetAddress: {Balance: TestFaucetSupply},
		}),
	)
	require.NoError(t, err)
	return chain, env
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
type contractFnCaller func(userWallet *ed25519.KeyPair, iotas uint64) (*Receipt, uint64, error)
type contractFnCallerWithGasLimit func(userWallet *ed25519.KeyPair, iotas uint64, gasLimit uint64) (*Receipt, uint64, error)

func DeployEVMContract(t *testing.T, chain *solo.Chain, env *solo.Solo, creator *ecdsa.PrivateKey, contractABI abi.ABI, contractBytecode []byte, args ...interface{}) (common.Address, contractFnCallerGenerator) {
	creatorAddress := crypto.PubkeyToAddress(creator.PublicKey)

	nonce := getNonceFor(t, chain, creatorAddress)

	constructorArguments, err := contractABI.Pack("", args...)
	require.NoError(t, err)

	data := append(contractBytecode, constructorArguments...)

	gaslimit := uint64(GetGasPerIotas(t, chain)) * uint64(utxodb.RequestFundsAmount)

	tx, err := types.SignTx(
		types.NewContractCreation(nonce, big.NewInt(0), gaslimit, evm.GasPrice, data),
		evm.Signer(),
		TestFaucetKey,
	)
	require.NoError(t, err)

	txdata, err := tx.MarshalBinary()
	require.NoError(t, err)

	req, toUpload := solo.NewCallParamsOptimized(Interface.Name, FuncSendTransaction, 1024,
		FieldTransactionData, txdata,
	)
	req.WithIotas(utxodb.RequestFundsAmount)
	for _, v := range toUpload {
		chain.Env.PutBlobDataIntoRegistry(v)
	}

	deployerWallet, _ := env.NewKeyPairWithFunds()

	_, err = chain.PostRequestSync(req, deployerWallet)

	require.NoError(t, err)

	contractAddress := crypto.CreateAddress(creatorAddress, nonce)

	callFn := func(sender *ecdsa.PrivateKey, name string, args ...interface{}) contractFnCaller {
		callWithGasLimit := createCallFnWithGasLimit(t, chain, env, contractABI, contractAddress, sender, name, args...)

		return func(userWallet *ed25519.KeyPair, iotas uint64) (*Receipt, uint64, error) {
			gaslimit := uint64(GetGasPerIotas(t, chain)) * iotas
			return callWithGasLimit(userWallet, iotas, gaslimit)
		}
	}

	return contractAddress, callFn
}

func createCallFnWithGasLimit(t *testing.T, chain *solo.Chain, env *solo.Solo, contractABI abi.ABI, contractAddress common.Address, sender *ecdsa.PrivateKey, name string, args ...interface{}) contractFnCallerWithGasLimit {
	senderAddress := crypto.PubkeyToAddress(sender.PublicKey)

	nonce := getNonceFor(t, chain, senderAddress)
	if nonce == 0 {
		nonce = 1
	}

	callArguments, err := contractABI.Pack(name, args...)
	require.NoError(t, err)

	return func(userWallet *ed25519.KeyPair, iotas uint64, gaslimit uint64) (*Receipt, uint64, error) {

		unsignedTx := types.NewTransaction(nonce, contractAddress, big.NewInt(0), gaslimit, evm.GasPrice, callArguments)

		tx, err := types.SignTx(
			unsignedTx,
			evm.Signer(),
			TestFaucetKey,
		)
		require.NoError(t, err)

		txdata, err := tx.MarshalBinary()
		require.NoError(t, err)
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

		gasFee, _, err := codec.DecodeUint64(result.MustGet(FieldGasFee))
		require.NoError(t, err)

		receipt, err := DecodeReceipt(result.MustGet(FieldResult))
		require.NoError(t, err)

		return receipt, gasFee, nil
	}
}

func callStorageRetrieve(t *testing.T, chain *solo.Chain, contractAddress common.Address) uint32 {
	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)
	callArguments, err := contractABI.Pack("retrieve")
	require.NoError(t, err)

	ret, err := chain.CallView(Interface.Name, FuncCallContract, FieldCallMsg, EncodeCallMsg(ethereum.CallMsg{
		From: TestFaucetAddress,
		To:   &contractAddress,
		Data: callArguments,
	}))
	require.NoError(t, err)

	var v uint32
	err = contractABI.UnpackIntoInterface(&v, "retrieve", ret.MustGet(FieldResult))
	require.NoError(t, err)
	return v
}

func GetGasPerIotas(t *testing.T, chain *solo.Chain) int64 {
	ret, err := chain.CallView(Interface.Name, FuncGetGasPerIota)
	require.NoError(t, err)
	gasPerIotas, _, err := codec.DecodeInt64(ret.MustGet(FieldResult))
	require.NoError(t, err)
	return gasPerIotas
}
