package tests

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
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

	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/scclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/tools/cluster"
)

// TODO remove this?
func setupWithNoChain(t *testing.T, opt ...waspClusterOpts) *ChainEnv {
	clu := newCluster(t, opt...)
	return &ChainEnv{t: t, Clu: clu}
}

type ChainEnv struct {
	t     *testing.T
	Clu   *cluster.Cluster
	Chain *cluster.Chain
}

func newChainEnv(t *testing.T, clu *cluster.Cluster, chain *cluster.Chain) *ChainEnv {
	return &ChainEnv{
		t:     t,
		Clu:   clu,
		Chain: chain,
	}
}

type contractEnv struct {
	*ChainEnv
	programHash hashing.HashValue
}

func (e *ChainEnv) deployWasmContract(wasmName string, initParams map[string]interface{}) *contractEnv {
	ret := &contractEnv{ChainEnv: e}

	wasmPath := "wasm/" + wasmName + "_bg.wasm"

	wasm, err := os.ReadFile(wasmPath)
	require.NoError(e.t, err)
	chClient := chainclient.New(e.Clu.L1Client(), e.Clu.WaspClient(0), e.Chain.ChainID, e.Chain.OriginatorKeyPair)

	reqTx, err := chClient.DepositFunds(1_000_000)
	require.NoError(e.t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, reqTx, false, 30*time.Second)
	require.NoError(e.t, err)

	ph, err := e.Chain.DeployWasmContract(wasmName, wasm, initParams)
	require.NoError(e.t, err)
	ret.programHash = ph
	e.t.Logf("deployContract: proghash = %s\n", ph.String())
	return ret
}

func (e *ChainEnv) createNewClient() *scclient.SCClient {
	keyPair, addr, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(e.t, err)
	retries := 0
	for {
		outs, err := e.Clu.L1Client().OutputMap(addr)
		require.NoError(e.t, err)
		if len(outs) > 0 {
			break
		}
		retries++
		if retries > 10 {
			panic("createNewClient - funds aren't available")
		}
		time.Sleep(300 * time.Millisecond)
	}
	return e.Chain.SCClient(isc.Hn(nativeIncCounterSCName), keyPair)
}

func SetupWithChain(t *testing.T, opt ...waspClusterOpts) *ChainEnv {
	e := setupWithNoChain(t, opt...)
	chain, err := e.Clu.DeployDefaultChain()
	require.NoError(t, err)
	return newChainEnv(e.t, e.Clu, chain)
}

func (e *ChainEnv) NewChainClient() *chainclient.Client {
	wallet, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(e.t, err)
	return chainclient.New(e.Clu.L1Client(), e.Clu.WaspClient(0), e.Chain.ChainID, wallet)
}

func (e *ChainEnv) DepositFunds(amount uint64, keyPair *cryptolib.KeyPair) {
	accountsClient := e.Chain.SCClient(accounts.Contract.Hname(), keyPair)
	tx, err := accountsClient.PostRequest(accounts.FuncDeposit.Name, chainclient.PostRequestParams{
		Transfer: isc.NewAssetsBaseTokens(amount),
	})
	require.NoError(e.t, err)
	txID, err := tx.ID()
	require.NoError(e.t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, false, 30*time.Second)
	require.NoError(e.t, err, "Error while WaitUntilAllRequestsProcessedSuccessfully for tx.ID=%v", txID.ToHex())
}

func (e *ChainEnv) TransferFundsTo(assets *isc.Assets, nft *isc.NFT, keyPair *cryptolib.KeyPair, targetAccount isc.AgentID) {
	accountsClient := e.Chain.SCClient(accounts.Contract.Hname(), keyPair)
	transferAssets := assets.Clone()
	transferAssets.AddBaseTokens(1 * isc.Million) // to pay for the fees
	tx, err := accountsClient.PostRequest(accounts.FuncTransferAllowanceTo.Name, chainclient.PostRequestParams{
		Transfer:                 transferAssets,
		Args:                     map[kv.Key][]byte{accounts.ParamAgentID: codec.EncodeAgentID(targetAccount)},
		NFT:                      nft,
		Allowance:                assets,
		AutoAdjustStorageDeposit: false,
	})
	require.NoError(e.t, err)
	txID, err := tx.ID()
	require.NoError(e.t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, false, 30*time.Second)
	require.NoError(e.t, err, "Error while WaitUntilAllRequestsProcessedSuccessfully for tx.ID=%v", txID.ToHex())
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
	_, err = e.Clu.MultiClient().WaitUntilEVMRequestProcessedSuccessfully(e.Chain.ChainID, tx.Hash(), false, 5*time.Second)
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
	return NewEVMJSONRPClient(e.t, e.Chain.ChainID.String(), e.Clu, nodeIndex)
}

func NewEVMJSONRPClient(t *testing.T, chainID string, clu *cluster.Cluster, nodeIndex int) *ethclient.Client {
	evmJSONRPCPath := fmt.Sprintf("/v1/chains/%v/evm", chainID)
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
