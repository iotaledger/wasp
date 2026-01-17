package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
	"github.com/iotaledger/wasp/v2/tools/cluster"
)

func TestValidatorFees(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cluster tests in short mode")
	}

	validatorKps := []*cryptolib.KeyPair{
		cryptolib.NewKeyPair(),
		cryptolib.NewKeyPair(),
		cryptolib.NewKeyPair(),
		cryptolib.NewKeyPair(),
	}
	// set custom addresses for each validator
	modifyConfig := func(nodeIndex int, configParams cluster.WaspConfigParams) cluster.WaspConfigParams {
		configParams.ValidatorKeyPair = validatorKps[nodeIndex]
		configParams.ValidatorAddress = validatorKps[nodeIndex].Address().String() // privtangle prefix
		return configParams
	}
	clu := newCluster(t, waspClusterOpts{nNodes: 4, modifyConfig: modifyConfig})
	cmt := []int{0, 1, 2, 3}
	chain, err := clu.DeployChainWithDistKeyGen(cmt, cmt, 4)
	require.NoError(t, err)
	chainID := chain.ChainID

	chEnv := newChainEnv(t, clu, chain)

	// set validator split fees to 50/50
	newGasFeePolicy := &gas.FeePolicy{
		GasPerToken:       util.Ratio32{A: 1, B: 10},
		ValidatorFeeShare: 50,
		EVMGasRatio:       gas.DefaultEVMGasRatio,
	}
	govClient := chain.Client(chain.OriginatorKeyPair)
	reqTx, err := govClient.PostRequest(context.Background(), governance.FuncSetFeePolicy.Message(newGasFeePolicy), chainclient.PostRequestParams{
		Transfer:  isc.NewAssets(iotagraphql.DefaultGasBudget + 10),
		GasBudget: iotagraphql.DefaultGasBudget,
	})
	require.NoError(t, err)
	_, err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), chain.ChainID, reqTx, false, 30*time.Second)
	require.NoError(t, err)

	// send a bunch of requests
	// assert each validator has received fees
	userWallet, _, err := chEnv.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	scClient := chainclient.New(clu.L1Client(), clu.WaspClient(0), chainID, userWallet)
	for i := 0; i < 20; i++ {
		reqTx, err := scClient.PostRequest(context.Background(), accounts.FuncDeposit.Message(), chainclient.PostRequestParams{
			Transfer:  isc.NewAssets(iotagraphql.DefaultGasBudget + 100),
			GasBudget: iotagraphql.DefaultGasBudget,
		})
		require.NoError(t, err)
		_, err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), chainID, reqTx, false, 30*time.Second)
		require.NoError(t, err)
	}
	for _, validatorKp := range validatorKps {
		tokens := chEnv.getBalanceOnChain(isc.NewAddressAgentID(validatorKp.Address()), coin.BaseTokenType)
		require.NotZero(t, tokens)
	}
}
