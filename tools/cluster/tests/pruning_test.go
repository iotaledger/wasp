package tests

import (
	"context"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/evm/evmtest"
	"github.com/iotaledger/wasp/v2/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/v2/tools/cluster"
)

func TestPruning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cluster tests in short mode")
	}

	// t.Parallel()
	blockKeepAmount := 10
	clu := newCluster(t, waspClusterOpts{
		nNodes: 4,
		modifyConfig: func(nodeIndex int, configParams cluster.WaspConfigParams) cluster.WaspConfigParams {
			// set node 0 as an "archive node"
			if nodeIndex == 0 {
				configParams.PruningMinStatesToKeep = -1
			} else {
				// all other nodes will only keep 10 blocks
				configParams.PruningMinStatesToKeep = blockKeepAmount
			}

			return configParams
		},
	})

	// set blockKeepAmount (active state pruning) to 10 as well
	chain, err := clu.DeployChainWithDistKeyGen(clu.Config.AllNodes(), clu.Config.AllNodes(), 4, int32(blockKeepAmount))
	require.NoError(t, err)
	env := newChainEnv(t, clu, chain)

	const numRequests = 100

	initialBlockIndex, err := env.Chain.BlockIndex()
	require.NoError(t, err)

	archiveClientIndex := 0
	lightClientIndex := 1

	storageContractAddr, transactions, err := env.sendNRequests(newClusterTestEnv(t, env, archiveClientIndex), numRequests, archiveClientIndex, false)
	require.NoError(t, err)

	txs := make([]*types.Transaction, 0, numRequests)
	err = env.verifyNRequests(context.Background(), transactions, numRequests, clu.Config.AllNodes(), storageContractAddr, func(tx *types.Transaction) {
		txs = append(txs, tx)
	})
	require.NoError(t, err)
	require.Len(t, txs, numRequests)

	archiveClient := env.EVMJSONRPClient(archiveClientIndex)
	lightClient := env.EVMJSONRPClient(lightClientIndex)

	bn, err := archiveClient.BlockNumber(context.Background())
	require.NoError(t, err)
	finalBlockIndex := uint32(bn)

	t.Run("the block number is correct", func(t *testing.T) {
		bn, err = lightClient.BlockNumber(context.Background())
		require.NoError(t, err)
		require.GreaterOrEqual(t, uint64(finalBlockIndex), bn)
	})

	t.Run("eth_getlogs", func(t *testing.T) {
		t.Parallel()
		filterQuery := ethereum.FilterQuery{
			Addresses: []common.Address{storageContractAddr},
			FromBlock: big.NewInt(int64(initialBlockIndex + 1)),
			ToBlock:   big.NewInt(int64(finalBlockIndex)),
		}

		// archive node
		logs, err := archiveClient.FilterLogs(context.Background(), filterQuery)
		require.NoError(t, err)
		require.Len(t, logs, numRequests)

		// retry the same query on a light node
		_, err = lightClient.FilterLogs(context.Background(), filterQuery)
		require.Error(t, err)
		testmisc.RequireErrorToBe(t, err, "trie root not found")
	})

	t.Run("eth_call", func(t *testing.T) {
		t.Skip("Calling a contract with an old block number fails")
		t.Parallel()
		contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
		require.NoError(t, err)

		callData, err := contractABI.Pack("retrieve")
		require.NoError(t, err)

		callMsg := ethereum.CallMsg{To: &storageContractAddr, Data: callData}
		ret, err := archiveClient.CallContract(context.Background(), callMsg, big.NewInt(int64(initialBlockIndex)))
		require.NoError(t, err)
		out, err := contractABI.Unpack("retrieve", ret)
		require.NoError(t, err)
		counter := out[0].(uint32)
		require.Greater(t, counter, 10)

		callMsg = ethereum.CallMsg{To: &storageContractAddr, Data: callData}
		ret, err = lightClient.CallContract(context.Background(), callMsg, big.NewInt(int64(initialBlockIndex)))
		require.NoError(t, err)
		out, err = contractABI.Unpack("retrieve", ret)
		require.Error(t, err)
		testmisc.RequireErrorToBe(t, err, "does not exist")
	})

	t.Run("eth_getBlockByNumber eth_getBlockByHash", func(t *testing.T) {
		t.Parallel()
		assertLightClient := func(i uint32, block *types.Block, err error) {
			if i <= finalBlockIndex-uint32(blockKeepAmount) {
				// older blocks are not available anymore
				require.Error(t, err)
				require.Nil(t, block)
			} else {
				require.NoError(t, err)
				require.NotNil(t, block)
			}
		}
		// check all blocks are reachable
		for i := uint32(0); i <= finalBlockIndex; i++ {
			block, err := archiveClient.BlockByNumber(context.Background(), big.NewInt(int64(i)))
			require.NoError(t, err)
			require.NotNil(t, block)

			blockLightClient, err := lightClient.BlockByNumber(context.Background(), big.NewInt(int64(i)))
			assertLightClient(i, blockLightClient, err)

			blockByHash, err := archiveClient.BlockByHash(context.Background(), block.Hash())
			require.NoError(t, err)
			require.NotNil(t, blockByHash)

			blockByHashLightClient, err := lightClient.BlockByHash(context.Background(), block.Hash())
			assertLightClient(i, blockByHashLightClient, err)
		}
	})

	t.Run(`
	eth_getTransactionByBlockHashAndIndex
	eth_getBlockTransactionCountByHash
	`, func(t *testing.T) {
		t.Parallel()
		block, err := archiveClient.BlockByNumber(context.Background(), big.NewInt(30))
		require.NoError(t, err)
		txCount, err := archiveClient.TransactionCount(context.Background(), block.Hash())
		require.NoError(t, err)
		require.GreaterOrEqual(t, txCount, uint(1))

		tx, err := archiveClient.TransactionInBlock(context.Background(), block.Hash(), 0)
		require.NoError(t, err)
		require.NotNil(t, tx)
	})

	t.Run("eth_getTransactionByHash", func(t *testing.T) {
		t.Parallel()
		tx, _, err := archiveClient.TransactionByHash(context.Background(), txs[10].Hash())
		require.NotNil(t, tx)
		require.NoError(t, err)

		tx, _, err = lightClient.TransactionByHash(context.Background(), txs[10].Hash())
		require.Error(t, err)
	})

	t.Run("eth_getBalance", func(t *testing.T) {
		t.Parallel()
		bal, err := archiveClient.BalanceAt(context.Background(), env.testContractEnv.EvmTesterAddr, big.NewInt(25))
		require.NoError(t, err)
		require.Positive(t, bal.Cmp(big.NewInt(0)))
	})

	t.Run("eth_getCode", func(t *testing.T) {
		t.Parallel()
		code, err := archiveClient.CodeAt(context.Background(), env.testContractEnv.EvmTesterAddr, big.NewInt(25))
		require.NoError(t, err)
		require.NotNil(t, code)
	})

	t.Run("eth_getTransactionReceipt", func(t *testing.T) {
		t.Parallel()
		rec, err := archiveClient.TransactionReceipt(context.Background(), txs[42].Hash())
		require.NoError(t, err)
		require.NotNil(t, rec)
	})

	t.Run("eth_getStorageAt", func(t *testing.T) {
		t.Parallel()
		val, err := archiveClient.StorageAt(context.Background(), env.testContractEnv.EvmTestContractAddr, common.BigToHash(big.NewInt(0)), big.NewInt(55))
		require.NoError(t, err)
		require.NotNil(t, val)
	})

	t.Run("isc view call", func(t *testing.T) {
		t.Parallel()
		// archive node
		var bi uint32 = 10
		res, err := chain.Client(nil, 0).CallView(
			context.Background(),
			blocklog.ViewGetRequestReceiptsForBlock.Message(&bi),
			"10",
		)
		require.NoError(t, err)
		receipts, err := blocklog.ViewGetRequestReceiptsForBlock.DecodeOutput(res)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(receipts.Receipts), 1)
		require.NoError(t, err)
		require.NotZero(t, receipts.Receipts[0].GasFeeCharged)

		// light node
		bi = 0
		_, err = chain.Client(nil, 1).CallView(
			context.Background(),
			blocklog.ViewGetRequestReceiptsForBlock.Message(&bi),
			"10",
		)
		require.Error(t, err)
	})
}
