package testcore

import (
	"testing"

	"github.com/iotaledger/wasp/packages/kv/codec"

	"github.com/iotaledger/wasp/packages/kv/collections"

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

		res, err := chain.CallView(governance.Name, governance.FuncGetAllowedCommitteeAddresses)
		require.NoError(t, err)
		require.EqualValues(t, 0, len(res))
	})
	t.Run("add/remove allowed rotation addresses", func(t *testing.T) {
		env := solo.New(t, false, false)
		chain := env.NewChain(nil, "chain1")
		defer chain.Log.Sync()

		_, addr1 := env.NewKeyPair()
		req := solo.NewCallParams(governance.Name, governance.FuncAddAllowedCommitteeAddress,
			governance.ParamStateAddress, addr1).WithIotas(1)
		_, err := chain.PostRequestSync(req, nil)
		require.NoError(t, err)

		res, err := chain.CallView(governance.Name, governance.FuncGetAllowedCommitteeAddresses)
		require.NoError(t, err)
		retArr := collections.NewArray16ReadOnly(res, governance.ParamAllowedAddresses)
		require.EqualValues(t, 1, retArr.MustLen())

		_, addr2 := env.NewKeyPair()
		req = solo.NewCallParams(governance.Name, governance.FuncAddAllowedCommitteeAddress,
			governance.ParamStateAddress, addr2).WithIotas(1)
		_, err = chain.PostRequestSync(req, nil)
		require.NoError(t, err)

		res, err = chain.CallView(governance.Name, governance.FuncGetAllowedCommitteeAddresses)
		require.NoError(t, err)
		retArr = collections.NewArray16ReadOnly(res, governance.ParamAllowedAddresses)
		require.EqualValues(t, 2, retArr.MustLen())

		a1, ok, err := codec.DecodeAddress(retArr.MustGetAt(0))
		require.NoError(t, err)
		require.True(t, ok)
		a2, ok, err := codec.DecodeAddress(retArr.MustGetAt(1))
		require.NoError(t, err)
		require.True(t, ok)

		require.True(t, addr1.Equals(a1) || addr1.Equals(a2))
		require.True(t, addr2.Equals(a1) || addr2.Equals(a2))

		req = solo.NewCallParams(governance.Name, governance.FuncRemoveAllowedCommitteeAddress,
			governance.ParamStateAddress, addr1).WithIotas(1)
		_, err = chain.PostRequestSync(req, nil)
		require.NoError(t, err)

		res, err = chain.CallView(governance.Name, governance.FuncGetAllowedCommitteeAddresses)
		require.NoError(t, err)
		retArr = collections.NewArray16ReadOnly(res, governance.ParamAllowedAddresses)
		require.EqualValues(t, 1, retArr.MustLen())

		require.EqualValues(t, addr2.Bytes(), retArr.MustGetAt(0))

		req = solo.NewCallParams(governance.Name, governance.FuncRemoveAllowedCommitteeAddress,
			governance.ParamStateAddress, addr1).WithIotas(1)
		_, err = chain.PostRequestSync(req, nil)
		require.NoError(t, err)

		res, err = chain.CallView(governance.Name, governance.FuncGetAllowedCommitteeAddresses)
		require.NoError(t, err)
		retArr = collections.NewArray16ReadOnly(res, governance.ParamAllowedAddresses)
		require.EqualValues(t, 1, retArr.MustLen())

		req = solo.NewCallParams(governance.Name, governance.FuncRemoveAllowedCommitteeAddress,
			governance.ParamStateAddress, addr2).WithIotas(1)
		_, err = chain.PostRequestSync(req, nil)
		require.NoError(t, err)

		res, err = chain.CallView(governance.Name, governance.FuncGetAllowedCommitteeAddresses)
		require.NoError(t, err)
		require.EqualValues(t, 0, len(res))
	})
}
