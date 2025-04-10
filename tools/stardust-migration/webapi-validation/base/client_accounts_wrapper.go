package base

import (
	stardust_client "github.com/nnikolash/wasp-types-exported/clients/apiclient"
	"github.com/stretchr/testify/require"

	rebased_client "github.com/iotaledger/wasp/clients/apiclient"
)

type AccountsClientWrapper struct {
	ValidationContext
}

// AccountsGetAccountBalance wraps both API calls for getting account balance
func (c *AccountsClientWrapper) AccountsGetAccountBalance(stateIndex uint32, agentID string) (*stardust_client.AssetsResponse, *rebased_client.AssetsResponse) {
	sRes, _, err := c.SClient.CorecontractsApi.AccountsGetAccountBalance(c.Ctx, MainnetChainID, agentID).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err)

	rRes, _, err := c.RClient.CorecontractsAPI.AccountsGetAccountBalance(c.Ctx, agentID).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err)

	return sRes, rRes
}

// AccountsGetAccountNFTIDs wraps both API calls for getting account NFT IDs
func (c *AccountsClientWrapper) AccountsGetAccountNFTIDs(stateIndex uint32, agentID string) (*stardust_client.AccountNFTsResponse, *rebased_client.AccountNFTsResponse) {
	sRes, _, err := c.SClient.CorecontractsApi.AccountsGetAccountNFTIDs(c.Ctx, MainnetChainID, agentID).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err)

	rRes, _, err := c.RClient.CorecontractsAPI.AccountsGetAccountNFTIDs(c.Ctx, agentID).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err)

	return sRes, rRes
}

// AccountsGetAccountNonce wraps both API calls for getting account nonce
func (c *AccountsClientWrapper) AccountsGetAccountNonce(stateIndex uint32, agentID string) (*stardust_client.AccountNonceResponse, *rebased_client.AccountNonceResponse) {
	sRes, _, err := c.SClient.CorecontractsApi.AccountsGetAccountNonce(c.Ctx, MainnetChainID, agentID).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err)

	rRes, _, err := c.RClient.CorecontractsAPI.AccountsGetAccountNonce(c.Ctx, agentID).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err)

	return sRes, rRes
}

// AccountsGetTotalAssets wraps both API calls for getting total assets
func (c *AccountsClientWrapper) AccountsGetTotalAssets(stateIndex uint32) (*stardust_client.AssetsResponse, *rebased_client.AssetsResponse) {
	sRes, _, err := c.SClient.CorecontractsApi.AccountsGetTotalAssets(c.Ctx, MainnetChainID).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err)

	rRes, _, err := c.RClient.CorecontractsAPI.AccountsGetTotalAssets(c.Ctx).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err)

	return sRes, rRes
}
