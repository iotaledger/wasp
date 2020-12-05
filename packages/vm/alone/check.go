package alone

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/stretchr/testify/require"
)

func (e *aloneEnvironment) CheckBalance(addr address.Address, col balance.Color, expected int64) {
	require.EqualValues(e.T, expected, e.GetBalance(addr, col))
}

func (e *aloneEnvironment) CheckBase() {
	req := NewCall(root.Interface.Name, root.FuncGetInfo)
	res1, err := e.PostRequest(req, nil)
	require.NoError(e.T, err)

	res2, err := e.CallView(req)
	require.NoError(e.T, err)

	require.EqualValues(e.T, res1.Hash(), res2.Hash())

	rootRec, err := e.FindContract(root.Interface.Name)
	require.NoError(e.T, err)
	require.EqualValues(e.T, root.EncodeContractRecord(&root.RootContractRecord), root.EncodeContractRecord(rootRec))

	accountsRec, err := e.FindContract(accountsc.Interface.Name)
	require.NoError(e.T, err)
	require.EqualValues(e.T, accountsc.Interface.Name, accountsRec.Name)
	require.EqualValues(e.T, accountsc.Interface.Description, accountsRec.Description)
	require.EqualValues(e.T, accountsc.Interface.ProgramHash, accountsRec.ProgramHash)
	require.EqualValues(e.T, e.OriginatorAgentID, accountsRec.Originator)

	blobRec, err := e.FindContract(blob.Interface.Name)
	require.NoError(e.T, err)
	require.EqualValues(e.T, blob.Interface.Name, blobRec.Name)
	require.EqualValues(e.T, blob.Interface.Description, blobRec.Description)
	require.EqualValues(e.T, blob.Interface.ProgramHash, blobRec.ProgramHash)
	require.EqualValues(e.T, e.OriginatorAgentID, blobRec.Originator)
}
