package base

import (
	stardust_client "github.com/nnikolash/wasp-types-exported/clients/apiclient"
	"github.com/stretchr/testify/require"

	rebased_client "github.com/iotaledger/wasp/clients/apiclient"
)

type GovernanceClientWrapper struct {
	ValidationContext
}

// GovernanceGetChainInfo wraps both API calls for getting chain info
func (c *GovernanceClientWrapper) GovernanceGetChainInfo(stateIndex uint32) (*stardust_client.GovChainInfoResponse, *rebased_client.GovChainInfoResponse) {
	sRes, _, err := c.SClient.CorecontractsApi.GovernanceGetChainInfo(c.Ctx, MainnetChainID).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err)

	rRes, _, err := c.RClient.CorecontractsAPI.GovernanceGetChainInfo(c.Ctx).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err)

	return sRes, rRes
}
