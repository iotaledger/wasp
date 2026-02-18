package tests

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/clients/apiextensions"
	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/evm/evmtest"
	"github.com/iotaledger/wasp/v2/packages/evm/jsonrpc/jsonrpctest"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/core/root"
	"github.com/iotaledger/wasp/v2/packages/vm/core/testcore/contracts/inccounter"
	"github.com/iotaledger/wasp/v2/packages/webapi/models"
	"github.com/iotaledger/wasp/v2/tools/cluster"
)

func (e *ChainEnv) checkCoreContracts() {
	for i := range e.Chain.AllPeers {
		cl := e.Chain.Client(nil, i)
		ret, err := cl.CallView(context.Background(), governance.ViewGetChainInfo.Message())
		require.NoError(e.t, err)
		info, err := governance.ViewGetChainInfo.DecodeOutput(ret)
		require.NoError(e.t, err)

		require.EqualValues(e.t, e.Chain.OriginatorID(), info.ChainAdmin)

		records, err := e.Chain.Client(nil, i).
			CallView(context.Background(), root.ViewGetContractRecords.Message())
		require.NoError(e.t, err)

		contractRegistry, err := root.ViewGetContractRecords.DecodeOutput(records)
		require.NoError(e.t, err)
		for _, rec := range corecontracts.All {
			foundHname := slices.ContainsFunc(contractRegistry, func(tuple lo.Tuple2[*isc.Hname, *root.ContractRecord]) bool {
				return *tuple.A == rec.Hname()
			})
			require.True(e.t, foundHname, "core contract %s %+v missing", rec.Name, rec.Hname())
		}
	}
}

func (e *ChainEnv) checkRootsOutside() {
	for _, rec := range corecontracts.All {
		recBack, err := e.findContract(rec.Name)
		require.NoError(e.t, err)
		require.NotNil(e.t, recBack)
		require.EqualValues(e.t, rec.Name, recBack.Name)
	}
}

func (e *ChainEnv) GetL1Balance(addr *iotago.Address, coinType coin.Type) coin.Value {
	l1client := e.Chain.Cluster.L1Client()
	getBalance, err := l1client.GetBalance(context.TODO(), iotagraphql.GetBalanceRequest{Owner: *addr})
	require.NoError(e.t, err)
	return coin.Value(getBalance.TotalBalance.Uint64())
}

func (e *ChainEnv) GetL2Balance(agentID isc.AgentID, coinType coin.Type, nodeIndex ...int) coin.Value {
	idx := 0
	if len(nodeIndex) > 0 {
		idx = nodeIndex[0]
	}

	balance, _, err := e.Chain.Cluster.WaspClient(idx).CorecontractsAPI.
		AccountsGetAccountBalance(context.Background(), agentID.String()).
		Execute()
	require.NoError(e.t, err)

	assets, err := apiextensions.AssetsFromAPIResponse(balance)
	require.NoError(e.t, err)

	return assets.CoinBalance(coinType)
}

func (e *ChainEnv) getBalanceOnChain(agentID isc.AgentID, coinType coin.Type, nodeIndex ...int) coin.Value {
	idx := 0
	if len(nodeIndex) > 0 {
		idx = nodeIndex[0]
	}

	balance, _, err := e.Chain.Cluster.WaspClient(idx).CorecontractsAPI.
		AccountsGetAccountBalance(context.Background(), agentID.String()).
		Execute()
	require.NoError(e.t, err)

	assets, err := apiextensions.AssetsFromAPIResponse(balance)
	require.NoError(e.t, err)

	return assets.CoinBalance(coinType)
}

func (e *ChainEnv) checkBalanceOnChain(agentID isc.AgentID, coinType coin.Type, expected coin.Value) {
	actual := e.getBalanceOnChain(agentID, coinType)
	require.EqualValues(e.t, expected, actual)
}

func (e *ChainEnv) getChainInfo() (isc.ChainID, isc.AgentID) {
	chainInfo, _, err := e.Chain.Cluster.WaspClient(0).ChainsAPI.
		GetChainInfo(context.Background()).
		Execute()
	require.NoError(e.t, err)

	chainID, err := isc.ChainIDFromString(chainInfo.ChainID)
	require.NoError(e.t, err)

	admin, err := isc.AgentIDFromString(chainInfo.ChainAdmin)
	require.NoError(e.t, err)

	return chainID, admin
}

func (e *ChainEnv) findContract(name string, nodeIndex ...int) (*root.ContractRecord, error) {
	i := 0
	if len(nodeIndex) > 0 {
		i = nodeIndex[0]
	}

	hname := isc.Hn(name)

	// TODO: Validate with develop
	ret, err := apiextensions.CallView(
		context.Background(),
		e.Chain.Cluster.WaspClient(i),

		apiextensions.CallViewReq(root.ViewFindContract.Message(hname)),
	)

	require.NoError(e.t, err)

	found, recBin, err := root.ViewFindContract.DecodeOutput(ret)
	require.NoError(e.t, err)

	if !found {
		return nil, nil
	}

	return *recBin, nil
}

// region waitUntilProcessed ///////////////////////////////////////////////////

const pollPeriod = 500 * time.Millisecond

func waitTrue(timeout time.Duration, fun func() bool) bool {
	deadline := time.Now().Add(timeout)
	for {
		if fun() {
			return true
		}
		time.Sleep(pollPeriod)
		if time.Now().After(deadline) {
			return false
		}
	}
}

func (e *ChainEnv) balanceEquals(agentID isc.AgentID, amount int) conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		ret, err := apiextensions.CallView(
			context.Background(),
			e.Chain.Cluster.WaspClient(nodeIndex),
			apiclient.ContractCallViewRequest{
				ContractHName: accounts.Contract.Hname().String(),
				FunctionHName: accounts.ViewBalanceBaseToken.Hname().String(),
				Arguments:     models.ToCallArgumentsJSON(accounts.ViewBalanceBaseToken.Message(&agentID).Params),
			})
		if err != nil {
			e.t.Logf("chainEnv::counterEquals: failed to call GetCounter: %v", err)
			return false
		}

		balance, err := accounts.ViewBalanceBaseToken.DecodeOutput(ret)
		require.NoError(e.t, err)

		fmt.Printf("CURRENT BALANCE: %d, EXPECTED: %d\n", balance, amount)

		return coin.Value(amount) == balance
	}
}

func (e *ChainEnv) checkNRequests(clusterTestEnv *clusterTestEnv, numRequests int64, writeNodeIndex int, readNodeIndexes []int, sleepBeforeVerify time.Duration) error {
	storageContractAddr, transactions, err := e.sendNRequests(clusterTestEnv, numRequests, writeNodeIndex, true)
	if err != nil {
		return err
	}

	time.Sleep(sleepBeforeVerify)

	err = e.verifyNRequests(context.Background(), transactions, numRequests, readNodeIndexes, storageContractAddr, nil)
	if err != nil {
		return err
	}
	return nil
}

func (e *ChainEnv) sendNRequests(clusterTestEnv *clusterTestEnv, numRequests int64, writeNodeIndex int, buffered bool) (storageContractAddr common.Address, transactions chan *types.Transaction, err error) {
	evmPvtKey, evmAddr := clusterTestEnv.NewAccountWithL2Funds()

	storageContractAddr, storageContractABI := e.DeploySolidityContract(evmPvtKey, evmtest.StorageContractABI, evmtest.StorageContractBytecode, uint32(0))

	jsonRPCClient := e.EVMJSONRPClient(writeNodeIndex)
	nonce := e.GetNonceEVM(evmAddr)

	if buffered {
		transactions = make(chan *types.Transaction, numRequests)
	} else {
		transactions = make(chan *types.Transaction)
	}

	go func() {
		defer close(transactions)
		for i := range numRequests {
			callArguments, err := storageContractABI.Pack("increment")
			if err != nil {
				e.t.Logf("sendNRequests: failed to send transaction: %v", err)
				return
			}
			tx, err := types.SignTx(
				types.NewTransaction(nonce+uint64(i), storageContractAddr, big.NewInt(0), 100000, e.GetGasPriceEVM(), callArguments),
				EVMSigner(),
				evmPvtKey,
			)
			if err != nil {
				e.t.Logf("sendNRequests: failed to send transaction: %v", err)
				return
			}
			err = jsonRPCClient.SendTransaction(context.Background(), tx)
			if err != nil {
				e.t.Logf("sendNRequests: failed to send transaction: %v", err)
				return
			}
			transactions <- tx
		}
	}()

	return storageContractAddr, transactions, nil
}

func (e *ChainEnv) verifyNRequests(ctx context.Context, transactions chan *types.Transaction, numRequests int64, readNodeIndexes []int, storageContractAddr common.Address, cb func(tx *types.Transaction)) error {

outer:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case tx, ok := <-transactions:
			if !ok {
				break outer
			}
			_, err := e.Clu.MultiClient().WaitUntilEVMRequestProcessedSuccessfully(ctx, e.Chain.ChainID, tx.Hash(), false, 90*time.Second)
			if err != nil {
				return err
			}
			if cb != nil {
				cb(tx)
			}
		}
	}

	// read the counter value from Storage.sol via retrieve()
	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	if err != nil {
		return err
	}

	callData, err := contractABI.Pack("retrieve")
	if err != nil {
		return err
	}

	for _, nodeIndex := range readNodeIndexes {
		jsonRPCClient := e.EVMJSONRPClient(nodeIndex)
		callMsg := ethereum.CallMsg{To: &storageContractAddr, Data: callData}
		ret, err := jsonRPCClient.CallContract(context.Background(), callMsg, nil)
		if err != nil {
			return err
		}
		out, err := contractABI.Unpack("retrieve", ret)
		if err != nil {
			return err
		}
		counter := out[0].(uint32)
		if counter != uint32(numRequests) {
			return fmt.Errorf("unexpected counter value on node %d", nodeIndex)
		}
	}
	return nil
}

func (e *ChainEnv) accountExists(agentID isc.AgentID) conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		return e.getBalanceOnChain(agentID, coin.BaseTokenType, nodeIndex) > 0
	}
}

func (e *ChainEnv) contractIsDeployed() conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		ret, err := e.findContract(inccounter.Contract.Name, nodeIndex)
		if err != nil {
			return false
		}
		return ret.Name == inccounter.Contract.Name
	}
}

type conditionFn func(t *testing.T, nodeIndex int) bool

func waitUntil(t *testing.T, fn conditionFn, nodeIndexes []int, timeout time.Duration, logMsg ...string) {
	for _, nodeIndex := range nodeIndexes {
		if len(logMsg) > 0 {
			t.Logf("-->Waiting for '%s' on node %v...", logMsg[0], nodeIndex)
		}
		w := waitTrue(timeout, func() bool {
			return fn(t, nodeIndex)
		})
		if !w {
			if len(logMsg) > 0 {
				t.Errorf("-->Waiting for %s on node %v... FAILED after %v", logMsg[0], nodeIndex, timeout)
			} else {
				t.Errorf("-->Waiting on node %v... FAILED after %v", nodeIndex, timeout)
			}
			t.Helper()
			t.Fatal()
		}
	}
}

// endregion ///////////////////////////////////////////////////////////////

func setupClusterTest(t *testing.T, clusterSize int, committee []int, dirnameOpt ...string) *ChainEnv {
	quorum := uint16((2*len(committee))/3 + 1)

	dirname := ""
	if len(dirnameOpt) > 0 {
		dirname = dirnameOpt[0]
	}
	clu := newCluster(t, waspClusterOpts{
		nNodes:  clusterSize,
		dirName: dirname,
	})

	addr, err := clu.RunDistributedKeyGeneration(committee, quorum)
	require.NoError(t, err)

	t.Logf("generated state address: %s", addr.String())

	chain, err := clu.DeployChain(clu.Config.AllNodes(), committee, quorum, addr, false)
	require.NoError(t, err)
	t.Logf("deployed chainID: %s", chain.ChainID)

	e := &ChainEnv{
		t:     t,
		Clu:   clu,
		Chain: chain,
	}

	return e
}

func newClusterTestEnv(t *testing.T, env *ChainEnv, nodeIndex int) *clusterTestEnv {
	evmJSONRPCPath := "/v1/chain/evm"
	jsonRPCEndpoint := env.Clu.Config.APIHost(nodeIndex) + evmJSONRPCPath
	rawClient, err := rpc.DialHTTP(jsonRPCEndpoint)
	require.NoError(t, err)
	client := ethclient.NewClient(rawClient)
	t.Cleanup(client.Close)

	waitTxConfirmed := func(txHash common.Hash) error {
		c := env.Chain.Client(nil, nodeIndex)
		reqID := isc.RequestIDFromEVMTxHash(txHash)
		receipt, _, err := c.WaspClient.ChainsAPI.
			WaitForRequest(context.Background(), reqID.String()).
			TimeoutSeconds(10).
			Execute()
		if err != nil {
			return err
		}

		if receipt.ErrorMessage != nil {
			return errors.New(*receipt.ErrorMessage)
		}

		return nil
	}

	e := &clusterTestEnv{
		Env: jsonrpctest.Env{
			T:               t,
			Client:          client,
			RawClient:       rawClient,
			ChainID:         evm.DefaultChainID,
			WaitTxConfirmed: waitTxConfirmed,
		},
		ChainEnv: *env,
	}
	e.Env.NewAccountWithL2Funds = e.newEthereumAccountWithL2Funds
	return e
}

const transferAllowanceToGasBudgetBaseTokens = 1 * isc.Million

func (e *clusterTestEnv) newEthereumAccountWithL2Funds(baseTokens ...coin.Value) (*ecdsa.PrivateKey, common.Address) {
	ethKey, ethAddr := newEthereumAccount()

	var walletKey *cryptolib.KeyPair
	var walletAddr *cryptolib.Address
	var err error
	err = cluster.Retry(func() error {
		walletKey, walletAddr, err = e.Clu.NewKeyPairWithFunds()
		return err
	}, 7)
	require.NoError(e.T, err)

	var amount coin.Value
	if len(baseTokens) > 0 {
		amount = baseTokens[0]
	} else {
		amount = e.Clu.L1BaseTokens(walletAddr) - transferAllowanceToGasBudgetBaseTokens - iotagraphql.DefaultGasBudget
	}
	tx, err := e.Chain.Client(walletKey).PostRequest(
		context.Background(),
		accounts.FuncTransferAllowanceTo.Message(isc.NewEthereumAddressAgentID(ethAddr)),
		chainclient.PostRequestParams{
			Transfer:  isc.NewAssets(amount + transferAllowanceToGasBudgetBaseTokens),
			Allowance: isc.NewAssets(amount),
			GasBudget: iotagraphql.DefaultGasBudget,
		},
	)
	require.NoError(e.T, err)

	// We have to wait not only for the committee to process the request, but also for access nodes to get that info.
	_, err = e.Chain.AllNodesMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), e.Chain.ChainID, tx, false, 30*time.Second)
	require.NoError(e.T, err)

	return ethKey, ethAddr
}

type clusterTestEnv struct {
	jsonrpctest.Env
	ChainEnv
}
