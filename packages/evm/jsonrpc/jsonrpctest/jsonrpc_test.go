// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpctest

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/packages/evm/evmflavors"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

func withEVMFlavors(t *testing.T, f func(*testing.T, *coreutil.ContractInfo)) {
	for _, evmFlavor := range evmflavors.All {
		t.Run(evmFlavor.Name, func(t *testing.T) { f(t, evmFlavor) })
	}
}

type soloTestEnv struct {
	Env
	solo *solo.Solo
}

func newSoloTestEnv(t *testing.T, evmFlavor *coreutil.ContractInfo) *soloTestEnv {
	evmtest.InitGoEthLogger(t)

	chainID := evm.DefaultChainID

	s := solo.New(t, true, false).WithNativeContract(evmflavors.Processors[evmFlavor.Name])
	chainOwner, _ := s.NewKeyPairWithFunds()
	chain := s.NewChain(chainOwner, "iscpchain")
	err := chain.DeployContract(chainOwner, evmFlavor.Name, evmFlavor.ProgramHash,
		evm.FieldChainID, codec.EncodeUint16(uint16(chainID)),
		evm.FieldGenesisAlloc, evmtypes.EncodeGenesisAlloc(core.GenesisAlloc{
			evmtest.FaucetAddress: {Balance: evmtest.FaucetSupply},
		}),
	)
	require.NoError(t, err)
	signer, _ := s.NewKeyPairWithFunds()
	backend := jsonrpc.NewSoloBackend(s, chain, signer)
	evmChain := jsonrpc.NewEVMChain(backend, chainID, evmFlavor.Name)

	accountManager := jsonrpc.NewAccountManager(evmtest.Accounts)

	rpcsrv := jsonrpc.NewServer(evmChain, accountManager)
	t.Cleanup(rpcsrv.Stop)

	rawClient := rpc.DialInProc(rpcsrv)
	client := ethclient.NewClient(rawClient)
	t.Cleanup(client.Close)

	return &soloTestEnv{
		Env: Env{
			T:         t,
			EVMFlavor: evmFlavor,
			Server:    rpcsrv,
			Client:    client,
			RawClient: rawClient,
			ChainID:   chainID,
		},
		solo: s,
	}
}

func generateKey(t *testing.T) (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return key, addr
}

func TestRPCGetBalance(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		_, receiverAddress := generateKey(t)
		require.Zero(t, big.NewInt(0).Cmp(env.Balance(receiverAddress)))
		env.RequestFunds(receiverAddress)
		require.Zero(t, big.NewInt(1e18).Cmp(env.Balance(receiverAddress)))
	})
}

func TestRPCGetCode(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		creator, creatorAddress := generateKey(t)

		// account address
		{
			env.RequestFunds(creatorAddress)
			require.Empty(t, env.Code(creatorAddress))
		}
		// contract address
		{
			contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
			require.NoError(t, err)
			_, contractAddress := env.DeployEVMContract(creator, contractABI, evmtest.StorageContractBytecode, uint32(42))
			require.NotEmpty(t, env.Code(contractAddress))
		}
	})
}

func TestRPCGetStorage(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		creator, creatorAddress := generateKey(t)

		env.RequestFunds(creatorAddress)

		contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
		require.NoError(t, err)
		_, contractAddress := env.DeployEVMContract(creator, contractABI, evmtest.StorageContractBytecode, uint32(42))

		// first static variable in contract (uint32 n) has slot 0. See:
		// https://docs.soliditylang.org/en/v0.6.6/miscellaneous.html#layout-of-state-variables-in-storage
		slot := common.Hash{}
		ret := env.Storage(contractAddress, slot)

		var v uint32
		err = contractABI.UnpackIntoInterface(&v, "retrieve", ret)
		require.NoError(t, err)
		require.Equal(t, uint32(42), v)
	})
}

func TestRPCBlockNumber(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		_, receiverAddress := generateKey(t)
		require.EqualValues(t, 0, env.BlockNumber())
		env.RequestFunds(receiverAddress)
		require.EqualValues(t, 1, env.BlockNumber())
	})
}

func TestRPCGetTransactionCount(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		_, receiverAddress := generateKey(t)
		require.EqualValues(t, 0, env.NonceAt(evmtest.FaucetAddress))
		env.RequestFunds(receiverAddress)
		require.EqualValues(t, 1, env.NonceAt(evmtest.FaucetAddress))
	})
}

func TestRPCGetBlockByNumber(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		_, receiverAddress := generateKey(t)
		require.EqualValues(t, 0, env.BlockByNumber(big.NewInt(0)).Number().Uint64())
		env.RequestFunds(receiverAddress)
		require.EqualValues(t, 1, env.BlockByNumber(big.NewInt(1)).Number().Uint64())
	})
}

func TestRPCGetBlockByHash(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		_, receiverAddress := generateKey(t)
		require.Nil(t, env.BlockByHash(common.Hash{}))
		require.EqualValues(t, 0, env.BlockByHash(env.BlockByNumber(big.NewInt(0)).Hash()).Number().Uint64())
		env.RequestFunds(receiverAddress)
		require.EqualValues(t, 1, env.BlockByHash(env.BlockByNumber(big.NewInt(1)).Hash()).Number().Uint64())
	})
}

func TestRPCGetTransactionByHash(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		_, receiverAddress := generateKey(t)
		require.Nil(t, env.TransactionByHash(common.Hash{}))
		env.RequestFunds(receiverAddress)
		block1 := env.BlockByNumber(big.NewInt(1))
		tx := env.TransactionByHash(block1.Transactions()[0].Hash())
		require.Equal(t, block1.Transactions()[0].Hash(), tx.Hash())
	})
}

func TestRPCGetTransactionByBlockHashAndIndex(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		_, receiverAddress := generateKey(t)
		require.Nil(t, env.TransactionByBlockHashAndIndex(common.Hash{}, 0))
		env.RequestFunds(receiverAddress)
		block1 := env.BlockByNumber(big.NewInt(1))
		tx := env.TransactionByBlockHashAndIndex(block1.Hash(), 0)
		require.Equal(t, block1.Transactions()[0].Hash(), tx.Hash())
	})
}

func TestRPCGetUncleByBlockHashAndIndex(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		_, receiverAddress := generateKey(t)
		require.Nil(t, env.UncleByBlockHashAndIndex(common.Hash{}, 0))
		env.RequestFunds(receiverAddress)
		block1 := env.BlockByNumber(big.NewInt(1))
		require.Nil(t, env.UncleByBlockHashAndIndex(block1.Hash(), 0))
	})
}

func TestRPCGetTransactionByBlockNumberAndIndex(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		_, receiverAddress := generateKey(t)
		require.Nil(t, env.TransactionByBlockNumberAndIndex(big.NewInt(3), 0))
		env.RequestFunds(receiverAddress)
		block1 := env.BlockByNumber(big.NewInt(1))
		tx := env.TransactionByBlockNumberAndIndex(block1.Number(), 0)
		require.EqualValues(t, block1.Hash(), *tx.BlockHash)
		require.EqualValues(t, 0, *tx.TransactionIndex)
	})
}

func TestRPCGetUncleByBlockNumberAndIndex(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		_, receiverAddress := generateKey(t)
		require.Nil(t, env.UncleByBlockNumberAndIndex(big.NewInt(3), 0))
		env.RequestFunds(receiverAddress)
		block1 := env.BlockByNumber(big.NewInt(1))
		require.Nil(t, env.UncleByBlockNumberAndIndex(block1.Number(), 0))
	})
}

func TestRPCGetTransactionCountByHash(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		_, receiverAddress := generateKey(t)
		env.RequestFunds(receiverAddress)
		block1 := env.BlockByNumber(big.NewInt(1))
		require.Positive(t, len(block1.Transactions()))
		require.EqualValues(t, len(block1.Transactions()), env.BlockTransactionCountByHash(block1.Hash()))
		require.EqualValues(t, 0, env.BlockTransactionCountByHash(common.Hash{}))
	})
}

func TestRPCGetUncleCountByBlockHash(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		_, receiverAddress := generateKey(t)
		env.RequestFunds(receiverAddress)
		block1 := env.BlockByNumber(big.NewInt(1))
		require.Zero(t, len(block1.Uncles()))
		require.EqualValues(t, len(block1.Uncles()), env.UncleCountByBlockHash(block1.Hash()))
		require.EqualValues(t, 0, env.UncleCountByBlockHash(common.Hash{}))
	})
}

func TestRPCGetTransactionCountByNumber(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		_, receiverAddress := generateKey(t)
		env.RequestFunds(receiverAddress)
		block1 := env.BlockByNumber(nil)
		require.Positive(t, len(block1.Transactions()))
		require.EqualValues(t, len(block1.Transactions()), env.BlockTransactionCountByNumber())
	})
}

func TestRPCGetUncleCountByBlockNumber(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		_, receiverAddress := generateKey(t)
		env.RequestFunds(receiverAddress)
		block1 := env.BlockByNumber(big.NewInt(1))
		require.Zero(t, len(block1.Uncles()))
		require.EqualValues(t, len(block1.Uncles()), env.UncleCountByBlockNumber(big.NewInt(1)))
	})
}

func TestRPCAccounts(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		accounts := env.Accounts()
		require.Equal(t, len(evmtest.Accounts), len(accounts))
	})
}

func TestRPCSign(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		signed := env.Sign(evmtest.AccountAddress(0), []byte("hello"))
		require.NotEmpty(t, signed)
	})
}

func TestRPCSignTransaction(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)

		from := evmtest.AccountAddress(0)
		to := evmtest.AccountAddress(1)
		gas := hexutil.Uint64(params.TxGas)
		nonce := hexutil.Uint64(env.NonceAt(from))
		signed := env.SignTransaction(&jsonrpc.SendTxArgs{
			From:     from,
			To:       &to,
			Gas:      &gas,
			GasPrice: (*hexutil.Big)(evm.GasPrice),
			Value:    (*hexutil.Big)(RequestFundsAmount),
			Nonce:    &nonce,
		})
		require.NotEmpty(t, signed)

		// test that the signed tx can be sent
		env.RequestFunds(from)
		err := env.RawClient.Call(nil, "eth_sendRawTransaction", hexutil.Encode(signed))
		require.NoError(t, err)
	})
}

func TestRPCSendTransaction(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)

		from := evmtest.AccountAddress(0)
		env.RequestFunds(from)

		to := evmtest.AccountAddress(1)
		gas := hexutil.Uint64(params.TxGas)
		nonce := hexutil.Uint64(env.NonceAt(from))
		txHash := env.MustSendTransaction(&jsonrpc.SendTxArgs{
			From:     from,
			To:       &to,
			Gas:      &gas,
			GasPrice: (*hexutil.Big)(evm.GasPrice),
			Value:    (*hexutil.Big)(RequestFundsAmount),
			Nonce:    &nonce,
		})
		require.NotEqualValues(t, common.Hash{}, txHash)
	})
}

func TestRPCGetTxReceiptRegularTx(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		_, creatorAddr := generateKey(t)

		tx := env.RequestFunds(creatorAddr)
		receipt := env.MustTxReceipt(tx.Hash())

		require.EqualValues(t, types.LegacyTxType, receipt.Type)
		require.EqualValues(t, types.ReceiptStatusSuccessful, receipt.Status)
		require.NotZero(t, receipt.CumulativeGasUsed)
		require.EqualValues(t, types.Bloom{}, receipt.Bloom)
		require.EqualValues(t, 0, len(receipt.Logs))

		require.EqualValues(t, tx.Hash(), receipt.TxHash)
		require.EqualValues(t, common.Address{}, receipt.ContractAddress)
		require.NotZero(t, receipt.GasUsed)

		require.EqualValues(t, big.NewInt(1), receipt.BlockNumber)
		require.EqualValues(t, env.BlockByNumber(big.NewInt(1)).Hash(), receipt.BlockHash)
		require.EqualValues(t, 0, receipt.TransactionIndex)
	})
}

func TestRPCGetTxReceiptContractCreation(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		creator, _ := generateKey(t)

		contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
		require.NoError(t, err)
		tx, contractAddress := env.DeployEVMContract(creator, contractABI, evmtest.StorageContractBytecode, uint32(42))
		receipt := env.MustTxReceipt(tx.Hash())

		require.EqualValues(t, types.LegacyTxType, receipt.Type)
		require.EqualValues(t, types.ReceiptStatusSuccessful, receipt.Status)
		require.NotZero(t, receipt.CumulativeGasUsed)
		require.EqualValues(t, types.Bloom{}, receipt.Bloom)
		require.EqualValues(t, 0, len(receipt.Logs))

		require.EqualValues(t, tx.Hash(), receipt.TxHash)
		require.EqualValues(t, contractAddress, receipt.ContractAddress)
		require.NotZero(t, receipt.GasUsed)

		require.EqualValues(t, big.NewInt(1), receipt.BlockNumber)
		require.EqualValues(t, env.BlockByNumber(big.NewInt(1)).Hash(), receipt.BlockHash)
		require.EqualValues(t, 0, receipt.TransactionIndex)
	})
}

func TestRPCGetTxReceiptMissing(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)

		_, err := env.TxReceipt(common.Hash{})
		require.Error(t, err)
		require.Equal(t, "not found", err.Error())
	})
}

func TestRPCCall(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		creator, creatorAddress := generateKey(t)
		contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
		require.NoError(t, err)
		_, contractAddress := env.DeployEVMContract(creator, contractABI, evmtest.StorageContractBytecode, uint32(42))

		callArguments, err := contractABI.Pack("retrieve")
		require.NoError(t, err)

		ret, err := env.Client.CallContract(context.Background(), ethereum.CallMsg{
			From: creatorAddress,
			To:   &contractAddress,
			Data: callArguments,
		}, nil)
		require.NoError(t, err)

		var v uint32
		err = contractABI.UnpackIntoInterface(&v, "retrieve", ret)
		require.NoError(t, err)
		require.Equal(t, uint32(42), v)
	})
}

func TestRPCGetLogs(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		newSoloTestEnv(t, evmFlavor).TestRPCGetLogs()
	})
}

func TestRPCGasLimit(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		newSoloTestEnv(t, evmFlavor).TestRPCGasLimit()
	})
}

func TestRPCEthChainID(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		env := newSoloTestEnv(t, evmFlavor)
		var chainID hexutil.Uint
		err := env.RawClient.Call(&chainID, "eth_chainId")
		require.NoError(t, err)
		require.EqualValues(t, evm.DefaultChainID, chainID)
	})
}
