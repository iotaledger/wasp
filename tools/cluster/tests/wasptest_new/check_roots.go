package wasptest

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/builtinvm"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
	"testing"
)

func checkRoots(t *testing.T, chain *cluster.Chain) {
	chain.WithSCState(root.Hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		require.EqualValues(t, []byte{0xFF}, state.Get(root.VarStateInitialized))

		chid, _ := state.GetChainID(root.VarChainID)
		require.EqualValues(t, &chain.ChainID, chid)

		aid, _ := state.GetAgentID(root.VarChainOwnerID)
		require.EqualValues(t, *chain.OriginatorID(), *aid)

		desc, _ := state.GetString(root.VarDescription)
		require.EqualValues(t, chain.Description, desc)

		contractRegistry := state.GetMap(root.VarContractRegistry)

		crBytes := contractRegistry.GetAt(root.Hname.Bytes())
		require.NotNil(t, crBytes)
		require.True(t, bytes.Equal(crBytes, util.MustBytes(root.GetRootContractRecord())))

		crBytes = contractRegistry.GetAt(accountsc.Hname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, builtinvm.VMType, cr.VMType)
		require.EqualValues(t, accountsc.ContractDescription, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		require.EqualValues(t, accountsc.ContractName, cr.Name)

		return true
	})
}
