package evmchainmanagement

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/contracts/native/evmchain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

func TestRequestGasFees(t *testing.T) {
	chain, env := evmchain.InitEVMChain(t)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)

	// deploy solidity `storage` contract (just to produce some fees to be claimed)
	evmchain.DeployEVMContract(t, chain, env, evmchain.TestFaucetKey, contractABI, evmtest.StorageContractBytecode, uint32(42))

	// change owner to evnchainmanagement SC
	managerAgentId := coretypes.NewAgentID(chain.ChainID.AsAddress(), coretypes.Hn(Interface.Name))
	_, err = chain.PostRequestSync(
		solo.NewCallParams(evmchain.Interface.Name, evmchain.FuncSetNextOwner, evmchain.FieldNextEvmOwner, managerAgentId).
			WithIotas(1),
		chain.OriginatorKeyPair,
	)
	require.NoError(t, err)

	//claim ownership
	_, err = chain.PostRequestSync(
		solo.NewCallParams(Interface.Name, FuncClaimOwnership).WithIotas(1),
		chain.OriginatorKeyPair,
	)
	require.NoError(t, err)

	// call requestGasFees manually, so that the manager SC request funds from the evm chain, check funds are received by the manager SC
	balance0, _ := chain.GetAccountBalance(managerAgentId).Get(ledgerstate.ColorIOTA)

	_, err = chain.PostRequestSync(
		solo.NewCallParams(Interface.Name, FuncRequestGasFees).WithIotas(1),
		chain.OriginatorKeyPair,
	)
	require.NoError(t, err)
	balance1, _ := chain.GetAccountBalance(managerAgentId).Get(ledgerstate.ColorIOTA)

	require.Greater(t, balance1, balance0)

}
