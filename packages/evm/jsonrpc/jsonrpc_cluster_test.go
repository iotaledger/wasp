package jsonrpc

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
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/contracts/native/evmchain"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/iotaledger/wasp/tools/cluster/testutil"
	"github.com/stretchr/testify/require"
)

type envWithCluster struct {
	t         *testing.T
	cluster   *cluster.Cluster
	chain     *cluster.Chain
	server    *rpc.Server
	rawClient *rpc.Client
	client    *ethclient.Client
}

func newEnvWithCluster(t *testing.T) *envWithCluster {
	evmtest.InitGoEthLogger(t)

	clu := testutil.NewCluster(t)

	chain, err := clu.DeployDefaultChain()
	require.NoError(t, err)

	_, err = chain.DeployContract(
		evmchain.Interface.Name,
		evmchain.Interface.ProgramHash.String(),
		"EVM chain on top of ISCP",
		map[string]interface{}{
			evmchain.FieldGenesisAlloc: evmchain.EncodeGenesisAlloc(core.GenesisAlloc{
				evmtest.FaucetAddress: {Balance: evmtest.FaucetSupply},
			}),
		},
	)
	require.NoError(t, err)

	signer, _, err := clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	backend := NewWaspClientBackend(chain.Client(signer))
	evmChain := NewEVMChain(backend)

	accountManager := NewAccountManager(evmtest.Accounts)

	rpcsrv := NewServer(evmChain, accountManager)
	t.Cleanup(rpcsrv.Stop)

	rawClient := rpc.DialInProc(rpcsrv)
	client := ethclient.NewClient(rawClient)
	t.Cleanup(client.Close)

	return &envWithCluster{t, clu, chain, rpcsrv, rawClient, client}
}

func (e *envWithCluster) nonceAt(address common.Address) uint64 {
	nonce, err := e.client.NonceAt(context.Background(), address, nil)
	require.NoError(e.t, err)
	return nonce
}

func (e *envWithCluster) getLogs(q ethereum.FilterQuery) []types.Log {
	logs, err := e.client.FilterLogs(context.Background(), q)
	require.NoError(e.t, err)
	return logs
}

func (e *envWithCluster) deployEVMContract(creator *ecdsa.PrivateKey, contractABI abi.ABI, contractBytecode []byte, args ...interface{}) (*types.Transaction, common.Address) {
	creatorAddress := crypto.PubkeyToAddress(creator.PublicKey)

	nonce := e.nonceAt(creatorAddress)

	constructorArguments, err := contractABI.Pack("", args...)
	require.NoError(e.t, err)

	data := concatenate(contractBytecode, constructorArguments)

	value := big.NewInt(0)

	gasLimit := e.estimateGas(ethereum.CallMsg{
		From:     creatorAddress,
		To:       nil, // contract creation
		GasPrice: evm.GasPrice,
		Value:    value,
		Data:     data,
	})

	tx, err := types.SignTx(
		types.NewContractCreation(nonce, value, gasLimit, evm.GasPrice, data),
		evm.Signer(),
		creator,
	)
	require.NoError(e.t, err)

	err = e.client.SendTransaction(context.Background(), tx)
	require.NoError(e.t, err)

	return tx, crypto.CreateAddress(creatorAddress, nonce)
}

func (e *envWithCluster) estimateGas(msg ethereum.CallMsg) uint64 {
	gas, err := e.client.EstimateGas(context.Background(), msg)
	require.NoError(e.t, err)
	return gas
}

func (e *envWithCluster) sendTransaction(args *SendTxArgs) common.Hash {
	var res common.Hash
	err := e.rawClient.Call(&res, "eth_sendTransaction", args)
	require.NoError(e.t, err)
	return res
}

func TestEVMWithCluster(t *testing.T) {
	env := newEnvWithCluster(t)
	creator, creatorAddress := evmtest.Accounts[0], evmtest.AccountAddress(0)
	contractABI, err := abi.JSON(strings.NewReader(evmtest.ERC20ContractABI))
	require.NoError(t, err)

	contractAddress := crypto.CreateAddress(creatorAddress, env.nonceAt(creatorAddress))

	filterQuery := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	require.Empty(t, env.getLogs(filterQuery))

	env.deployEVMContract(creator, contractABI, evmtest.ERC20ContractBytecode, "TestCoin", "TEST")

	require.Equal(t, 1, len(env.getLogs(filterQuery)))

	recipientAddress := evmtest.AccountAddress(1)
	nonce := hexutil.Uint64(env.nonceAt(creatorAddress))
	callArguments, err := contractABI.Pack("transfer", recipientAddress, big.NewInt(1337))
	value := big.NewInt(0)
	gas := hexutil.Uint64(env.estimateGas(ethereum.CallMsg{
		From:  creatorAddress,
		To:    &contractAddress,
		Value: value,
		Data:  callArguments,
	}))
	require.NoError(t, err)
	env.sendTransaction(&SendTxArgs{
		From:     creatorAddress,
		To:       &contractAddress,
		Gas:      &gas,
		GasPrice: (*hexutil.Big)(evm.GasPrice),
		Value:    (*hexutil.Big)(value),
		Nonce:    &nonce,
		Data:     (*hexutil.Bytes)(&callArguments),
	})

	require.Equal(t, 2, len(env.getLogs(filterQuery)))
}
