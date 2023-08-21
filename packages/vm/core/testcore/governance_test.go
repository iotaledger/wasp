package testcore

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/governance/governanceimpl"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func TestGovernance1(t *testing.T) {
	corecontracts.PrintWellKnownHnames()

	t.Run("empty list of allowed rotation addresses", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		chain := env.NewChain()

		lst := chain.GetAllowedStateControllerAddresses()
		require.EqualValues(t, 0, len(lst))
	})
	t.Run("add/remove allowed rotation addresses", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		chain := env.NewChain()

		_, addr1 := env.NewKeyPair()
		err := chain.AddAllowedStateController(addr1, nil)
		require.NoError(t, err)
		res := chain.GetAllowedStateControllerAddresses()
		require.EqualValues(t, 1, len(res))

		_, addr2 := env.NewKeyPair()
		err = chain.AddAllowedStateController(addr2, nil)
		require.NoError(t, err)
		res = chain.GetAllowedStateControllerAddresses()
		require.EqualValues(t, 2, len(res))

		require.True(t, addr1.Equal(res[0]) || addr1.Equal(res[1]))
		require.True(t, addr2.Equal(res[0]) || addr2.Equal(res[1]))

		err = chain.RemoveAllowedStateController(addr1, nil)
		require.NoError(t, err)
		res = chain.GetAllowedStateControllerAddresses()
		require.EqualValues(t, 1, len(res))
		require.True(t, addr2.Equal(res[0]))

		err = chain.RemoveAllowedStateController(addr1, nil)
		require.NoError(t, err)
		res = chain.GetAllowedStateControllerAddresses()
		require.EqualValues(t, 1, len(res))
		require.True(t, addr2.Equal(res[0]))

		err = chain.RemoveAllowedStateController(addr2, nil)
		require.NoError(t, err)
		res = chain.GetAllowedStateControllerAddresses()
		require.EqualValues(t, 0, len(res))
	})
}

func TestRotate(t *testing.T) {
	corecontracts.PrintWellKnownHnames()

	t.Run("not allowed address", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		chain := env.NewChain()

		kp, addr := env.NewKeyPair()
		err := chain.RotateStateController(addr, kp, nil)
		require.Error(t, err)
		strings.Contains(err.Error(), "checkRotateCommitteeRequest: address is not allowed as next state address")
	})
	t.Run("unauthorized", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		chain := env.NewChain()

		kp, addr := env.NewKeyPairWithFunds()
		err := chain.RotateStateController(addr, kp, kp)
		require.Error(t, err)
		strings.Contains(err.Error(), "checkRotateStateControllerRequest: unauthorized access")
	})
	t.Run("rotate success", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		chain := env.NewChain()

		chain.WaitUntilMempoolIsEmpty()

		newKP, newAddr := env.NewKeyPair()
		err := chain.AddAllowedStateController(newAddr, nil)
		require.NoError(t, err)

		err = chain.RotateStateController(newAddr, newKP, nil)
		require.NoError(t, err)

		chain.WaitUntilMempoolIsEmpty()

		ca := chain.GetControlAddresses()
		require.True(t, ca.StateAddress.Equal(newAddr))

		req := solo.NewCallParams("dummy", "dummy").WithMaxAffordableGasBudget()
		_, err = chain.PostRequestSync(req, nil)
		testmisc.RequireErrorToBe(t, err, vm.ErrContractNotFound)
	})
}

func TestAccessNodes(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	node1KP, _ := env.NewKeyPairWithFunds()
	node1OwnerKP, node1OwnerAddr := env.NewKeyPairWithFunds()
	chainKP, _ := env.NewKeyPairWithFunds()
	chain, _ := env.NewChainExt(chainKP, 0, "chain1")
	var res dict.Dict
	var err error

	//
	// Initially the state is empty.
	res, err = chain.CallView(
		governance.Contract.Name,
		governance.ViewGetChainNodes.Name,
		governance.GetChainNodesRequest{}.AsDict(),
	)
	require.NoError(t, err)
	getChainNodesResponse := governance.GetChainNodesResponseFromDict(res)
	require.Empty(t, getChainNodesResponse.AccessNodeCandidates)
	require.Empty(t, getChainNodesResponse.AccessNodes)

	//
	// Add a single access node candidate.
	_, err = chain.PostRequestSync(
		solo.NewCallParams(
			governance.Contract.Name,
			governance.FuncAddCandidateNode.Name,
			(&governance.AccessNodeInfo{
				NodePubKey:   node1KP.GetPublicKey().AsBytes(),
				ForCommittee: false,
				AccessAPI:    "http://my-api/url",
			}).AddCertificate(node1KP, node1OwnerAddr).ToAddCandidateNodeParams(),
		).WithMaxAffordableGasBudget(),
		node1OwnerKP, // Sender should match data used to create the Cert field value.
	)
	require.NoError(t, err)

	res, err = chain.CallView(
		governance.Contract.Name,
		governance.ViewGetChainNodes.Name,
		governance.GetChainNodesRequest{}.AsDict(),
	)
	require.NoError(t, err)
	getChainNodesResponse = governance.GetChainNodesResponseFromDict(res)
	require.Equal(t, 1, len(getChainNodesResponse.AccessNodeCandidates)) // Candidate registered.
	require.Equal(t, "http://my-api/url", getChainNodesResponse.AccessNodeCandidates[0].AccessAPI)
	require.Empty(t, getChainNodesResponse.AccessNodes)

	//
	// Accept the node as an access node.
	_, err = chain.PostRequestSync(
		solo.NewCallParams(
			governance.Contract.Name,
			governance.FuncChangeAccessNodes.Name,
			governance.NewChangeAccessNodesRequest().Accept(node1KP.GetPublicKey()).AsDict(),
		).WithMaxAffordableGasBudget(),
		chainKP,
	)
	require.NoError(t, err)

	res, err = chain.CallView(
		governance.Contract.Name,
		governance.ViewGetChainNodes.Name,
		governance.GetChainNodesRequest{}.AsDict(),
	)
	require.NoError(t, err)
	getChainNodesResponse = governance.GetChainNodesResponseFromDict(res)
	require.Equal(t, 1, len(getChainNodesResponse.AccessNodeCandidates)) // Candidate registered.
	require.Equal(t, "http://my-api/url", getChainNodesResponse.AccessNodeCandidates[0].AccessAPI)
	require.Equal(t, 1, len(getChainNodesResponse.AccessNodes))

	//
	// Revoke the access node (by the node owner).
	_, err = chain.PostRequestSync(
		solo.NewCallParams(
			governance.Contract.Name,
			governance.FuncRevokeAccessNode.Name,
			(&governance.AccessNodeInfo{
				NodePubKey: node1KP.GetPublicKey().AsBytes(),
			}).AddCertificate(node1KP, node1OwnerAddr).ToAddCandidateNodeParams(),
		).WithMaxAffordableGasBudget(),
		node1OwnerKP, // Sender should match data used to create the Cert field value.
	)
	require.NoError(t, err)

	res, err = chain.CallView(
		governance.Contract.Name,
		governance.ViewGetChainNodes.Name,
		governance.GetChainNodesRequest{}.AsDict(),
	)
	require.NoError(t, err)
	getChainNodesResponse = governance.GetChainNodesResponseFromDict(res)
	require.Empty(t, getChainNodesResponse.AccessNodeCandidates)
	require.Empty(t, getChainNodesResponse.AccessNodes)
}

func TestMaintenanceMode(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true}).
		WithNativeContract(inccounter.Processor)
	ch := env.NewChain()

	ownerWallet, ownerAddr := env.NewKeyPairWithFunds()
	ownerAgentID := isc.NewAgentID(ownerAddr)
	ch.DepositBaseTokensToL2(10*isc.Million, ownerWallet)

	userWallet, _ := env.NewKeyPairWithFunds()
	ch.DepositBaseTokensToL2(10*isc.Million, userWallet)

	// set owner of the chain
	{
		_, err2 := ch.PostRequestSync(
			solo.NewCallParams(governance.Contract.Name, governance.FuncDelegateChainOwnership.Name,
				governance.ParamChainOwner, codec.Encode(ownerAgentID),
			).WithMaxAffordableGasBudget(),
			nil,
		)
		require.NoError(t, err2)

		_, err2 = ch.PostRequestSync(
			solo.NewCallParams(governance.Contract.Name, governance.FuncClaimChainOwnership.Name).WithMaxAffordableGasBudget(),
			ownerWallet,
		)
		require.NoError(t, err2)
	}

	// call the gov "maintenance status view", check it is OFF
	{
		// TODO: Add maintenance status to wrapped core contracts
		ret, err2 := ch.CallView(governance.Contract.Name, governance.ViewGetMaintenanceStatus.Name)
		require.NoError(t, err2)
		maintenanceStatus := codec.MustDecodeBool(ret.Get(governance.VarMaintenanceStatus))
		require.False(t, maintenanceStatus)
	}

	// test non-chain owner cannot call init maintenance
	{
		_, err2 := ch.PostRequestSync(
			solo.NewCallParams(governance.Contract.Name, governance.FuncStartMaintenance.Name).WithMaxAffordableGasBudget(),
			userWallet,
		)
		require.ErrorContains(t, err2, "unauthorized")
	}

	// owner can start maintenance mode
	{
		_, err2 := ch.PostRequestSync(
			solo.NewCallParams(governance.Contract.Name, governance.FuncStartMaintenance.Name).WithMaxAffordableGasBudget(),
			ownerWallet,
		)
		require.NoError(t, err2)
	}

	// call the gov "maintenance status view", check it is ON
	{
		ret, err2 := ch.CallView(governance.Contract.Name, governance.ViewGetMaintenanceStatus.Name)
		require.NoError(t, err2)
		maintenanceStatus := codec.MustDecodeBool(ret.Get(governance.VarMaintenanceStatus))
		require.True(t, maintenanceStatus)
	}

	// calls to non-maintenance endpoints are not processed
	ch.WaitForRequestsMark()

	var reqs []isc.OffLedgerRequest
	{
		for _, wallet := range []*cryptolib.KeyPair{userWallet, ownerWallet} {
			req := solo.NewCallParams(inccounter.Contract.Name, inccounter.FuncIncCounter.Name).
				WithMaxAffordableGasBudget().
				NewRequestOffLedger(ch, wallet)
			env.AddRequestsToMempool(ch, []isc.Request{req})
			reqs = append(reqs, req)
		}
	}

	// give some time for the requests to be picked up from the mempool
	require.False(t, ch.WaitForRequestsThrough(2, 200*time.Millisecond))

	// requests are skipped
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
			governance.Contract.Name,
			governance.FuncSetFeePolicy.Name,
			dict.Dict{
				governance.ParamFeePolicyBytes: fp.Bytes(),
			},
		), ownerWallet)
		require.NoError(t, err2)
	}

	// calls to governance from non-owners should be processed, but fail
	{
		_, err2 := ch.PostRequestSync(solo.NewCallParams(
			governance.Contract.Name,
			governance.FuncSetFeePolicy.Name,
			dict.Dict{
				governance.ParamFeePolicyBytes: fp.Bytes(),
			},
		), userWallet)
		require.ErrorContains(t, err2, "unauthorized")
	}

	// test non-chain owner cannot call stop maintenance
	{
		_, err2 := ch.PostRequestSync(
			solo.NewCallParams(governance.Contract.Name, governance.FuncStopMaintenance.Name).WithMaxAffordableGasBudget(),
			userWallet,
		)
		require.ErrorContains(t, err2, "unauthorized")
	}

	// requests are still skipped
	for _, req := range reqs {
		require.False(t, ch.IsRequestProcessed(req.ID()))
	}

	ch.WaitForRequestsMark()

	// owner can stop maintenance mode
	{
		_, err2 := ch.PostRequestSync(
			solo.NewCallParams(governance.Contract.Name, governance.FuncStopMaintenance.Name).WithMaxAffordableGasBudget(),
			ownerWallet,
		)
		require.NoError(t, err2)
	}

	// normal requests are now processed successfully (pending requests issued during maintenance should be processed now)
	require.True(t, ch.WaitForRequestsThrough(3, 1*time.Second))
	for _, req := range reqs {
		require.True(t, ch.IsRequestProcessed(req.ID()))
	}
}

var (
	claimOwnershipFunc   = coreutil.Func("claimOwnership")
	startMaintenanceFunc = coreutil.Func("initMaintenance")
)

func createOwnerContract(t *testing.T) (*solo.Chain, *coreutil.ContractInfo) {
	ownerContract := coreutil.NewContract("chain owner contract")
	ownerContractProcessor := ownerContract.Processor(nil,
		claimOwnershipFunc.WithHandler(func(ctx isc.Sandbox) dict.Dict {
			return ctx.Call(governance.Contract.Hname(), governance.FuncClaimChainOwnership.Hname(), nil, nil)
		}),
		startMaintenanceFunc.WithHandler(func(ctx isc.Sandbox) dict.Dict {
			return ctx.Call(governance.Contract.Hname(), governance.FuncStartMaintenance.Hname(), nil, nil)
		}),
	)
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true}).
		WithNativeContract(ownerContractProcessor)
	ch := env.NewChain()

	err := ch.DeployContract(nil, ownerContract.Name, ownerContract.ProgramHash)
	require.NoError(t, err)

	return ch, ownerContract
}

func TestDisallowMaintenanceDeadlock1(t *testing.T) {
	ch, ownerContract := createOwnerContract(t)

	ownerContractAgentID := isc.NewContractAgentID(ch.ChainID, ownerContract.Hname())
	userWallet, _ := ch.Env.NewKeyPairWithFunds()

	// from the initial owner - set maintenance
	_, err := ch.PostRequestSync(
		solo.NewCallParams(governance.Contract.Name, governance.FuncStartMaintenance.Name).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	// set the "owner contract" as the new chain owner
	_, err = ch.PostRequestSync(
		solo.NewCallParams(governance.Contract.Name, governance.FuncDelegateChainOwnership.Name,
			governance.ParamChainOwner, codec.Encode(ownerContractAgentID)).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	// the "owner contract" cannot claim ownership
	_, err = ch.PostRequestSync(
		solo.NewCallParams(ownerContract.Name, claimOwnershipFunc.Name).WithMaxAffordableGasBudget(),
		userWallet,
	)
	require.ErrorContains(t, err, "skipped")
}

func TestDisallowMaintenanceDeadlock2(t *testing.T) {
	ch, ownerContract := createOwnerContract(t)

	ownerContractAgentID := isc.NewContractAgentID(ch.ChainID, ownerContract.Hname())
	userWallet, _ := ch.Env.NewKeyPairWithFunds()

	// set the "owner contract" as the new chain owner
	_, err := ch.PostRequestSync(
		solo.NewCallParams(governance.Contract.Name, governance.FuncDelegateChainOwnership.Name,
			governance.ParamChainOwner, codec.Encode(ownerContractAgentID)).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	_, err = ch.PostRequestSync(
		solo.NewCallParams(ownerContract.Name, claimOwnershipFunc.Name).WithMaxAffordableGasBudget(),
		userWallet,
	)
	require.NoError(t, err)

	// the "owner contract" is unable to start maintenance
	_, err = ch.PostRequestSync(
		solo.NewCallParams(ownerContract.Name, startMaintenanceFunc.Name).WithMaxAffordableGasBudget(),
		userWallet,
	)
	require.ErrorContains(t, err, "unauthorized")
}

func TestMetadata(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	ch := env.NewChain()

	// deposit some extra tokens to the common account to accommodate for the SD change
	ch.SendFromL1ToL2AccountBaseTokens(10*isc.Million, 9*isc.Million, accounts.CommonAccount(), nil)

	/*
		Values with the length == 0 will reset the state value
		Values with the length > 0 will set the state value
		Nil values will be ignored and not change the state value
	*/

	testValue := "TESTYTEST"

	testMetadata := isc.PublicChainMetadata{
		EVMJsonRPCURL:   testValue,
		EVMWebSocketURL: testValue,
		Name:            testValue,
		Description:     testValue,
		Website:         testValue,
	}

	// Set proper metadata value
	_, err := ch.PostRequestSync(
		solo.NewCallParams(
			governance.Contract.Name,
			governance.FuncSetMetadata.Name,
			governance.ParamMetadata,
			testMetadata.Bytes(),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	res, err := ch.CallView(
		governance.Contract.Name,
		governance.ViewGetMetadata.Name,
	)
	require.NoError(t, err)
	resMetadata := res.Get(governance.ParamMetadata)

	// Chain name should be equal to the configured one.
	require.Equal(t, testMetadata.Bytes(), resMetadata)

	// Call SetMetadata without args. The metadata should be the same as it was previously configured and not be emptied.
	_, err = ch.PostRequestSync(
		solo.NewCallParams(
			governance.Contract.Name,
			governance.FuncSetMetadata.Name,
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	res, err = ch.CallView(
		governance.Contract.Name,
		governance.ViewGetMetadata.Name,
	)
	require.NoError(t, err)
	resMetadata = res.Get(governance.ParamMetadata)

	// Chain name should be equal to the configured one.
	require.Equal(t, testMetadata.Bytes(), resMetadata)

	// Call SetMetadata with an empty arg. The metadata call should fail.
	_, err = ch.PostRequestSync(
		solo.NewCallParams(
			governance.Contract.Name,
			governance.FuncSetMetadata.Name,
			governance.ParamMetadata,
			[]byte{},
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.Error(t, err)

	// Test invalid custom metadata
	_, err = ch.PostRequestSync(
		solo.NewCallParams(
			governance.Contract.Name,
			governance.FuncSetMetadata.Name,
			governance.ParamMetadata,
			string(make([]byte, governanceimpl.MaxCustomMetadataLength+1)),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.Error(t, err)

	// set invalid custom metadata
	hugePublicChainMetadata := isc.PublicChainMetadata{
		Website: string(make([]byte, governanceimpl.MaxCustomMetadataLength+1)),
	}
	_, err = ch.PostRequestSync(
		solo.NewCallParams(
			governance.Contract.Name,
			governance.FuncSetMetadata.Name,
			governance.ParamPublicURL,
			hugePublicChainMetadata.Bytes(),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.Error(t, err)
}

func TestL1Metadata(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	ch := env.NewChain()

	// deposit some extra tokens to the common account to accommodate for the SD change
	ch.SendFromL1ToL2AccountBaseTokens(10*isc.Million, 9*isc.Million, accounts.CommonAccount(), nil)

	// set max valid size custom metadata
	publicURLMetadata := "https://iota.org"

	_, err := ch.PostRequestSync(
		solo.NewCallParams(
			governance.Contract.Name,
			governance.FuncSetMetadata.Name,
			governance.ParamPublicURL,
			publicURLMetadata,
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	// assert metadata is correct on view call
	res, err := ch.CallView(
		governance.Contract.Name,
		governance.ViewGetMetadata.Name,
	)
	require.NoError(t, err)
	resMetadata := res.Get(governance.ParamPublicURL)
	require.Equal(t, publicURLMetadata, string(resMetadata))

	// assert metadata is correct on L1 alias output
	ao, err := ch.LatestAliasOutput(chain.ActiveOrCommittedState)
	require.NoError(t, err)
	sm, err := transaction.StateMetadataFromBytes(ao.GetStateMetadata())
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
			governance.Contract.Name,
			governance.FuncSetFeePolicy.Name,
			governance.ParamFeePolicyBytes,
			newFeePolicy.Bytes(),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	// assert gas policy changed on L1 metadata
	ao, err = ch.LatestAliasOutput(chain.ActiveOrCommittedState)
	require.NoError(t, err)
	sm, err = transaction.StateMetadataFromBytes(ao.GetStateMetadata())
	require.NoError(t, err)
	require.Equal(t, publicURLMetadata, sm.PublicURL)
	require.True(t, reflect.DeepEqual(sm.GasFeePolicy, newFeePolicy))
}

func TestGovernanceGasFee(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true, Debug: true, PrintStackTrace: true})
	ch := env.NewChain()
	fp := ch.GetGasFeePolicy()
	fp.GasPerToken.A *= 1000000
	ch.SetGasFeePolicy(nil, fp)
	fp.GasPerToken.A /= 1000000
	ch.SetGasFeePolicy(nil, fp) // should not fail with "gas budget exceeded"
}

func TestGovernanceZeroGasFee(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true, Debug: true, PrintStackTrace: true})
	ch := env.NewChain()

	user1, userAddr1 := env.NewKeyPairWithFunds()
	userAgentID1 := isc.NewAgentID(userAddr1)
	_, userAddr2 := env.NewKeyPairWithFunds()
	userAgentID2 := isc.NewAgentID(userAddr2)

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
			governance.Contract.Name,
			governance.FuncSetFeePolicy.Name,
			governance.VarGasFeePolicyBytes,
			fp.Bytes(),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	estimateGas, estimateGasFee, err := ch.EstimateGasOnLedger(solo.NewCallParams(
		accounts.Contract.Name,
		accounts.FuncDeposit.Name,
	), user1, true)
	require.NoError(t, err)
	require.Zero(t, estimateGasFee)

	userL2Bal1 := ch.L2BaseTokens(userAgentID1)
	userL1Bal1 := ch.Env.L1BaseTokens(userAddr1)

	gasGreaterThanEstimatedGas := estimateGas + 100
	_, err = ch.PostRequestSync(
		solo.NewCallParams(
			accounts.Contract.Name,
			accounts.FuncTransferAllowanceTo.Name,
			dict.Dict{
				accounts.ParamAgentID: codec.EncodeAgentID(userAgentID2),
			},
		).
			AddBaseTokens(gasGreaterThanEstimatedGas).
			AddAllowanceBaseTokens(gasGreaterThanEstimatedGas).
			WithGasBudget(gasGreaterThanEstimatedGas),
		user1,
	)
	require.NoError(t, err)

	userL2Bal2 := ch.L2BaseTokens(userAgentID1)
	userL1Bal2 := ch.Env.L1BaseTokens(userAddr1)
	require.Equal(t, userL2Bal1, userL2Bal2)
	require.Equal(t, userL1Bal1-gasGreaterThanEstimatedGas, userL1Bal2)
	require.Greater(t, ch.LastReceipt().GasBurned, uint64(0))
	require.Zero(t, ch.LastReceipt().GasFeeCharged)

	gasLessThanEstimatedGas := estimateGas - 100
	_, err = ch.PostRequestSync(
		solo.NewCallParams(
			accounts.Contract.Name,
			accounts.FuncTransferAllowanceTo.Name,
			dict.Dict{
				accounts.ParamAgentID: codec.EncodeAgentID(userAgentID2),
			},
		).
			AddBaseTokens(gasLessThanEstimatedGas).
			WithGasBudget(gasLessThanEstimatedGas),
		user1,
	)
	require.NoError(t, err)

	userL2Bal3 := ch.L2BaseTokens(userAgentID1)
	require.Equal(t, userL2Bal2+gasLessThanEstimatedGas, userL2Bal3)
	require.Greater(t, ch.LastReceipt().GasBurned, uint64(0))
	require.Zero(t, ch.LastReceipt().GasFeeCharged)
}

func TestGovernanceSetMustGetPayoutAgentID(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true, Debug: true, PrintStackTrace: true})
	ch := env.NewChain()

	user, userAddr := env.NewKeyPairWithFunds()
	userAgentID := isc.NewAgentID(userAddr)

	_, err := ch.PostRequestSync(
		solo.NewCallParams(
			governance.Contract.Name,
			governance.FuncSetPayoutAgentID.Name,
			governance.ParamSetPayoutAgentID,
			userAgentID.Bytes(),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	retDict, err := ch.CallView(
		governance.Contract.Name,
		governance.ViewGetPayoutAgentID.Name,
	)
	require.NoError(t, err)
	retAgentID, err := codec.DecodeAgentID(retDict.Get(governance.ParamSetPayoutAgentID))
	require.NoError(t, err)
	require.Equal(t, userAgentID, retAgentID)

	_, err = ch.PostRequestSync(
		solo.NewCallParams(
			governance.Contract.Name,
			governance.FuncSetPayoutAgentID.Name,
			governance.ParamSetPayoutAgentID,
			userAgentID.Bytes(),
		).WithMaxAffordableGasBudget(),
		user,
	)
	require.ErrorContains(t, err, "unauthorized access")
}

func TestGovernanceSetGetMinCommonAccountBalance(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true, Debug: true, PrintStackTrace: true})
	ch := env.NewChain()

	initRetDict, err := ch.CallView(
		governance.Contract.Name,
		governance.ViewGetMinCommonAccountBalance.Name,
	)
	require.NoError(t, err)
	retByte := initRetDict.Get(governance.ParamSetMinCommonAccountBalance)
	retMinCommonAccountBalance, err := codec.DecodeUint64(retByte)
	require.NoError(t, err)
	require.Equal(t, governance.DefaultMinBaseTokensOnCommonAccount, retMinCommonAccountBalance)

	minCommonAccountBalance := uint64(123456)
	_, err = ch.PostRequestSync(
		solo.NewCallParams(
			governance.Contract.Name,
			governance.FuncSetMinCommonAccountBalance.Name,
			governance.ParamSetMinCommonAccountBalance,
			codec.EncodeUint64(minCommonAccountBalance),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	retDict, err := ch.CallView(
		governance.Contract.Name,
		governance.ViewGetMinCommonAccountBalance.Name,
	)
	require.NoError(t, err)
	retByte = retDict.Get(governance.ParamSetMinCommonAccountBalance)
	retMinCommonAccountBalance, err = codec.DecodeUint64(retByte)
	require.NoError(t, err)
	require.Equal(t, minCommonAccountBalance, retMinCommonAccountBalance)
}

func TestGovCallsNoBalance(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{})
	ch := env.NewChain(false)

	// the owner can call gov funcs without funds
	_, err := ch.PostRequestOffLedger(
		solo.NewCallParams(governance.Contract.Name, governance.FuncStartMaintenance.Name),
		nil,
	)
	require.NoError(t, err)
	_, err = ch.PostRequestOffLedger(
		solo.NewCallParams(governance.Contract.Name, governance.FuncStopMaintenance.Name),
		nil,
	)
	require.NoError(t, err)
}

func TestGasPayout(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{
		AutoAdjustStorageDeposit: true,
		Debug:                    true,
		PrintStackTrace:          true,
	})
	ch := env.NewChain(false)
	user1, user1Addr := env.NewKeyPairWithFunds()
	user1AgentID := isc.NewAgentID(user1Addr)

	// transfer some tokens from a new account (user1)
	ownerBal1 := ch.L2Assets(ch.OriginatorAgentID)
	user1Bal1 := ch.L2Assets(user1AgentID)
	transferAmt := uint64(2000)
	_, err := ch.PostRequestSync(
		solo.NewCallParams(
			accounts.Contract.Name,
			accounts.FuncDeposit.Name,
		).AddBaseTokens(transferAmt),
		user1,
	)
	require.NoError(t, err)
	gasFees := ch.LastReceipt().GasFeeCharged

	// assert gas payout works as expected, owner gets the fees
	ownerBal2 := ch.L2Assets(ch.OriginatorAgentID)
	commonBal2 := ch.L2CommonAccountAssets()
	user1Bal2 := ch.L2Assets(user1AgentID)
	require.Equal(t, ownerBal1.BaseTokens+gasFees, ownerBal2.BaseTokens)
	require.Equal(t, user1Bal1.BaseTokens+transferAmt-gasFees, user1Bal2.BaseTokens)

	// change the payoutAddress, so that user1 now receives the fees charged by the chain
	_, err = ch.PostRequestOffLedger(
		solo.NewCallParams(
			governance.Contract.Name,
			governance.FuncSetPayoutAgentID.Name,
			governance.ParamSetPayoutAgentID,
			user1AgentID.Bytes(),
		),
		nil,
	)
	require.NoError(t, err)

	// no balance changes (owner calls to gov contract don't pay fees)
	ownerBal3 := ch.L2Assets(ch.OriginatorAgentID)
	commonBal3 := ch.L2CommonAccountAssets()
	user1Bal3 := ch.L2Assets(user1AgentID)
	require.Equal(t, ownerBal2.BaseTokens, ownerBal3.BaseTokens)
	require.Equal(t, commonBal2.BaseTokens, commonBal3.BaseTokens)
	require.Equal(t, user1Bal2.BaseTokens, user1Bal3.BaseTokens)

	// assert new payoutAddr is correctly set
	retDict, err := ch.CallView(
		governance.Contract.Name,
		governance.ViewGetPayoutAgentID.Name,
	)
	require.NoError(t, err)
	retAgentID, err := codec.DecodeAgentID(retDict.Get(governance.ParamSetPayoutAgentID))
	require.NoError(t, err)
	require.Equal(t, user1AgentID, retAgentID)

	// send a new request (another deposit from user1)
	_, err = ch.PostRequestSync(
		solo.NewCallParams(
			accounts.Contract.Name,
			accounts.FuncDeposit.Name,
		).AddBaseTokens(transferAmt),
		user1,
	)
	require.NoError(t, err)
	gasFees = ch.LastReceipt().GasFeeCharged
	ownerBal4 := ch.L2Assets(ch.OriginatorAgentID)
	commonBal4 := ch.L2CommonAccountAssets()
	user1Bal4 := ch.L2Assets(user1AgentID)
	require.Equal(t, ownerBal3.BaseTokens, ownerBal4.BaseTokens)
	// because common account has less balance than minimum, fees go to the common account
	require.Less(t, commonBal3.BaseTokens, governance.DefaultMinBaseTokensOnCommonAccount)
	require.Equal(t, commonBal3.BaseTokens+gasFees, commonBal4.BaseTokens)
	require.Equal(t, user1Bal3.BaseTokens+transferAmt-gasFees, user1Bal4.BaseTokens)

	// top-up the common account, so its the minBalance - 10 tokens, assert what happens with the fees
	err = ch.TransferAllowanceTo(
		isc.NewAssetsBaseTokens(governance.DefaultMinBaseTokensOnCommonAccount-commonBal4.BaseTokens-10),
		accounts.CommonAccount(),
		nil,
	)
	require.NoError(t, err)
	commonBal5 := ch.L2CommonAccountAssets()
	user1Bal5 := ch.L2Assets(user1AgentID)
	gasFees = ch.LastReceipt().GasFeeCharged
	require.Equal(t, governance.DefaultMinBaseTokensOnCommonAccount, commonBal5.BaseTokens)
	require.Equal(t, user1Bal4.BaseTokens+gasFees-10, user1Bal5.BaseTokens)
}
