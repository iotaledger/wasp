package tests

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/evm/evmtest"
	"github.com/iotaledger/wasp/v2/packages/evm/evmutil"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/tools/cluster"
)

type ChainEnv struct {
	t               *testing.T
	Clu             *cluster.Cluster
	Chain           *cluster.Chain
	testContractEnv *TestContractEnv
}

func SetupWithChain(t *testing.T, opts ...waspClusterOpts) *ChainEnv {
	clu := newCluster(t, opts...)
	e := &ChainEnv{t: t, Clu: clu}
	chain, err := e.Clu.DeployDefaultChain()
	require.NoError(t, err)
	return newChainEnv(e.t, e.Clu, chain)
}

func SetupWithChainWithOpts(t *testing.T, opt *waspClusterOpts, committeeNodes []int, quorum uint16, blockKeepAmount ...int32) *ChainEnv {
	clu := newCluster(t, *opt)
	e := &ChainEnv{t: t, Clu: clu}
	chain, err := clu.DeployChainWithDKG(clu.Config.AllNodes(), committeeNodes, quorum, blockKeepAmount...)
	require.NoError(t, err)
	return newChainEnv(e.t, e.Clu, chain)
}

func newChainEnv(t *testing.T, clu *cluster.Cluster, chain *cluster.Chain) *ChainEnv {
	env := &ChainEnv{
		t:     t,
		Clu:   clu,
		Chain: chain,
	}
	env.testContractEnv = env.NewTestContractEnv(t)
	return env
}

func (e *ChainEnv) NewChainClient(keyPair cryptolib.Signer, nodeIndex ...int) *chainclient.Client {
	return e.Chain.Client(keyPair)
}

func (e *ChainEnv) NewRandomChainClient() (*chainclient.Client, *cryptolib.KeyPair) {
	keypair, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(e.t, err)
	return e.NewChainClient(keypair), keypair
}

// wait until the TX has been confirmed
func (e *ChainEnv) DepositFunds(amount coin.Value, keyPair *cryptolib.KeyPair) {
	client := e.Chain.Client(keyPair)
	params := chainclient.PostRequestParams{
		Transfer:  isc.NewAssets(amount),
		Allowance: isc.NewAssets(amount - iotaclient.DefaultGasBudget),
		GasBudget: iotaclient.DefaultGasBudget,
	}
	tx, err := client.PostRequest(context.Background(), accounts.FuncDeposit.Message(), params)
	require.NoError(e.t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), tx, true, 30*time.Second)
	require.NoError(e.t, err, "Error while WaitUntilAllRequestsProcessedSuccessfully for tx.ID=%v", tx.Digest)
}

func (e *ChainEnv) TransferFundsTo(assets *isc.Assets, keyPair *cryptolib.KeyPair, targetAccount isc.AgentID) {
	client := e.Chain.Client(keyPair)
	transferAssets := assets.Clone()
	l2GasFee := 1 * isc.Million
	tx, err := client.PostRequest(context.Background(), accounts.FuncTransferAllowanceTo.Message(targetAccount), chainclient.PostRequestParams{
		Transfer:    transferAssets.AddBaseTokens(coin.Value(l2GasFee)),
		Allowance:   assets,
		GasBudget:   iotaclient.DefaultGasBudget,
		L2GasBudget: uint64(l2GasFee),
	})
	require.NoError(e.t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), tx, false, 30*time.Second)
	require.NoError(e.t, err, "Error while WaitUntilAllRequestsProcessedSuccessfully for tx.ID=%v", tx.Digest)
}

// DeploySolidityContract deploys a given solidity contract with a given private key, returns the create contract address
// it will send the EVM request to node #0, using the default EVM chainID, that can be changed if needed
func (e *ChainEnv) DeploySolidityContract(creator *ecdsa.PrivateKey, abiJSON string, bytecode []byte, args ...interface{}) (common.Address, abi.ABI) {
	creatorAddress := crypto.PubkeyToAddress(creator.PublicKey)

	nonce := e.GetNonceEVM(creatorAddress)

	contractABI, err := abi.JSON(strings.NewReader(abiJSON))
	require.NoError(e.t, err)
	constructorArguments, err := contractABI.Pack("", args...)
	require.NoError(e.t, err)

	data := []byte{}
	data = append(data, bytecode...)
	data = append(data, constructorArguments...)

	value := big.NewInt(0)

	jsonRPCClient := e.EVMJSONRPClient(0) // send request to node #0
	gasLimit, err := jsonRPCClient.EstimateGas(context.Background(),
		ethereum.CallMsg{
			From:  creatorAddress,
			Value: value,
			Data:  data,
		})
	require.NoError(e.t, err)

	tx, err := types.SignTx(
		types.NewContractCreation(nonce, value, gasLimit, e.GetGasPriceEVM(), data),
		EVMSigner(),
		creator,
	)
	require.NoError(e.t, err)

	err = jsonRPCClient.SendTransaction(context.Background(), tx)
	require.NoError(e.t, err)

	// await tx confirmed
	_, err = e.Chain.CommitteeMultiClient().WaitUntilEVMRequestProcessedSuccessfully(context.Background(), tx.Hash(), false, 30*time.Second)
	require.NoError(e.t, err)

	return crypto.CreateAddress(creatorAddress, nonce), contractABI
}

func (e *ChainEnv) GetNonceEVM(addr common.Address) uint64 {
	nonce, err := e.EVMJSONRPClient(0).NonceAt(context.Background(), addr, nil)
	require.NoError(e.t, err)
	return nonce
}

func (e *ChainEnv) GetGasPriceEVM() *big.Int {
	res, err := e.EVMJSONRPClient(0).SuggestGasPrice(context.Background())
	require.NoError(e.t, err)
	return res
}

func (e *ChainEnv) EVMJSONRPClient(nodeIndex int) *ethclient.Client {
	return NewEVMJSONRPClient(e.t, e.Clu, nodeIndex)
}

func NewEVMJSONRPClient(t *testing.T, clu *cluster.Cluster, nodeIndex int) *ethclient.Client {
	evmJSONRPCPath := "/v1/chain/evm"
	jsonRPCEndpoint := clu.Config.APIHost(nodeIndex) + evmJSONRPCPath
	rawClient, err := rpc.DialHTTP(jsonRPCEndpoint)
	require.NoError(t, err)
	jsonRPCClient := ethclient.NewClient(rawClient)
	t.Cleanup(jsonRPCClient.Close)
	return jsonRPCClient
}

func EVMSigner() types.Signer {
	return evmutil.Signer(big.NewInt(int64(evm.DefaultChainID))) // use default evm chainID
}

func newEthereumAccount() (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	return key, crypto.PubkeyToAddress(key.PublicKey)
}

type TestContractEnv struct {
	t                   *testing.T
	EvmTesterAddr       common.Address
	EvmTestContractAddr common.Address
	EvmTestContractABI  abi.ABI
	EvmTesterPrivateKey *ecdsa.PrivateKey
}

func (e *ChainEnv) NewTestContractEnv(t *testing.T) *TestContractEnv {
	keyPair, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	evmPvtKey, evmAddr := solo.NewEthereumAccount()
	evmAgentID := isc.NewEthereumAddressAgentID(evmAddr)
	e.TransferFundsTo(isc.NewAssets(1*isc.Million), keyPair, evmAgentID)
	contractAddr, contractABI := e.DeploySolidityContract(evmPvtKey, evmtest.StorageContractABI, evmtest.StorageContractBytecode, uint32(42))
	return &TestContractEnv{
		t:                   t,
		EvmTestContractAddr: contractAddr,
		EvmTestContractABI:  contractABI,
		EvmTesterAddr:       evmAddr,
		EvmTesterPrivateKey: evmPvtKey,
	}
}

func (e *ChainEnv) CallStore(archiveClient, lightClient *ethclient.Client, input uint64) *types.Transaction {
	if archiveClient == nil {
		archiveClient = e.EVMJSONRPClient(0)
	}
	if lightClient == nil {
		lightClient = e.EVMJSONRPClient(1)
	}

	callArguments, err := e.testContractEnv.EvmTestContractABI.Pack("store", uint32(input))
	require.NoError(e.testContractEnv.t, err)
	nonce := e.GetNonceEVM(e.testContractEnv.EvmTestContractAddr)
	tx, err := types.SignTx(
		types.NewTransaction(nonce, e.testContractEnv.EvmTestContractAddr, big.NewInt(0), 100000, e.GetGasPriceEVM(), callArguments),
		EVMSigner(),
		e.testContractEnv.EvmTesterPrivateKey,
	)
	require.NoError(e.t, err)
	err = archiveClient.SendTransaction(context.Background(), tx)
	require.NoError(e.t, err)
	// await tx confirmed
	for i := 0; i < 3; i++ {
		_, err = e.Clu.MultiClient().WaitUntilEVMRequestProcessedSuccessfully(context.Background(), tx.Hash(), false, 30*time.Second)
		if err == nil {
			break
		}
		time.Sleep(15 * time.Second)
	}
	require.NoError(e.testContractEnv.t, err)
	return tx
}

func (e *ChainEnv) CallRetrieve(archiveClient *ethclient.Client) uint32 {
	if archiveClient == nil {
		archiveClient = e.EVMJSONRPClient(0)
	}
	callArgs, err := e.testContractEnv.EvmTestContractABI.Pack("retrieve")
	require.NoError(e.t, err)
	callMsg := ethereum.CallMsg{
		To:   &e.testContractEnv.EvmTestContractAddr,
		Data: callArgs,
	}
	ret, err := archiveClient.CallContract(context.Background(), callMsg, nil)
	require.NoError(e.t, err)
	val, err := e.testContractEnv.EvmTestContractABI.Unpack("retrieve", ret)
	require.NoError(e.t, err)
	return val[0].(uint32)
}
