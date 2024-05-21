package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/tools/cluster/templates"
)

func TestValidatorFees(t *testing.T) {
	validatorKps := []*cryptolib.KeyPair{
		cryptolib.NewKeyPair(),
		cryptolib.NewKeyPair(),
		cryptolib.NewKeyPair(),
		cryptolib.NewKeyPair(),
	}
	// set custom addresses for each validator
	modifyConfig := func(nodeIndex int, configParams templates.WaspConfigParams) templates.WaspConfigParams {
		configParams.ValidatorKeyPair = validatorKps[nodeIndex]
		configParams.ValidatorAddress = validatorKps[nodeIndex].Address().Bech32("atoi") // privtangle prefix
		return configParams
	}
	clu := newCluster(t, waspClusterOpts{nNodes: 4, modifyConfig: modifyConfig})
	cmt := []int{0, 1, 2, 3}
	chain, err := clu.DeployChainWithDKG(cmt, cmt, 4)
	require.NoError(t, err)
	chainID := chain.ChainID

	chEnv := newChainEnv(t, clu, chain)
	chEnv.deployNativeIncCounterSC()
	waitUntil(t, chEnv.contractIsDeployed(), clu.Config.AllNodes(), 30*time.Second)

	// set validator split fees to 50/50
	{
		originatorSCClient := chain.Client(chain.OriginatorKeyPair)
		newGasFeePolicy := &gas.FeePolicy{
			GasPerToken:       util.Ratio32{A: 1, B: 10},
			ValidatorFeeShare: 50,
			EVMGasRatio:       gas.DefaultEVMGasRatio,
		}
		req, err2 := originatorSCClient.PostOffLedgerRequest(
			context.Background(),
			governance.FuncSetFeePolicy.Message(newGasFeePolicy),
			chainclient.PostRequestParams{Nonce: 0},
		)
		require.NoError(t, err2)
		_, err2 = clu.MultiClient().WaitUntilRequestProcessedSuccessfully(chain.ChainID, req.ID(), false, 30*time.Second)
		require.NoError(t, err2)
	}
	// send a bunch of requests

	// assert each validator has received fees
	userWallet, _, err := chEnv.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	scClient := chainclient.New(clu.L1Client(), clu.WaspClient(0), chainID, userWallet)
	for i := 0; i < 20; i++ {
		reqTx, err := scClient.PostRequest(inccounter.FuncIncCounter.Message(nil))
		require.NoError(t, err)
		_, err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(chainID, reqTx, false, 30*time.Second)
		require.NoError(t, err)
	}
	for _, validatorKp := range validatorKps {
		tokens := chEnv.getBalanceOnChain(isc.NewAgentID(validatorKp.Address()), isc.BaseTokenID)
		require.NotZero(t, tokens)
	}
}
