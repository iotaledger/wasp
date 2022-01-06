package testcore

import (
	"strings"
	"testing"

	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/stretchr/testify/require"
)

func TestGovernance1(t *testing.T) {
	core.PrintWellKnownHnames()

	t.Run("empty list of allowed rotation addresses", func(t *testing.T) {
		env := solo.New(t, false, false)
		chain := env.NewChain(nil, "chain1")
		defer chain.Log.Sync()

		lst := chain.GetAllowedStateControllerAddresses()
		require.EqualValues(t, 0, len(lst))
	})
	t.Run("add/remove allowed rotation addresses", func(t *testing.T) {
		env := solo.New(t, false, false)
		chain := env.NewChain(nil, "chain1")
		defer chain.Log.Sync()

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

		require.True(t, addr1.Equals(res[0]) || addr1.Equals(res[1]))
		require.True(t, addr2.Equals(res[0]) || addr2.Equals(res[1]))

		err = chain.RemoveAllowedStateController(addr1, nil)
		require.NoError(t, err)
		res = chain.GetAllowedStateControllerAddresses()
		require.EqualValues(t, 1, len(res))
		require.True(t, addr2.Equals(res[0]))

		err = chain.RemoveAllowedStateController(addr1, nil)
		require.NoError(t, err)
		res = chain.GetAllowedStateControllerAddresses()
		require.EqualValues(t, 1, len(res))
		require.True(t, addr2.Equals(res[0]))

		err = chain.RemoveAllowedStateController(addr2, nil)
		require.NoError(t, err)
		res = chain.GetAllowedStateControllerAddresses()
		require.EqualValues(t, 0, len(res))
	})
}

func TestRotate(t *testing.T) {
	core.PrintWellKnownHnames()

	t.Run("not allowed address", func(t *testing.T) {
		env := solo.New(t, false, false)
		chain := env.NewChain(nil, "chain1")
		defer chain.Log.Sync()

		kp, addr := env.NewKeyPair()
		err := chain.RotateStateController(addr, kp, nil)
		require.Error(t, err)
		strings.Contains(err.Error(), "checkRotateCommitteeRequest: address is not allowed as next state address")
	})
	t.Run("unauthorized", func(t *testing.T) {
		env := solo.New(t, false, false)
		chain := env.NewChain(nil, "chain1")
		defer chain.Log.Sync()

		kp, addr := env.NewKeyPairWithFunds()
		err := chain.RotateStateController(addr, kp, kp)
		require.Error(t, err)
		strings.Contains(err.Error(), "checkRotateStateControllerRequest: unauthorized access")
	})
	t.Run("rotate success", func(t *testing.T) {
		env := solo.New(t, false, false)
		chain := env.NewChain(nil, "chain1")
		defer chain.Log.Sync()

		newKP, newAddr := env.NewKeyPair()
		err := chain.AddAllowedStateController(newAddr, nil)
		require.NoError(t, err)

		err = chain.RotateStateController(newAddr, newKP, nil)
		require.NoError(t, err)

		require.True(t, chain.WaitForRequestsThrough(3))

		ca := chain.GetControlAddresses()
		require.EqualValues(t, ca.StateAddress.Base58(), newAddr.Base58())

		req := solo.NewCallParams("dummy", "dummy").AddAssetsIotas(1)
		_, err = chain.PostRequestSync(req, nil)
		require.NoError(t, err)

		require.True(t, chain.WaitForRequestsThrough(4))
	})
}

func TestAccessNodes(t *testing.T) {
	env := solo.New(t, true, true)
	node1KP, _ := env.NewKeyPairWithFunds()
	node1OwnerKP, node1OwnerAddr := env.NewKeyPairWithFunds()
	chainKP, _ := env.NewKeyPairWithFunds()
	chain := env.NewChain(chainKP, "chain1")
	defer chain.Log.Sync()
	var res dict.Dict
	var err error

	//
	// Initially the state is empty.
	res, err = chain.CallView(
		governance.Contract.Name,
		governance.FuncGetChainNodes.Name,
		governance.GetChainNodesRequest{}.AsDict(),
	)
	require.NoError(t, err)
	getChainNodesResponse := governance.NewGetChainNodesResponseFromDict(res)
	require.Empty(t, getChainNodesResponse.AccessNodeCandidates)
	require.Empty(t, getChainNodesResponse.AccessNodes)

	//
	// Add a single access node candidate.
	_, err = chain.PostRequestSync(
		solo.NewCallParamsFromDic(
			governance.Contract.Name,
			governance.FuncAddCandidateNode.Name,
			(&governance.AccessNodeInfo{
				NodePubKey:   node1KP.PublicKey.Bytes(),
				ForCommittee: false,
				AccessAPI:    "http://my-api/url",
			}).AddCertificate(node1KP.PrivateKey, node1OwnerAddr).ToAddCandidateNodeParams(),
		).AddAssetsIotas(1),
		node1OwnerKP, // Sender should match data used to create the Cert field value.
	)
	require.NoError(t, err)

	res, err = chain.CallView(
		governance.Contract.Name,
		governance.FuncGetChainNodes.Name,
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
		solo.NewCallParamsFromDic(
			governance.Contract.Name,
			governance.FuncChangeAccessNodes.Name,
			governance.NewChangeAccessNodesRequest().Accept(node1KP.PublicKey).AsDict(),
		).AddAssetsIotas(1),
		chainKP,
	)
	require.NoError(t, err)

	res, err = chain.CallView(
		governance.Contract.Name,
		governance.FuncGetChainNodes.Name,
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
		solo.NewCallParamsFromDic(
			governance.Contract.Name,
			governance.FuncRevokeAccessNode.Name,
			(&governance.AccessNodeInfo{
				NodePubKey: node1KP.PublicKey.Bytes(),
			}).AddCertificate(node1KP.PrivateKey, node1OwnerAddr).ToAddCandidateNodeParams(),
		).AddAssetsIotas(1),
		node1OwnerKP, // Sender should match data used to create the Cert field value.
	)
	require.NoError(t, err)

	res, err = chain.CallView(
		governance.Contract.Name,
		governance.FuncGetChainNodes.Name,
		governance.GetChainNodesRequest{}.AsDict(),
	)
	require.NoError(t, err)
	getChainNodesResponse = governance.NewGetChainNodesResponseFromDict(res)
	require.Empty(t, getChainNodesResponse.AccessNodeCandidates)
	require.Empty(t, getChainNodesResponse.AccessNodes)
}
