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
	e.CheckBase()
	e.Infof("\n%s\n", e.String())
}

func TestRepeatInit(t *testing.T) {
	e := New(t, false, false)
	req := NewCall(root.Interface.Name, "init")
	_, err := e.PostRequest(req, nil)
	require.Error(t, err)
}

func TestGetInfo(t *testing.T) {
	e := New(t, false, false)
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
	require.EqualValues(t, e.OriginatorAgentID, rec.Originator)
	require.EqualValues(t, inccounter.ProgramHash, rec.ProgramHash)

	recFind, err := e.FindContract(name)
	require.NoError(t, err)
	require.EqualValues(t, root.EncodeContractRecord(recFind), root.EncodeContractRecord(rec))
}

func TestDeployDouble(t *testing.T) {
	name := "testInc"
	e := New(t, false, false)
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
	require.EqualValues(t, e.OriginatorAgentID, rec.Originator)
	require.EqualValues(t, inccounter.ProgramHash, rec.ProgramHash)
}
