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

	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/tools/cluster/templates"
)

func TestPruning(t *testing.T) {
	clu := newCluster(t, waspClusterOpts{
		nNodes: 4,
		modifyConfig: func(nodeIndex int, configParams templates.WaspConfigParams) templates.WaspConfigParams {
			// set node 0 as an "archive node"
			if nodeIndex == 0 {
				configParams.PruningMinStatesToKeep = -1
			} else {
				// all other nodes will only keep 10 blocks
				configParams.PruningMinStatesToKeep = 10
			}

			return configParams
		},
	})

	// set blockKeepAmount to 10 as well
	chain, err := clu.DeployChainWithDKG(clu.Config.AllNodes(), clu.Config.AllNodes(), 4, 10)
	require.NoError(t, err)
	env := newChainEnv(t, clu, chain)

	// let's send 100 EVM requests (wait for each request individually, so that the chain height increases as much as possible)
	const numRequests = 100

	// deposit funds for EVM
	keyPair, _, err := clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	evmPvtKey, evmAddr := solo.NewEthereumAccount()
	evmAgentID := isc.NewEthereumAddressAgentID(evmAddr)
	env.TransferFundsTo(isc.NewAssetsBaseTokens(utxodb.FundsFromFaucetAmount-1*isc.Million), nil, keyPair, evmAgentID)

	// deploy solidity inccounter
	storageContractAddr, storageContractABI := env.DeploySolidityContract(evmPvtKey, evmtest.StorageContractABI, evmtest.StorageContractBytecode, uint32(42))

	initialBlockIndex, err := env.Chain.BlockIndex()
	require.NoError(t, err)

	jsonRPCClient := env.EVMJSONRPClient(0) // send request to node #0
	nonce := env.GetNonceEVM(evmAddr)
	for i := uint64(0); i < numRequests; i++ {
		// send tx to change the stored value
		callArguments, err2 := storageContractABI.Pack("store", uint32(i))
		require.NoError(t, err2)
		tx, err2 := types.SignTx(
			types.NewTransaction(nonce+i, storageContractAddr, big.NewInt(0), 100000, evm.GasPrice, callArguments),
			EVMSigner(),
			evmPvtKey,
		)
		require.NoError(t, err2)
		err2 = jsonRPCClient.SendTransaction(context.Background(), tx)
		require.NoError(t, err2)
		// await tx confirmed
		_, err2 = clu.MultiClient().WaitUntilEVMRequestProcessedSuccessfully(env.Chain.ChainID, tx.Hash(), false, 5*time.Second)
		require.NoError(t, err2)
	}

	jsonRPCClientLightNode := env.EVMJSONRPClient(1) // send request to node #0
	finalBlockIndex := initialBlockIndex + numRequests

	// assert the block number is correct
	{
		// archive node
		bn, err := jsonRPCClient.BlockNumber(context.Background())
		require.NoError(t, err)
		require.EqualValues(t, finalBlockIndex, bn)
		// light node
		bn, err = jsonRPCClientLightNode.BlockNumber(context.Background())
		require.NoError(t, err)
		require.EqualValues(t, finalBlockIndex, bn)
	}

	// test eth_getlogs
	{
		filterQuery := ethereum.FilterQuery{
			Addresses: []common.Address{storageContractAddr},
			FromBlock: big.NewInt(int64(initialBlockIndex + 1)),
			ToBlock:   big.NewInt(int64(finalBlockIndex)),
		}

		// archive node
		logs, err := jsonRPCClient.FilterLogs(context.Background(), filterQuery)
		require.NoError(t, err)
		require.Len(t, logs, numRequests)

		// retry the same query on a light node
		logs, err = jsonRPCClientLightNode.FilterLogs(context.Background(), filterQuery)
		require.NoError(t, err)
		require.Len(t, logs, 10) // TODO should this return 10 or an error?
	}
}
