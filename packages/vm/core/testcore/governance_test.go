// excluded temporarily because of compilation errors

package testcore

import (
	"reflect"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/testutil/testdbhash"
	"github.com/iotaledger/wasp/v2/packages/transaction"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance/governanceimpl"
	"github.com/iotaledger/wasp/v2/packages/vm/core/testcore/contracts/inccounter"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

func TestGovernanceAccessNodes(t *testing.T) {
	env := solo.New(t)
	node1KP, _ := env.NewKeyPairWithFunds(env.NewSeedFromTestNameAndTimestamp(t.Name()))
	node1OwnerKP, node1OwnerAddr := env.NewKeyPairWithFunds(env.NewSeedFromTestNameAndTimestamp(t.Name()))
	chainKP, _ := env.NewKeyPairWithFunds(env.NewSeedFromTestNameAndTimestamp(t.Name()))
	chain, _ := env.NewChainExt(chainKP, 0, "chain1", 0, 0)

	//
	// Initially the state is empty.
	res, err := chain.CallView(governance.ViewGetChainNodes.Message())
	require.NoError(t, err)
	candidates, accessNodes := lo.Must2(governance.ViewGetChainNodes.DecodeOutput(res))
	require.Empty(t, candidates)
	require.Empty(t, accessNodes)

	//
	// Add a single access node candidate.
	_, err = chain.PostRequestSync(
		solo.NewCallParams(governance.FuncAddCandidateNode.Message(
			node1KP.GetPublicKey(),
			governance.NewNodeOwnershipCertificate(node1KP, node1OwnerAddr).Bytes(),
			"http://my-api/url",
			false,
		)).AddBaseTokens(iotaclient.DefaultGasBudget),
		node1OwnerKP, // Sender should match data used to create the Cert field value.
	)
	require.NoError(t, err)

	testdbhash.VerifyContractStateHash(env, governance.Contract, "", t.Name()+"1")

	res, err = chain.CallView(governance.ViewGetChainNodes.Message())
	require.NoError(t, err)
	candidates, accessNodes = lo.Must2(governance.ViewGetChainNodes.DecodeOutput(res))
	require.Len(t, candidates, 1) // Candidate registered.
	require.Equal(t, node1KP.GetPublicKey().Bytes(), candidates[0].NodePubKey.Bytes())
	require.Equal(t, "http://my-api/url", candidates[0].AccessAPI)
	require.Empty(t, accessNodes)

	//
	// Accept the node as an access node.
	_, err = chain.PostRequestSync(
		solo.NewCallParams(governance.FuncChangeAccessNodes.Message(
			governance.ChangeAccessNodeActions{
				governance.AcceptAccessNodeAction(node1KP.GetPublicKey()),
			},
		)).AddBaseTokens(iotaclient.DefaultGasBudget),
		chainKP,
	)
	require.NoError(t, err)

	testdbhash.VerifyContractStateHash(env, governance.Contract, "", t.Name()+"2")

	res, err = chain.CallView(governance.ViewGetChainNodes.Message())
	require.NoError(t, err)
	candidates, accessNodes = lo.Must2(governance.ViewGetChainNodes.DecodeOutput(res))
	require.Len(t, candidates, 1)
	require.Len(t, accessNodes, 1)
	require.Equal(t, node1KP.GetPublicKey().Bytes(), accessNodes[0].Bytes())

	//
	// Revoke the access node (by the node owner).
	_, err = chain.PostRequestSync(
		solo.NewCallParams(governance.FuncRevokeAccessNode.Message(
			node1KP.GetPublicKey(),
			governance.NewNodeOwnershipCertificate(node1KP, node1OwnerAddr).Bytes(),
		)).WithMaxAffordableGasBudget(),
		node1OwnerKP, // Sender should match data used to create the Cert field value.
	)
	require.NoError(t, err)

	res, err = chain.CallView(governance.ViewGetChainNodes.Message())
	require.NoError(t, err)
	candidates, accessNodes = lo.Must2(governance.ViewGetChainNodes.DecodeOutput(res))
	require.Empty(t, candidates)
	require.Empty(t, accessNodes)
}

func TestMaintenanceMode(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain()

	ownerWallet, ownerAddr := env.NewKeyPairWithFunds(env.NewSeedFromTestNameAndTimestamp(t.Name()))
	ownerAgentID := isc.NewAddressAgentID(ownerAddr)
	ch.DepositBaseTokensToL2(10*isc.Million, ownerWallet)

	userWallet, _ := env.NewKeyPairWithFunds(env.NewSeedFromTestNameAndTimestamp(t.Name()))
	ch.DepositBaseTokensToL2(10*isc.Million, userWallet)

	// set owner of the chain
	{
		_, err2 := ch.PostRequestSync(
			solo.NewCallParams(governance.FuncDelegateChainAdmin.Message(ownerAgentID)).
				WithMaxAffordableGasBudget(),
			nil,
		)
		require.NoError(t, err2)

		testdbhash.VerifyContractStateHash(env, governance.Contract, "", t.Name())

		_, err2 = ch.PostRequestSync(
			solo.NewCallParams(governance.FuncClaimChainAdmin.Message()).WithMaxAffordableGasBudget(),
			ownerWallet,
		)
		require.NoError(t, err2)
	}

	// call the gov "maintenance status view", check it is OFF
	{
		// TODO: Add maintenance status to wrapped core contracts
		ret, err2 := ch.CallView(governance.ViewGetMaintenanceStatus.Message())
		require.NoError(t, err2)
		maintenanceStatus := lo.Must(governance.ViewGetMaintenanceStatus.DecodeOutput(ret))
		require.False(t, maintenanceStatus)
	}

	// test non-chain admin cannot call init maintenance
	{
		_, err2 := ch.PostRequestSync(
			solo.NewCallParams(governance.FuncStartMaintenance.Message()).WithMaxAffordableGasBudget(),
			userWallet,
		)
		require.ErrorContains(t, err2, "unauthorized")
	}

	// owner can start maintenance mode
	{
		_, err2 := ch.PostRequestSync(
			solo.NewCallParams(governance.FuncStartMaintenance.Message()).WithMaxAffordableGasBudget(),
			ownerWallet,
		)
		require.NoError(t, err2)
	}

	// call the gov "maintenance status view", check it is ON
	{
		ret, err2 := ch.CallView(governance.ViewGetMaintenanceStatus.Message())
		require.NoError(t, err2)
		maintenanceStatus := lo.Must(governance.ViewGetMaintenanceStatus.DecodeOutput(ret))
		require.True(t, maintenanceStatus)
	}

	var reqs []isc.OnLedgerRequest
	{
		for _, wallet := range []*cryptolib.KeyPair{userWallet, ownerWallet} {
			req, _, err := ch.SendRequest(solo.NewCallParams(inccounter.FuncIncCounter.Message(nil)).
				WithMaxAffordableGasBudget(), wallet)
			require.NoError(t, err)
			reqs = append(reqs, req)
		}
	}

	// requests are skipped
	_, res := ch.RunRequestBatch(2)
	require.Empty(t, res)
	for _, req := range reqs {
		require.False(t, ch.IsRequestProcessed(req.ID()))
	}

	fp := &gas.FeePolicy{
		GasPerToken:       util.Ratio32{A: 1, B: 10},
		ValidatorFeeShare: 1,
		EVMGasRatio:       gas.DefaultEVMGasRatio,
	}

	// calls to governance are processed (try changing fees for example)
	{
		_, err2 := ch.PostRequestSync(solo.NewCallParams(
			governance.FuncSetFeePolicy.Message(fp),
		), ownerWallet)
		require.NoError(t, err2)
	}

	// calls to governance from non-owners should be processed, but fail
	{
		_, err2 := ch.PostRequestSync(solo.NewCallParams(
			governance.FuncSetFeePolicy.Message(fp),
		), userWallet)
		require.ErrorContains(t, err2, "unauthorized")
	}

	// test non-chain admin cannot call stop maintenance
	{
		_, err2 := ch.PostRequestSync(
			solo.NewCallParams(governance.FuncStopMaintenance.Message()).WithMaxAffordableGasBudget(),
			userWallet,
		)
		require.ErrorContains(t, err2, "unauthorized")
	}

	// requests are still skipped
	_, res = ch.RunRequestBatch(2)
	require.Empty(t, res)
	for _, req := range reqs {
		require.False(t, ch.IsRequestProcessed(req.ID()))
	}

	// owner can stop maintenance mode
	{
		_, err2 := ch.PostRequestSync(
			solo.NewCallParams(governance.FuncStopMaintenance.Message()).WithMaxAffordableGasBudget(),
			ownerWallet,
		)
		require.NoError(t, err2)
	}

	// normal requests are now processed successfully (pending requests issued during maintenance should be processed now)
	ch.RunAllReceivedRequests(2)
	for _, req := range reqs {
		require.True(t, ch.IsRequestProcessed(req.ID()))
	}
}

func TestGovernanceMetadata(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain(true)

	/*
		Values with the length == 0 will reset the state value
		Values with the length > 0 will set the state value
		Nil values will be ignored and not change the state value
	*/

	testValue := "TESTYTEST"

	testMetadata := &isc.PublicChainMetadata{
		EVMJsonRPCURL:   testValue,
		EVMWebSocketURL: testValue,
		Name:            testValue,
		Description:     testValue,
		Website:         testValue,
	}

	// Set proper metadata value
	_, err := ch.PostRequestOffLedger(
		solo.NewCallParams(
			governance.FuncSetMetadata.Message(nil, &testMetadata),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	testdbhash.VerifyContractStateHash(env, governance.Contract, "", t.Name())

	res, err := ch.CallView(governance.ViewGetMetadata.Message())
	require.NoError(t, err)
	_, resMetadata := lo.Must2(governance.ViewGetMetadata.DecodeOutput(res))

	// Chain name should be equal to the configured one.
	require.Equal(t, testMetadata.Bytes(), resMetadata.Bytes())

	// Call SetMetadata without args. The metadata should be the same as it was previously configured and not be emptied.
	_, err = ch.PostRequestOffLedger(
		solo.NewCallParams(governance.FuncSetMetadata.Message(nil, nil)).
			WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	res, err = ch.CallView(governance.ViewGetMetadata.Message())
	require.NoError(t, err)
	_, resMetadata = lo.Must2(governance.ViewGetMetadata.DecodeOutput(res))

	// Chain name should be equal to the configured one.
	require.Equal(t, testMetadata.Bytes(), resMetadata.Bytes())

	// Call SetMetadata with an empty arg. The metadata call should fail.
	_, err = ch.PostRequestOffLedger(
		solo.NewCallParamsEx(
			governance.Contract.Name,
			governance.FuncSetMetadata.Name,
			isc.NewCallArguments([]byte{}),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.Error(t, err)

	// Test invalid custom metadata
	_, err = ch.PostRequestOffLedger(
		solo.NewCallParamsEx(
			governance.Contract.Name,
			governance.FuncSetMetadata.Name,
			isc.NewCallArguments(
				make([]byte, governanceimpl.MaxCustomMetadataLength+1),
			),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.Error(t, err)

	// set invalid custom metadata
	hugePublicChainMetadata := &isc.PublicChainMetadata{
		Website: string(make([]byte, governanceimpl.MaxCustomMetadataLength+1)),
	}
	_, err = ch.PostRequestOffLedger(
		solo.NewCallParams(
			governance.FuncSetMetadata.Message(nil, &hugePublicChainMetadata),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.Error(t, err)
}

func TestGovernanceL1Metadata(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain()

	// deposit some extra tokens to the common account to accommodate for the SD change
	ch.SendFromL1ToL2AccountBaseTokens(solo.BaseTokensForL2Gas, solo.BaseTokensForL2Gas, accounts.CommonAccount(), nil)

	// set max valid size custom metadata
	publicURLMetadata := "https://iota.org"

	_, err := ch.PostRequestSync(
		solo.NewCallParams(
			governance.FuncSetMetadata.Message(&publicURLMetadata, nil),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	testdbhash.VerifyContractStateHash(env, governance.Contract, "", t.Name())

	// assert metadata is correct on view call
	res, err := ch.CallView(governance.ViewGetMetadata.Message())
	require.NoError(t, err)
	resPubURL, _ := lo.Must2(governance.ViewGetMetadata.DecodeOutput(res))
	require.Equal(t, publicURLMetadata, resPubURL)

	// assert metadata is correct on L1 alias output
	anchor := ch.GetLatestAnchor()
	require.NoError(t, err)
	sm, err := transaction.StateMetadataFromBytes(anchor.GetStateMetadata())
	require.NoError(t, err)
	require.Equal(t, publicURLMetadata, sm.PublicURL)
	require.True(t, reflect.DeepEqual(sm.GasFeePolicy, gas.DefaultFeePolicy()))

	// try changing the gas policy
	newFeePolicy := &gas.FeePolicy{
		GasPerToken: util.Ratio32{
			A: 1,
			B: 2,
		},
		EVMGasRatio: util.Ratio32{
			A: 3,
			B: 4,
		},
		ValidatorFeeShare: 5,
	}
	_, err = ch.PostRequestSync(
		solo.NewCallParams(
			governance.FuncSetFeePolicy.Message(newFeePolicy),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	// assert gas policy changed on L1 metadata
	anchor = ch.GetLatestAnchor()
	require.NoError(t, err)
	sm, err = transaction.StateMetadataFromBytes(anchor.GetStateMetadata())
	require.NoError(t, err)
	require.Equal(t, publicURLMetadata, sm.PublicURL)
	require.True(t, reflect.DeepEqual(sm.GasFeePolicy, newFeePolicy))
}

func TestGovernanceGasFee(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{Debug: true, PrintStackTrace: true})
	ch := env.NewChain()
	fp := ch.GetGasFeePolicy()
	fp.GasPerToken.A *= 1000000
	ch.SetGasFeePolicy(nil, fp)
	fp.GasPerToken.A /= 1000000
	ch.SetGasFeePolicy(nil, fp) // should not fail with "gas budget exceeded"
}

func TestGovernanceZeroGasFee(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{Debug: true, PrintStackTrace: true})
	ch := env.NewChain()

	user1, userAddr1 := env.NewKeyPairWithFunds()
	userAgentID1 := isc.NewAddressAgentID(userAddr1)
	_, userAddr2 := env.NewKeyPairWithFunds()
	userAgentID2 := isc.NewAddressAgentID(userAddr2)

	fp := &gas.FeePolicy{
		EVMGasRatio: gas.DefaultEVMGasRatio,
		GasPerToken: util.Ratio32{
			A: 0,
			B: 0,
		},
		ValidatorFeeShare: 1,
	}
	_, err := ch.PostRequestSync(
		solo.NewCallParams(
			governance.FuncSetFeePolicy.Message(fp),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	_, estimate, err := ch.EstimateGasOnLedger(solo.NewCallParams(
		accounts.FuncDeposit.Message(),
	), user1)
	require.NoError(t, err)
	require.Zero(t, estimate.GasFeeCharged)

	userL2Bal1 := ch.L2BaseTokens(userAgentID1)

	gasGreaterThanEstimatedGas := coin.Value(estimate.GasBurned + 100)
	_, err = ch.PostRequestSync(
		solo.NewCallParams(
			accounts.FuncTransferAllowanceTo.Message(userAgentID2),
		).
			AddBaseTokens(gasGreaterThanEstimatedGas).
			AddAllowanceBaseTokens(gasGreaterThanEstimatedGas).
			WithGasBudget(uint64(gasGreaterThanEstimatedGas)),
		user1,
	)
	require.NoError(t, err)

	userL2Bal2 := ch.L2BaseTokens(userAgentID1)
	require.Equal(t, userL2Bal1, userL2Bal2)
	require.Greater(t, ch.LastReceipt().GasBurned, uint64(0))
	require.Zero(t, ch.LastReceipt().GasFeeCharged)

	gasLessThanEstimatedGas := coin.Value(estimate.GasBurned - 100)
	_, err = ch.PostRequestSync(
		solo.NewCallParams(
			accounts.FuncTransferAllowanceTo.Message(userAgentID2),
		).
			AddBaseTokens(gasLessThanEstimatedGas).
			WithGasBudget(uint64(gasLessThanEstimatedGas)),
		user1,
	)
	require.NoError(t, err)

	userL2Bal3 := ch.L2BaseTokens(userAgentID1)
	require.Equal(t, userL2Bal2+gasLessThanEstimatedGas, userL2Bal3)
	require.Greater(t, ch.LastReceipt().GasBurned, uint64(0))
	require.Zero(t, ch.LastReceipt().GasFeeCharged)
}

func TestGovernanceSetMustGetPayoutAgentID(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{Debug: true, PrintStackTrace: true})
	ch := env.NewChain()

	user, userAddr := env.NewKeyPairWithFunds()
	userAgentID := isc.NewAddressAgentID(userAddr)

	_, err := ch.PostRequestSync(
		solo.NewCallParams(
			governance.FuncSetPayoutAgentID.Message(userAgentID),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	retDict, err := ch.CallView(governance.ViewGetPayoutAgentID.Message())
	require.NoError(t, err)
	retAgentID := lo.Must(governance.ViewGetPayoutAgentID.DecodeOutput(retDict))
	require.Equal(t, userAgentID, retAgentID)

	_, err = ch.PostRequestSync(
		solo.NewCallParams(governance.FuncSetPayoutAgentID.Message(userAgentID)).
			AddBaseTokens(1*isc.Million).
			WithMaxAffordableGasBudget(),
		user,
	)
	require.ErrorContains(t, err, "unauthorized access")
}

func TestGovernanceGasCoinTargetValue(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{Debug: true, PrintStackTrace: true})
	ch := env.NewChain()
	initRetDict, err := ch.CallView(governance.ViewGetGasCoinTargetValue.Message())
	require.NoError(t, err)
	retMinCommonAccountBalance := lo.Must(governance.ViewGetGasCoinTargetValue.DecodeOutput(initRetDict))
	require.EqualValues(t, isc.GasCoinTargetValue, retMinCommonAccountBalance)

	gasCoinTargetValue := coin.Value(123456)
	_, err = ch.PostRequestSync(
		solo.NewCallParams(
			governance.FuncSetGasCoinTargetValue.Message(gasCoinTargetValue),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	retDict, err := ch.CallView(governance.ViewGetGasCoinTargetValue.Message())
	require.NoError(t, err)
	retMinCommonAccountBalance = lo.Must(governance.ViewGetGasCoinTargetValue.DecodeOutput(retDict))
	require.NoError(t, err)
	require.EqualValues(t, gasCoinTargetValue, retMinCommonAccountBalance)
}

func TestGovernanceCallsNoBalance(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain(false)

	// the owner can call gov funcs without funds
	_, err := ch.PostRequestOffLedger(
		solo.NewCallParams(governance.FuncStartMaintenance.Message()),
		nil,
	)
	require.NoError(t, err)
	_, err = ch.PostRequestOffLedger(
		solo.NewCallParams(governance.FuncStopMaintenance.Message()),
		nil,
	)
	require.NoError(t, err)
}

func TestGovernanceGasPayout(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{
		Debug:           true,
		PrintStackTrace: true,
	})
	ch := env.NewChain(false)
	user1, user1Addr := env.NewKeyPairWithFunds()
	user1AgentID := isc.NewAddressAgentID(user1Addr)

	// transfer some tokens from a new account (user1)
	ownerBal1 := ch.L2Assets(ch.AdminAgentID())
	commonBal1 := ch.L2CommonAccountAssets()
	user1Bal1 := ch.L2Assets(user1AgentID)
	transferAmt := solo.BaseTokensForL2Gas
	_, _, vmRes, _, err := ch.PostRequestSyncTx(
		solo.NewCallParams(accounts.FuncDeposit.Message()).
			AddBaseTokens(transferAmt),
		user1,
	)
	require.NoError(t, err)
	gasFees := ch.LastReceipt().GasFeeCharged

	// assert gas payout works as expected, owner gets the fees minus common account top up
	addedToCommonAccount1 := min(
		isc.GasCoinTargetValue-commonBal1.BaseTokens(),
		vmRes.Receipt.GasFeeCharged,
	)
	ownerBal2 := ch.L2Assets(ch.AdminAgentID())
	user1Bal2 := ch.L2Assets(user1AgentID)
	require.Equal(t, ownerBal1.BaseTokens()+gasFees-addedToCommonAccount1, ownerBal2.BaseTokens())
	require.Equal(t, user1Bal1.BaseTokens()+transferAmt-gasFees, user1Bal2.BaseTokens())

	// change the payoutAddress, so that user1 now receives the fees charged by the chain
	_, err = ch.PostRequestOffLedger(
		solo.NewCallParams(
			governance.FuncSetPayoutAgentID.Message(user1AgentID),
		),
		nil,
	)
	require.NoError(t, err)

	// no balance changes (owner calls to gov contract don't pay fees)
	ownerBal3 := ch.L2Assets(ch.AdminAgentID())
	commonBal3 := ch.L2CommonAccountAssets()
	user1Bal3 := ch.L2Assets(user1AgentID)
	require.Equal(t, ownerBal2.BaseTokens(), ownerBal3.BaseTokens())
	require.Equal(t, user1Bal2.BaseTokens(), user1Bal3.BaseTokens())

	// assert new payoutAddr is correctly set
	retDict, err := ch.CallView(governance.ViewGetPayoutAgentID.Message())
	require.NoError(t, err)
	retAgentID := lo.Must(governance.ViewGetPayoutAgentID.DecodeOutput(retDict))
	require.NoError(t, err)
	require.Equal(t, user1AgentID, retAgentID)

	// send a new request (another deposit from user1)
	_, err = ch.PostRequestSync(
		solo.NewCallParams(accounts.FuncDeposit.Message()).
			AddBaseTokens(transferAmt),
		user1,
	)
	require.NoError(t, err)

	addedToCommonAccount3 := min(
		isc.GasCoinTargetValue-commonBal3.BaseTokens(),
		vmRes.Receipt.GasFeeCharged,
	)
	ownerBal4 := ch.L2Assets(ch.AdminAgentID())
	commonBal4 := ch.L2CommonAccountAssets()
	user1Bal4 := ch.L2Assets(user1AgentID)

	require.Equal(t, ownerBal3.BaseTokens(), ownerBal4.BaseTokens())
	require.Equal(t, commonBal3.BaseTokens()+addedToCommonAccount3, commonBal4.BaseTokens())
	require.Equal(t, user1Bal3.BaseTokens()+transferAmt-addedToCommonAccount3, user1Bal4.BaseTokens())
}
