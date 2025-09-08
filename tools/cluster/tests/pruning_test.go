package tests

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/v2/tools/cluster"
)

func TestPruning(t *testing.T) {
	t.Skip("TODO: fix test")
	t.Parallel()
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
	chain, err := clu.DeployChainWithDKG(clu.Config.AllNodes(), clu.Config.AllNodes(), 4, int32(blockKeepAmount))
	require.NoError(t, err)
	env := newChainEnv(t, clu, chain)

	var baseTokensToWithdraw coin.Value = 100
	chClient := env.NewChainClient(env.Chain.OriginatorKeyPair)
	req, err := chClient.PostOffLedgerRequest(context.Background(), accounts.FuncWithdraw.Message(),
		chainclient.PostRequestParams{
			Allowance: isc.NewAssets(baseTokensToWithdraw),
		},
	)
	require.NoError(t, err)
	_, err = chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(context.Background(), req.ID(), true, 30*time.Second)
	require.NoError(t, err)

	// let's send 100 EVM requests (wait for each request individually, so that the chain height increases as much as possible)
	const numRequests = 100

	initialBlockIndex, err := env.Chain.BlockIndex()
	require.NoError(t, err)

	archiveClient := env.EVMJSONRPClient(0)
	lightClient := env.EVMJSONRPClient(1)

	txs := make([]*types.Transaction, numRequests)
	for i := uint64(0); i < numRequests; i++ {
		tx := env.CallStore(archiveClient, lightClient, i)
		txs[i] = tx
		time.Sleep(10 * time.Second)
	}

	finalBlockIndex := initialBlockIndex + numRequests

	t.Run("the block number is correct", func(t *testing.T) {
		t.Parallel()
		// archive node
		bn, err := archiveClient.BlockNumber(context.Background())
		require.NoError(t, err)
		require.EqualValues(t, finalBlockIndex, bn)
		// light node
		bn, err = lightClient.BlockNumber(context.Background())
		require.NoError(t, err)
		require.EqualValues(t, finalBlockIndex, bn)
	})

	t.Run("eth_getlogs", func(t *testing.T) {
		t.Parallel()
		filterQuery := ethereum.FilterQuery{
			Addresses: []common.Address{env.testContractEnv.EvmTestContractAddr},
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
		testmisc.RequireErrorToBe(t, err, "does not exist")
	})

	t.Run("eth_call", func(t *testing.T) {
		t.Parallel()
		callArgs, err := env.testContractEnv.EvmTestContractABI.Pack("retrieve")
		require.NoError(t, err)
		callMsg := ethereum.CallMsg{
			To:   &env.testContractEnv.EvmTestContractAddr,
			Data: callArgs,
		}
		ret, err := archiveClient.CallContract(context.Background(), callMsg, big.NewInt(50))
		require.NoError(t, err)
		val, err := env.testContractEnv.EvmTestContractABI.Unpack("retrieve", ret)
		require.NoError(t, err)
		require.EqualValues(t, 50-initialBlockIndex-1, val[0].(uint32))
		_, err = lightClient.CallContract(context.Background(), callMsg, big.NewInt(50))
		require.Error(t, err)
		testmisc.RequireErrorToBe(t, err, "does not exist")
	})

	t.Run("eth_getBlockByNumber eth_getBlockByHash eth_getTransactionCount", func(t *testing.T) {
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

			txCount, err := archiveClient.TransactionCount(context.Background(), block.Hash())
			require.NoError(t, err)
			if i >= initialBlockIndex {
				require.EqualValues(t, 1, txCount) // in this particular test, we make sure only 1 tx exists per block
			}
		}
	})

	t.Run(`
	eth_getTransactionByBlockNumberAndIndex
	eth_getTransactionByBlockHashAndIndex
	eth_getBlockTransactionCountByHash
	eth_getBlockTransactionCountByNumber
	`, func(t *testing.T) {
		t.Parallel()
		// eth_getTransactionByBlockNumberAndIndex and eth_getBlockTransactionCountByNumber are not exposed in ethclient.Client
		block, err := archiveClient.BlockByNumber(context.Background(), big.NewInt(30))
		require.NoError(t, err)
		txCount, err := archiveClient.TransactionCount(context.Background(), block.Hash())
		require.NoError(t, err)
		require.EqualValues(t, 1, txCount)

		tx, err := archiveClient.TransactionInBlock(context.Background(), block.Hash(), 0)
		require.NoError(t, err)
		require.NotNil(t, tx)
	})

	t.Run("eth_getTransactionByHash", func(t *testing.T) {
		t.Parallel()
		tx, _, err := archiveClient.TransactionByHash(context.Background(), txs[10].Hash())
		require.NotNil(t, tx)
		require.NoError(t, err)
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
		require.Len(t, receipts.Receipts, 1)
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
