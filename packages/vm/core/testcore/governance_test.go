package testcore

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/governance/governanceimpl"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
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
	getChainNodesResponse := governance.NewGetChainNodesResponseFromDict(res)
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
	getChainNodesResponse = governance.NewGetChainNodesResponseFromDict(res)
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
	getChainNodesResponse = governance.NewGetChainNodesResponseFromDict(res)
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
	getChainNodesResponse = governance.NewGetChainNodesResponseFromDict(res)
	require.Empty(t, getChainNodesResponse.AccessNodeCandidates)
	require.Empty(t, getChainNodesResponse.AccessNodes)
}

func TestDisallowMaintenanceDeadlock(t *testing.T) {
	// contracts of the same chain cannot turn on maintenance mode

	claimOwnershipFunc := coreutil.Func("claimOwnership")
	startMaintenceFunc := coreutil.Func("initMaintenance")
	stopMaintenceFunc := coreutil.Func("stopMaintenance")
	ownerContract := coreutil.NewContract("chain owner contract", "N/A")
	ownerContractProcessor := ownerContract.Processor(nil,
		claimOwnershipFunc.WithHandler(func(ctx isc.Sandbox) dict.Dict {
			return ctx.Call(governance.Contract.Hname(), governance.FuncClaimChainOwnership.Hname(), nil, nil)
		}),
		startMaintenceFunc.WithHandler(func(ctx isc.Sandbox) dict.Dict {
			return ctx.Call(governance.Contract.Hname(), governance.FuncStartMaintenance.Hname(), nil, nil)
		}),
		stopMaintenceFunc.WithHandler(func(ctx isc.Sandbox) dict.Dict {
			return ctx.Call(governance.Contract.Hname(), governance.FuncStopMaintenance.Hname(), nil, nil)
		}),
	)
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true}).
		WithNativeContract(ownerContractProcessor)
	ch := env.NewChain()

	ownerContractAgentID := isc.NewContractAgentID(ch.ChainID, ownerContract.Hname())
	userWallet, _ := env.NewKeyPairWithFunds()

	err := ch.DeployContract(nil, ownerContract.Name, ownerContract.ProgramHash)
	require.NoError(t, err)

	// from the initial owner - set maintenance
	_, err = ch.PostRequestSync(
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

	_, err = ch.PostRequestSync(
		solo.NewCallParams(ownerContract.Name, claimOwnershipFunc.Name).WithMaxAffordableGasBudget(),
		userWallet,
	)
	require.NoError(t, err)

	// the "owner contact" is able to stop maintenance mode
	_, err = ch.PostRequestSync(
		solo.NewCallParams(ownerContract.Name, stopMaintenceFunc.Name).WithMaxAffordableGasBudget(),
		userWallet,
	)
	require.NoError(t, err)

	// the "owner contract" is unable to start a new maintenance
	_, err = ch.PostRequestSync(
		solo.NewCallParams(ownerContract.Name, startMaintenceFunc.Name).WithMaxAffordableGasBudget(),
		userWallet,
	)
	require.Error(t, err)
}

func TestCustomL1Metadata(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	ch := env.NewChain()

	// set max valid size custom metadata
	_, err := ch.PostRequestSync(
		solo.NewCallParams(
			governance.Contract.Name,
			governance.FuncSetCustomMetadata.Name,
			governance.ParamCustomMetadata,
			[]byte(strings.Repeat("9", governanceimpl.MaxCustomMetadataLength)),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	// set custom metadata
	customMetadata := []byte("http://foobar.com")
	_, err = ch.PostRequestSync(
		solo.NewCallParams(
			governance.Contract.Name,
			governance.FuncSetCustomMetadata.Name,
			governance.ParamCustomMetadata,
			customMetadata,
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.NoError(t, err)

	// set invalid custom metadata
	_, err = ch.PostRequestSync(
		solo.NewCallParams(
			governance.Contract.Name,
			governance.FuncSetCustomMetadata.Name,
			governance.ParamCustomMetadata,
			[]byte(strings.Repeat("9", governanceimpl.MaxCustomMetadataLength+1)),
		).WithMaxAffordableGasBudget(),
		nil,
	)
	require.Error(t, err)

	// assert metadata is correct on view call
	res, err := ch.CallView(
		governance.Contract.Name,
		governance.ViewGetCustomMetadata.Name,
	)
	require.NoError(t, err)
	resMetadata := res.MustGet(governance.ParamCustomMetadata)
	require.Equal(t, customMetadata, resMetadata)

	// assert metadata is correct on L1 alias output
	ao, _ := ch.LatestAliasOutput()
	sm, err := vmcontext.StateMetadataFromBytes(ao.GetStateMetadata())
	require.NoError(t, err)
	require.Equal(t, customMetadata, sm.CustomMetadata)
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
	ao, _ = ch.LatestAliasOutput()
	sm, err = vmcontext.StateMetadataFromBytes(ao.GetStateMetadata())
	require.NoError(t, err)
	require.Equal(t, customMetadata, sm.CustomMetadata)
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
