package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func TestMaintenance(t *testing.T) {
	env := setupAdvancedInccounterTest(t, 4, []int{0, 1, 2, 3})

	ownerWallet, ownerAddr, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	ownerAgentID := isc.NewAgentID(ownerAddr)
	env.DepositFunds(10*isc.Million, ownerWallet)
	ownerSCClient := env.Chain.SCClient(governance.Contract.Hname(), ownerWallet)
	ownerIncCounterSCClient := env.Chain.SCClient(nativeIncCounterSCHname, ownerWallet)

	userWallet, _, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	env.DepositFunds(10*isc.Million, userWallet)
	userSCClient := env.Chain.SCClient(governance.Contract.Hname(), userWallet)
	userIncCounterSCClient := env.Chain.SCClient(nativeIncCounterSCHname, userWallet)

	// set owner of the chain
	{
		originatorSCClient := env.Chain.SCClient(governance.Contract.Hname(), env.Chain.OriginatorKeyPair)
		tx, err := originatorSCClient.PostRequest(governance.FuncDelegateChainOwnership.Name, chainclient.PostRequestParams{
			Args: dict.Dict{
				governance.ParamChainOwner: codec.Encode(ownerAgentID),
			},
		})
		require.NoError(t, err)
		_, err = env.Clu.MultiClient().WaitUntilAllRequestsProcessedSuccessfully(env.Chain.ChainID, tx, 10*time.Second)
		require.NoError(t, err)

		req, err := ownerSCClient.PostOffLedgerRequest(governance.FuncClaimChainOwnership.Name)
		require.NoError(t, err)
		_, err = env.Clu.MultiClient().WaitUntilRequestProcessedSuccessfully(env.Chain.ChainID, req.ID(), 10*time.Second)
		require.NoError(t, err)
	}

	// call the gov "maintenance status view", check it is OFF
	{
		ret, err := ownerSCClient.CallView(governance.ViewGetMaintenanceStatus.Name, nil)
		require.NoError(t, err)
		maintenanceStatus := codec.MustDecodeBool(ret.MustGet(governance.VarMaintenanceStatus))
		require.False(t, maintenanceStatus)
	}

	// test non-chain owner cannot call init maintenance
	{
		req, err := userSCClient.PostOffLedgerRequest(governance.FuncStartMaintenance.Name)
		require.NoError(t, err)
		rec, err := env.Clu.MultiClient().WaitUntilRequestProcessed(env.Chain.ChainID, req.ID(), 10*time.Second)
		require.NoError(t, err)
		require.Error(t, rec.Error)
	}

	// owner can start maintenance mode
	{
		req, err := ownerSCClient.PostOffLedgerRequest(governance.FuncStartMaintenance.Name)
		require.NoError(t, err)
		_, err = env.Clu.MultiClient().WaitUntilRequestProcessedSuccessfully(env.Chain.ChainID, req.ID(), 10*time.Second)
		require.NoError(t, err)
	}

	// call the gov "maintenance status view", check it is ON
	{
		ret, err := ownerSCClient.CallView(governance.ViewGetMaintenanceStatus.Name, nil)
		require.NoError(t, err)
		maintenanceStatus := codec.MustDecodeBool(ret.MustGet(governance.VarMaintenanceStatus))
		require.True(t, maintenanceStatus)
	}

	// get the current block number
	blockIndex, err := env.Chain.BlockIndex()
	require.NoError(t, err)

	// calls to non-maintenance endpoints are not processed
	notProccessedReq1, err := userIncCounterSCClient.PostOffLedgerRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)
	time.Sleep(10 * time.Second) // not ideal, but I don't think there is a good way to wait for something that will NOT be processed
	rec, err := env.Chain.GetRequestReceipt(notProccessedReq1.ID())
	require.NoError(t, err)
	require.Nil(t, rec)

	// calls to non-maintenance endpoints are not processed, even when done by the chain owner
	notProccessedReq2, err := ownerIncCounterSCClient.PostOffLedgerRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)
	time.Sleep(10 * time.Second) // not ideal, but I don't think there is a good way to wait for something that will NOT be processed
	rec, err = env.Chain.GetRequestReceipt(notProccessedReq2.ID())
	require.NoError(t, err)
	require.Nil(t, rec)

	// assert that block number is still the same
	blockIndex2, err := env.Chain.BlockIndex()
	require.NoError(t, err)
	require.EqualValues(t, blockIndex, blockIndex2)

	// calls to governance are processed (try changing fees for example)
	newGasFeePolicy := gas.GasFeePolicy{
		GasFeeTokenID:     nil,
		GasPerToken:       10,
		ValidatorFeeShare: 1,
	}
	{
		req, err := ownerSCClient.PostOffLedgerRequest(governance.FuncSetFeePolicy.Name, chainclient.PostRequestParams{
			Args: dict.Dict{
				governance.ParamFeePolicyBytes: newGasFeePolicy.Bytes(),
			},
		})
		require.NoError(t, err)
		_, err = env.Clu.MultiClient().WaitUntilRequestProcessedSuccessfully(env.Chain.ChainID, req.ID(), 10*time.Second)
		require.NoError(t, err)
	}

	// calls to governance from non-owners should be processed, but fail
	{
		req, err := userSCClient.PostOffLedgerRequest(governance.FuncSetFeePolicy.Name, chainclient.PostRequestParams{
			Args: dict.Dict{
				governance.ParamFeePolicyBytes: newGasFeePolicy.Bytes(),
			},
		})
		require.NoError(t, err)
		receipt, err := env.Clu.MultiClient().WaitUntilRequestProcessed(env.Chain.ChainID, req.ID(), 10*time.Second)
		require.NoError(t, err)
		require.Error(t, receipt.Error)
	}

	// test non-chain owner cannot call stop maintenance
	{
		req, err := userSCClient.PostOffLedgerRequest(governance.FuncStopMaintenance.Name)
		require.NoError(t, err)
		rec, err := env.Clu.MultiClient().WaitUntilRequestProcessed(env.Chain.ChainID, req.ID(), 10*time.Second)
		require.NoError(t, err)
		require.Error(t, rec.Error)
	}

	// owner can stop maintenance mode
	{
		req, err := ownerSCClient.PostOffLedgerRequest(governance.FuncStopMaintenance.Name)
		require.NoError(t, err)
		_, err = env.Clu.MultiClient().WaitUntilRequestProcessedSuccessfully(env.Chain.ChainID, req.ID(), 10*time.Second)
		require.NoError(t, err)
	}

	// normal requests are now processed successfully (pending requests issued during maintenance should be processed now)
	{
		_, err = env.Clu.MultiClient().WaitUntilRequestProcessedSuccessfully(env.Chain.ChainID, notProccessedReq1.ID(), 10*time.Second)
		require.NoError(t, err)
		_, err = env.Clu.MultiClient().WaitUntilRequestProcessedSuccessfully(env.Chain.ChainID, notProccessedReq2.ID(), 10*time.Second)
		require.NoError(t, err)
		require.EqualValues(t, 2, env.getNativeContractCounter(nativeIncCounterSCHname))
	}
}
