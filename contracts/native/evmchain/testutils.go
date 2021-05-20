package evmchain

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/evm"
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

func DeployEVMContract(t *testing.T, chain *solo.Chain, env *solo.Solo, creator *ecdsa.PrivateKey, contractABI abi.ABI, contractBytecode []byte, args ...interface{}) (common.Address, contractFnCallerGenerator) {
	creatorAddress := crypto.PubkeyToAddress(creator.PublicKey)

	nonce := getNonceFor(t, chain, creatorAddress)

	// initialize number as 42
	constructorArguments, err := contractABI.Pack("", args...)
	require.NoError(t, err)

	data := append(contractBytecode, constructorArguments...)

	tx, err := types.SignTx(
		types.NewContractCreation(nonce, big.NewInt(0), evm.GasLimit, evm.GasPrice, data),
		evm.Signer(),
		TestFaucetKey,
	)
	require.NoError(t, err)

	txdata, err := tx.MarshalBinary()
	require.NoError(t, err)

	req, toUpload := solo.NewCallParamsOptimized(Interface.Name, FuncSendTransaction, 1024,
		FieldTransactionData, txdata,
	)
	req.WithIotas(100000)
	for _, v := range toUpload {
		chain.Env.PutBlobDataIntoRegistry(v)
	}

	_, err = chain.PostRequestSync(req, nil)

	require.NoError(t, err)

	contractAddress := crypto.CreateAddress(creatorAddress, nonce)

	callFn := func(sender *ecdsa.PrivateKey, name string, args ...interface{}) contractFnCaller {
		senderAddress := crypto.PubkeyToAddress(sender.PublicKey)

		nonce := getNonceFor(t, chain, senderAddress)

		callArguments, err := contractABI.Pack(name, args...)
		require.NoError(t, err)

		unsignedTx := types.NewTransaction(nonce, contractAddress, big.NewInt(0), evm.GasLimit, evm.GasPrice, callArguments)

		tx, err := types.SignTx(
			unsignedTx,
			evm.Signer(),
			TestFaucetKey,
		)
		require.NoError(t, err)

		txdata, err := tx.MarshalBinary()
		require.NoError(t, err)

		return func(userWallet *ed25519.KeyPair, iotas uint64) (*Receipt, uint64, error) {
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

	return contractAddress, callFn
}

// helper to reuse code to call the `retrieve` view in the storage contract
func GetCallRetrieveView(t *testing.T, chain *solo.Chain, contractAddress common.Address, contractABI abi.ABI) func() uint32 {
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
