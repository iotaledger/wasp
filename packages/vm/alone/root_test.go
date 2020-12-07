package alone

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRootBasic(t *testing.T) {
	e := New(t, false, false)
	defer e.WaitEmptyBacklog()

	e.CheckBase()
	e.Infof("\n%s\n", e.String())
}

func TestRootRepeatInit(t *testing.T) {
	e := New(t, false, false)
	defer e.WaitEmptyBacklog()

	req := NewCall(root.Interface.Name, "init")
	_, err := e.PostRequest(req, nil)
	require.Error(t, err)
}

func TestGetInfo(t *testing.T) {
	e := New(t, false, false)
	defer e.WaitEmptyBacklog()

	chainID, chainOwnerID, contracts := e.GetInfo()

	require.EqualValues(t, e.ChainID, chainID)
	require.EqualValues(t, e.OriginatorAgentID, chainOwnerID)
	require.EqualValues(t, 3, len(contracts))

	_, ok := contracts[root.Interface.Hname()]
	require.True(t, ok)
	recBlob, ok := contracts[blob.Interface.Hname()]
	require.True(t, ok)
	_, ok = contracts[accountsc.Interface.Hname()]
	require.True(t, ok)

	rec, err := e.FindContract(blob.Interface.Name)
	require.NoError(t, err)
	require.EqualValues(t, root.EncodeContractRecord(recBlob), root.EncodeContractRecord(rec))
}

func TestDeployExample(t *testing.T) {
	name := "testInc"
	e := New(t, false, false)
	defer e.WaitEmptyBacklog()

	err := e.DeployContract(nil, name, inccounter.ProgramHash)
	require.NoError(t, err)

	chainID, chainOwnerID, contracts := e.GetInfo()

	require.EqualValues(t, e.ChainID, chainID)
	require.EqualValues(t, e.OriginatorAgentID, chainOwnerID)
	require.EqualValues(t, 4, len(contracts))

	_, ok := contracts[root.Interface.Hname()]
	require.True(t, ok)
	_, ok = contracts[blob.Interface.Hname()]
	require.True(t, ok)
	_, ok = contracts[accountsc.Interface.Hname()]
	require.True(t, ok)

	rec, ok := contracts[coretypes.Hn(name)]
	require.True(t, ok)

	require.EqualValues(t, name, rec.Name)
	require.EqualValues(t, "N/A", rec.Description)
	require.EqualValues(t, 0, rec.NodeFee)
	require.EqualValues(t, e.OriginatorAgentID, rec.Creator)
	require.EqualValues(t, inccounter.ProgramHash, rec.ProgramHash)

	recFind, err := e.FindContract(name)
	require.NoError(t, err)
	require.EqualValues(t, root.EncodeContractRecord(recFind), root.EncodeContractRecord(rec))
}

func TestDeployDouble(t *testing.T) {
	name := "testInc"
	e := New(t, false, false)
	defer e.WaitEmptyBacklog()

	err := e.DeployContract(nil, name, inccounter.ProgramHash)
	require.NoError(t, err)

	err = e.DeployContract(nil, name, inccounter.ProgramHash)
	require.Error(t, err)

	chainID, chainOwnerID, contracts := e.GetInfo()

	require.EqualValues(t, e.ChainID, chainID)
	require.EqualValues(t, e.OriginatorAgentID, chainOwnerID)
	require.EqualValues(t, 4, len(contracts))

	_, ok := contracts[root.Interface.Hname()]
	require.True(t, ok)
	_, ok = contracts[blob.Interface.Hname()]
	require.True(t, ok)
	_, ok = contracts[accountsc.Interface.Hname()]
	require.True(t, ok)

	rec, ok := contracts[coretypes.Hn(name)]
	require.True(t, ok)

	require.EqualValues(t, name, rec.Name)
	require.EqualValues(t, "N/A", rec.Description)
	require.EqualValues(t, 0, rec.NodeFee)
	require.EqualValues(t, e.OriginatorAgentID, rec.Creator)
	require.EqualValues(t, inccounter.ProgramHash, rec.ProgramHash)
}

func TestChangeOwnerAuthorized(t *testing.T) {
	e := New(t, false, true)
	defer e.WaitEmptyBacklog()

	newOwner := e.NewSigScheme()
	newOwnerAgentID := coretypes.NewAgentIDFromAddress(newOwner.Address())
	req := NewCall(root.Interface.Name, root.FuncAllowChangeChainOwner, root.ParamChainOwner, newOwnerAgentID)
	_, err := e.PostRequest(req, nil)
	require.NoError(t, err)

	_, ownerBack, _ := e.GetInfo()
	require.EqualValues(t, e.OriginatorAgentID, ownerBack)

	req = NewCall(root.Interface.Name, root.FuncChangeChainOwner)
	_, err = e.PostRequest(req, newOwner)
	require.NoError(t, err)

	_, ownerBack, _ = e.GetInfo()
	require.EqualValues(t, newOwnerAgentID, ownerBack)
}

func TestChangeOwnerUnauthorized(t *testing.T) {
	e := New(t, false, false)
	defer e.WaitEmptyBacklog()

	newOwner := e.NewSigScheme()
	newOwnerAgentID := coretypes.NewAgentIDFromAddress(newOwner.Address())
	req := NewCall(root.Interface.Name, root.FuncAllowChangeChainOwner, root.ParamChainOwner, newOwnerAgentID)
	_, err := e.PostRequest(req, newOwner)
	require.Error(t, err)

	_, ownerBack, _ := e.GetInfo()
	require.EqualValues(t, e.OriginatorAgentID, ownerBack)
}
