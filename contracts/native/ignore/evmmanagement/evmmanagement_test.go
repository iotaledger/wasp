package evmchainmanagement

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/native/evmchain"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

func TestRequestGasFees(t *testing.T) {
	chain, _ := evmchain.InitEVMChain(t)

	// ret, err := chain.CallView(Interface.Name, FuncGetBalance, FieldAddress, faucetAddress.Bytes())
	// require.NoError(t, err)

	// bal := big.NewInt(0)
	// bal.SetBytes(ret.MustGet(FieldBalance))
	// require.Zero(t, faucetSupply.Cmp(bal))

	// change owner to evnchainmanagement SC
	_, err = chain.PostRequestSync(
		solo.NewCallParams(evmchain.Interface.Name, evmchain.FuncSetOwner, FieldEvmOwner, manager.Name).
			WithIotas(1),
		chain.OriginatorKeyPair,
	)
	require.NoError(t, err)

	// call requestGasFees manually, to make the manager SC request funds from the evm chain

}
